package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/go-ntlmssp"
	"golang.org/x/net/proxy"
)

const ewsRequestTimeout = 30 * time.Second

// ewsCreateItemEnvelope hands a fully-formed MIME message to Exchange for
// delivery via its MimeContent element, so the exact same message built for
// SMTP (headers, calendar attachment, etc.) can be reused verbatim here.
const ewsCreateItemEnvelope = `<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"
               xmlns:t="http://schemas.microsoft.com/exchange/services/2006/types"
               xmlns:m="http://schemas.microsoft.com/exchange/services/2006/messages">
  <soap:Header>
    <t:RequestServerVersion Version="Exchange2016"/>
  </soap:Header>
  <soap:Body>
    <m:CreateItem MessageDisposition="SendOnly">
      <m:Items>
        <t:Message>
          <t:MimeContent CharacterSet="UTF-8">%s</t:MimeContent>
        </t:Message>
      </m:Items>
    </m:CreateItem>
  </soap:Body>
</soap:Envelope>`

type ewsEnvelope struct {
	Body struct {
		Fault struct {
			FaultString string `xml:"faultstring"`
		} `xml:"Fault"`
		CreateItemResponse struct {
			ResponseMessages struct {
				CreateItemResponseMessage struct {
					ResponseClass string `xml:"ResponseClass,attr"`
					ResponseCode  string `xml:"ResponseCode"`
					MessageText   string `xml:"MessageText"`
				} `xml:"CreateItemResponseMessage"`
			} `xml:"ResponseMessages"`
		} `xml:"CreateItemResponse"`
	} `xml:"Body"`
}

// sendEWS delivers msg (a complete RFC 5322 MIME message, the same one used
// for SMTP) through Exchange Web Services over HTTPS. It's a drop-in
// alternative transport for environments where the SMTP port is blocked or
// throttled but EWS/OWA on 443 is reachable.
func (m Mail) sendEWS(msg string) error {
	endpoint := m.Config.EWSURL
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://%s/EWS/Exchange.asmx", m.Config.Address)
	}

	envelope := fmt.Sprintf(ewsCreateItemEnvelope, base64.StdEncoding.EncodeToString([]byte(msg)))

	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(envelope))
	if err != nil {
		return fmt.Errorf("failed to build EWS request: %w", err)
	}
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.SetBasicAuth(m.Config.Username, m.Config.Password)

	client, err := m.ewsHTTPClient()
	if err != nil {
		return fmt.Errorf("failed to build EWS http client: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("EWS request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read EWS response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("EWS request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var parsed ewsEnvelope
	if err := xml.Unmarshal(respBody, &parsed); err != nil {
		return fmt.Errorf("failed to parse EWS response: %w", err)
	}

	if parsed.Body.Fault.FaultString != "" {
		return fmt.Errorf("EWS SOAP fault: %s", parsed.Body.Fault.FaultString)
	}

	respMsg := parsed.Body.CreateItemResponse.ResponseMessages.CreateItemResponseMessage
	if respMsg.ResponseClass != "Success" {
		return fmt.Errorf("EWS CreateItem failed: %s (%s)", respMsg.MessageText, respMsg.ResponseCode)
	}

	return nil
}

// ewsHTTPClient builds an http.Client configured for NTLM authentication,
// optionally routed through the SMTP config's SOCKS5 proxy.
func (m Mail) ewsHTTPClient() (*http.Client, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: m.Config.SkipTLS},
	}

	if m.Config.Proxy.Enabled() {
		p := m.Config.Proxy
		proxyAddr := net.JoinHostPort(p.Host, p.Port)

		var auth proxy.Auth
		if p.Username != "" {
			auth = proxy.Auth{User: p.Username, Password: p.Password}
		}

		dialer, err := proxy.SOCKS5("tcp", proxyAddr, &auth, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
		}

		transport.DialContext = func(_ context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		}
	}

	return &http.Client{
		Timeout:   ewsRequestTimeout,
		Transport: ntlmssp.Negotiator{RoundTripper: transport},
	}, nil
}
