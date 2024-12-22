package models

import "time"

type CodecType string
type Resolution string
type ContainerFormat string

const (
    // Video Codecs
    H264 CodecType = "h264"
    H265 CodecType = "h265"
    VP9  CodecType = "vp9"
    AV1  CodecType = "av1"
)

const (
    // Containers
    MP4  ContainerFormat = "mp4"
    MKV  ContainerFormat = "mkv"
    WebM ContainerFormat = "webm"
    AVI  ContainerFormat = "avi"
)

const (
    // Standard Resolutions
    Res360p  Resolution = "360p"   // 640x360
    Res480p  Resolution = "480p"   // 854x480
    Res540p  Resolution = "540p"   // 960x540
    Res720p  Resolution = "720p"   // 1280x720
    Res1080p Resolution = "1080p"  // 1920x1080
)

type MediaRequest struct {
    SourcePath      string          `json:"sourcePath"`      // Source video path
    CodecType       CodecType       `json:"codecType"`       // Target codec
    ContainerFormat ContainerFormat `json:"containerFormat"` // Output container format
    Resolutions     []Resolution    `json:"resolutions"`     // Target resolutions
    DestinationPath string          `json:"destinationPath"` // Output directory
}

type MediaJob struct {
    ID              string          `json:"id"`
    SourcePath      string          `json:"sourcePath"`
    CodecType       CodecType       `json:"codecType"`
    ContainerFormat ContainerFormat `json:"containerFormat"`
    Resolutions     []Resolution    `json:"resolutions"`
    DestinationPath string          `json:"destinationPath"`
    Status          string          `json:"status"`
    CurrentFile     string          `json:"currentFile,omitempty"`
    ProcessedFiles  []string        `json:"processedFiles"`
    StartTime       time.Time       `json:"startTime"`
    EndTime         time.Time       `json:"endTime,omitempty"`
    Error           string          `json:"error,omitempty"`
}

func (r Resolution) GetDimensions() (width, height int) {
    switch r {
    case Res360p:
        return 640, 360
    case Res480p:
        return 854, 480
    case Res540p:
        return 960, 540
    case Res720p:
        return 1280, 720
    case Res1080p:
        return 1920, 1080
    default:
        return 1280, 720 // Default to 720p
    }
}