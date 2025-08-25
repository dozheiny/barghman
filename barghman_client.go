package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	ptime "github.com/yaa110/go-persian-calendar"
)

var (
	ErrUnexpectedStatusCode    = errors.New("unexpected status code")
	ErrInvalidOutageDateFormat = errors.New("invalid outage date format")
)

const PlannedBlackOutURL = "https://uiapi.saapa.ir/api/ebills/PlannedBlackoutsReport"

type PlannedBlackOutResponse struct {
	TimeStamp  time.Time `json:"TimeStamp"`
	Status     int       `json:"status"`
	SessionKey string    `json:"SessionKey"`
	Message    string    `json:"message"`
	Data       []Data    `json:"data"`
	Error      any       `json:"error"`
}

type PlannedBlackoutRequest struct {
	BillID   string `json:"bill_id"`
	FromDate string `json:"from_date"`
	ToDate   string `json:"to_date"`
}

type Data struct {
	RegDate         string `json:"reg_date"`
	Registrar       string `json:"registrar"`
	ReasonOutage    string `json:"reason_outage"`
	OutageDate      string `json:"outage_date"`
	OutageTime      string `json:"outage_time"`
	OutageStartTime string `json:"outage_start_time"`
	OutageStopTime  string `json:"outage_stop_time"`
	IsPlanned       bool   `json:"is_planned"`
	Address         string `json:"address"`
	OutageAddress   string `json:"outage_address"`
	City            int    `json:"city"`
	OutageNumber    int    `json:"outage_number"`
	TrackingCode    int    `json:"tracking_code"`
}

func PlannedBlackOut(ctx context.Context, authToken, billID string, startDate, endDate time.Time) ([]Data, error) {
	slog.Debug("going to call blackout", "from time", startDate.String(), "to time", endDate.String(), "bill id", billID)

	payload := PlannedBlackoutRequest{
		BillID:   billID,
		FromDate: ptime.New(startDate).Format("YYYY/MM/DD"),
		ToDate:   ptime.New(endDate).Format("YYYY/MM/DD"),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("failed to marshal request", "error", err)
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, PlannedBlackOutURL, bytes.NewBuffer(body))
	if err != nil {
		slog.Error("failed to create new request", "error", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")
	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Origin", "https://ios.bargheman.com")

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("failed to send request", "error", err)
		return nil, err
	}

	defer response.Body.Close()

	respbody, err := io.ReadAll(response.Body)
	if err != nil {
		slog.Error("failed to read response body", "error", err)
		return nil, err
	}

	slog.Debug("response of barghman", "body", string(respbody))

	if response.StatusCode != http.StatusOK {
		slog.Error("unexpected status code", "status_code", response.StatusCode)
		return nil, ErrUnexpectedStatusCode
	}

	var plannedBlackOutResponse PlannedBlackOutResponse
	if err := json.Unmarshal(respbody, &plannedBlackOutResponse); err != nil {
		slog.Error("failed to decode response", "error", err)
		return nil, err
	}

	if plannedBlackOutResponse.Status != http.StatusOK {
		slog.Error("unexpected status code", "status_code", plannedBlackOutResponse.Status)
		return nil, ErrUnexpectedStatusCode
	}

	return plannedBlackOutResponse.Data, nil
}

func (d Data) ToFileContent(loc *time.Location, billID string, recipients []string, sequence uint) (*FileContent, error) {
	startDate, endDate, err := d.ParseTime(loc)
	if err != nil {
		slog.Error("Failed to parse data time", "error", err)
		return nil, err
	}

	return &FileContent{
		UID:                 fmt.Sprintf("%s_%d_%s", billID, d.OutageNumber, startDate.Format(time.DateOnly)),
		BillID:              billID,
		Sequence:            sequence,
		OutageNumber:        d.OutageNumber,
		FarsiOutageDate:     d.OutageDate,
		StartOutageDateTime: startDate,
		EndOutageDateTime:   endDate,
		Recipients:          recipients,
		Address:             d.Address,
		ReasonOutage:        d.ReasonOutage,
	}, nil
}

func (d Data) ParseTime(loc *time.Location) (time.Time, time.Time, error) {
	outdateDate := strings.Split(d.OutageDate, "/")
	if len(outdateDate) != 3 {
		slog.Error("invalid outage date format", "outage_date", d.OutageDate)
		return time.Time{}, time.Time{}, ErrInvalidOutageDateFormat
	}

	outageStartTime := strings.Split(d.OutageStartTime, ":")
	if len(outageStartTime) != 2 {
		slog.Error("invalid outage start time format", "outage_start_time", d.OutageStartTime)
		return time.Time{}, time.Time{}, ErrInvalidOutageDateFormat
	}

	outageStopTime := strings.Split(d.OutageStopTime, ":")
	if len(outageStopTime) != 2 {
		slog.Error("invalid outage stop time format", "outage_stop_time", d.OutageStopTime)
		return time.Time{}, time.Time{}, ErrInvalidOutageDateFormat
	}

	year, err := strconv.Atoi(outdateDate[0])
	if err != nil {
		slog.Error("invalid outage date format", "outage_date", d.OutageDate)
		return time.Time{}, time.Time{}, ErrInvalidOutageDateFormat
	}

	month, err := strconv.Atoi(outdateDate[1])
	if err != nil {
		slog.Error("invalid outage date format", "outage_date", d.OutageDate)
		return time.Time{}, time.Time{}, ErrInvalidOutageDateFormat
	}

	day, err := strconv.Atoi(outdateDate[2])
	if err != nil {
		slog.Error("invalid outage date format", "outage_date", d.OutageDate)
		return time.Time{}, time.Time{}, ErrInvalidOutageDateFormat
	}

	startHour, err := strconv.Atoi(outageStartTime[0])
	if err != nil {
		slog.Error("invalid outage start time format", "outage_start_time", d.OutageStartTime)
		return time.Time{}, time.Time{}, ErrInvalidOutageDateFormat
	}

	startMinute, err := strconv.Atoi(outageStartTime[1])
	if err != nil {
		slog.Error("invalid outage start time format", "outage_start_time", d.OutageStartTime)
		return time.Time{}, time.Time{}, ErrInvalidOutageDateFormat
	}

	stopHour, err := strconv.Atoi(outageStopTime[0])
	if err != nil {
		slog.Error("invalid outage stop time format", "outage_stop_time", d.OutageStopTime)
		return time.Time{}, time.Time{}, ErrInvalidOutageDateFormat
	}

	stopMinute, err := strconv.Atoi(outageStopTime[1])
	if err != nil {
		slog.Error("invalid outage stop time format", "outage_stop_time", d.OutageStopTime)
		return time.Time{}, time.Time{}, ErrInvalidOutageDateFormat
	}

	startDate := ptime.Date(year, ptime.Month(month), day, startHour, startMinute, 0, 0, loc)
	stopDate := ptime.Date(year, ptime.Month(month), day, stopHour, stopMinute, 0, 0, loc)

	return startDate.Time(), stopDate.Time(), nil
}
