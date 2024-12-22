package models

import "time"

type CompressionType string

const (
    Tar  CompressionType = "tar"
    Zip  CompressionType = "zip"
    Gzip CompressionType = "gzip"
)

type BackupRequest struct {
    Paths           []string        `json:"paths"`           // List of source paths to backup
    DestinationPath string         `json:"destinationPath"` // Destination path for the backup
    CompressionType CompressionType `json:"compressionType"` // Type of compression to use
}

type Backup struct {
    ID              string         `json:"id"`
    Paths           []string       `json:"paths"`
    DestinationPath string         `json:"destinationPath"`
    CompressionType CompressionType `json:"compressionType"`
    Status          string         `json:"status"`
    StartTime       time.Time      `json:"startTime"`
    EndTime         time.Time      `json:"endTime,omitempty"`
    Error           string         `json:"error,omitempty"`
}