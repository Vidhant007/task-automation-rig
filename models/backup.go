package models

import "time"

type CompressionType string

const (
    Tar    CompressionType = "tar"     // Basic tar archive
    TarGz  CompressionType = "tar.gz"  // Tar with gzip compression
    TarBz2 CompressionType = "tar.bz2" // Tar with bzip2 compression
    TarXz  CompressionType = "tar.xz"  // Tar with xz compression
    Zip    CompressionType = "zip"     // ZIP archive
    SevenZ CompressionType = "7z"      // 7-Zip archive
    Rar    CompressionType = "rar"     // RAR archive
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