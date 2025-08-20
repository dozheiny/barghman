package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	ptime "github.com/yaa110/go-persian-calendar"
)

var ErrContentLengthMismatch = errors.New("content length mismatch")

type FileContent struct {
	UID                 string    `json:"uid" toml:"uid"`
	BillID              string    `json:"bill_id" toml:"bill_id"`
	Sequence            uint      `json:"sequence" toml:"sequence"`
	OutageNumber        int       `json:"outage_number" toml:"outage_number"`
	StartOutageDateTime time.Time `json:"start_outage_datetime" toml:"outage_datetime"`
	EndOutageDateTime   time.Time `json:"end_outage_datetime" toml:"end_outage_datetime"`
	Recipients          []string  `json:"recipients" toml:"recipients"`
	Address             string    `json:"address" toml:"address"`
	ReasonOutage        string    `json:"reason_outage" toml:"reason_outage"`
}

// Pattern: "{bill_id}_{outage-number}_{outage-date}.json"
func (f *FileContent) FileName() string {
	return FileName(f.BillID, f.OutageNumber, f.StartOutageDateTime)
}

func FileName(billID string, outageNumber int, date time.Time) string {
	return fmt.Sprintf("%s_%d_%s.json", billID, outageNumber, date.Format(time.DateOnly))
}

func (f *FileContent) Write(file *os.File) error {
	content, err := json.Marshal(f)
	if err != nil {
		slog.Error("Encode data failed", "error", err)
		return err
	}

	if _, err := file.WriteAt(content, 0); err != nil {
		slog.Error("Failed to write content into file", "error", err)
		return err
	}

	return nil
}

func LoadOrCreateFile(billID string, outageNumber int, date time.Time) (*os.File, error) {
	filePath := cachePathDir + FileName(billID, outageNumber, date)

	slog.Debug("file path to open or create", "file path", filePath)
	return os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0o644)
}

func (f *FileContent) Summary() string {
	return fmt.Sprintf("Power Outage on %s", f.Address)
}

func (f *FileContent) Description() string {
	return strings.ReplaceAll(fmt.Sprintf("Blackout!\nAddress: %s\nDate: %s\nFrom %s until %s\nReason: %s",
		f.Address, ptime.New(f.StartOutageDateTime).Format("YYYY/MM/DD"), f.StartOutageDateTime.Format(time.TimeOnly), f.EndOutageDateTime.Format(time.TimeOnly), f.ReasonOutage), "\n", "\\n")
}
