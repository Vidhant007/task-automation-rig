package controllers

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"
    "os/exec"
    "log"
    "bufio"
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "task-automation-rig/models"
)

type MediaController struct {
    jobs map[string]*models.MediaJob
}

func NewMediaController() *MediaController {
    return &MediaController{
        jobs: make(map[string]*models.MediaJob),
    }
}

func (c *MediaController) CreateMediaJob(ctx *fiber.Ctx) error {
    var request models.MediaRequest
    if err := ctx.BodyParser(&request); err != nil {
        return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }

    if request.SourcePath == "" || request.DestinationPath == "" {
        return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Source and destination paths are required",
        })
    }

    job := &models.MediaJob{
        ID:              uuid.New().String(),
        SourcePath:      request.SourcePath,
        CodecType:       request.CodecType,
        ContainerFormat: request.ContainerFormat,
        Resolutions:     request.Resolutions,
        DestinationPath: request.DestinationPath,
        Status:         "pending",
        ProcessedFiles:  make([]string, 0),
        StartTime:      time.Now(),
    }

    c.jobs[job.ID] = job
    go c.processMediaJob(job)

    return ctx.Status(fiber.StatusAccepted).JSON(job)
}

func (c *MediaController) GetMediaJob(ctx *fiber.Ctx) error {
    id := ctx.Params("id")
    job, exists := c.jobs[id]
    if !exists {
        return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "Job not found",
        })
    }
    return ctx.JSON(job)
}

func (c *MediaController) ListMediaJobs(ctx *fiber.Ctx) error {
    jobList := make([]*models.MediaJob, 0, len(c.jobs))
    for _, job := range c.jobs {
        jobList = append(jobList, job)
    }
    return ctx.JSON(jobList)
}

func (c *MediaController) processMediaJob(job *models.MediaJob) {
    job.Status = "in_progress"
    log.Printf("Starting media job: %s\n", job.ID)
    
    if err := os.MkdirAll(job.DestinationPath, 0755); err != nil {
        log.Printf("Failed to create destination directory: %v\n", err)
        job.Status = "failed"
        job.Error = fmt.Sprintf("Failed to create destination directory: %v", err)
        job.EndTime = time.Now()
        return
    }

    fileInfo, err := os.Stat(job.SourcePath)
    if err != nil {
        log.Printf("Failed to access source path: %v\n", err)
        job.Status = "failed"
        job.Error = fmt.Sprintf("Failed to access source path: %v", err)
        job.EndTime = time.Now()
        return
    }

    if fileInfo.IsDir() {
        log.Printf("Processing directory: %s\n", job.SourcePath)
        err = filepath.Walk(job.SourcePath, func(path string, info os.FileInfo, err error) error {
            if err != nil {
                return err
            }
            if !info.IsDir() && isVideoFile(path) {
                log.Printf("Found video file: %s\n", path)
                if err := c.processVideo(job, path); err != nil {
                    log.Printf("Error processing video %s: %v\n", path, err)
                    return err
                }
                job.ProcessedFiles = append(job.ProcessedFiles, path)
            }
            return nil
        })
    } else {
        log.Printf("Processing single file: %s\n", job.SourcePath)
        if isVideoFile(job.SourcePath) {
            err = c.processVideo(job, job.SourcePath)
            if err == nil {
                job.ProcessedFiles = append(job.ProcessedFiles, job.SourcePath)
            }
        } else {
            err = fmt.Errorf("not a video file")
            log.Printf("Error: %s is not a video file\n", job.SourcePath)
        }
    }

    if err != nil {
        log.Printf("Job failed: %v\n", err)
        job.Status = "failed"
        job.Error = err.Error()
    } else {
        log.Printf("Job completed successfully\n")
        job.Status = "completed"
    }
    
    job.EndTime = time.Now()
}

func (c *MediaController) processVideo(job *models.MediaJob, videoPath string) error {
    job.CurrentFile = videoPath
    log.Printf("Starting processing for video: %s\n", videoPath)
    log.Printf("Target codec: %s, Container: %s\n", job.CodecType, job.ContainerFormat)
    
    baseFile := filepath.Base(videoPath)
    fileName := strings.TrimSuffix(baseFile, filepath.Ext(baseFile))
    ext := getContainerExtension(job.ContainerFormat)

    for _, resolution := range job.Resolutions {
        width, height := resolution.GetDimensions()
        outputPath := filepath.Join(job.DestinationPath, 
            fmt.Sprintf("%s_%s_%s%s", 
                fileName, 
                string(job.CodecType), 
                string(resolution),
                ext))

        log.Printf("Processing resolution %s (%dx%d)\n", resolution, width, height)
        log.Printf("Output path: %s\n", outputPath)

        args := []string{
            "-i", videoPath,
            "-c:v", getFFmpegCodec(job.CodecType)}

        switch job.CodecType {
        case models.VP9:
            args = append(args, "-b:v", "0", "-crf", "31", "-deadline", "good", "-cpu-used", "4")
        case models.H265:
            args = append(args, "-crf", "28", "-preset", "medium", "-x265-params", "log-level=error")
        case models.AV1:
            args = append(args, "-crf", "30", "-strict", "experimental", "-cpu-used", "4")
        default: // H264
            args = append(args, "-crf", "23", "-preset", "medium", "-movflags", "+faststart")
        }

        args = append(args, "-vf", fmt.Sprintf("scale=%d:%d", width, height), "-progress", "pipe:1")

        switch job.ContainerFormat {
        case models.WebM:
            args = append(args, "-f", "webm")
        case models.MKV:
            args = append(args, "-f", "matroska")
        case models.AVI:
            args = append(args, "-f", "avi")
        default: // MP4
            args = append(args, "-f", "mp4")
        }

        args = append(args, "-c:a", "copy", outputPath)

        cmd := exec.Command("ffmpeg", args...)
        log.Printf("Executing command: ffmpeg %v\n", args)

        stdout, err := cmd.StdoutPipe()
        if err != nil {
            log.Printf("Error creating stdout pipe: %v\n", err)
            return err
        }

        stderr, err := cmd.StderrPipe()
        if err != nil {
            log.Printf("Error creating stderr pipe: %v\n", err)
            return err
        }

        if err := cmd.Start(); err != nil {
            log.Printf("Error starting FFmpeg: %v\n", err)
            return err
        }

        go func() {
            scanner := bufio.NewScanner(stdout)
            for scanner.Scan() {
                log.Printf("FFmpeg progress: %s\n", scanner.Text())
            }
        }()

        go func() {
            scanner := bufio.NewScanner(stderr)
            for scanner.Scan() {
                log.Printf("FFmpeg output: %s\n", scanner.Text())
            }
        }()

        if err := cmd.Wait(); err != nil {
            log.Printf("FFmpeg command failed: %v\n", err)
            return err
        }

        log.Printf("Completed processing resolution %s\n", resolution)
    }

    log.Printf("Completed processing video: %s\n", videoPath)
    return nil
}

func isVideoFile(path string) bool {
    ext := strings.ToLower(filepath.Ext(path))
    videoExts := []string{".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm"}
    for _, videoExt := range videoExts {
        if ext == videoExt {
            return true
        }
    }
    return false
}

func getFFmpegCodec(codecType models.CodecType) string {
    switch codecType {
    case models.H264:
        return "libx264"
    case models.H265:
        return "libx265"
    case models.VP9:
        return "libvpx-vp9"
    case models.AV1:
        return "libaom-av1"
    default:
        return "libx264"
    }
}

func getContainerExtension(format models.ContainerFormat) string {
    switch format {
    case models.MKV:
        return ".mkv"
    case models.WebM:
        return ".webm"
    case models.AVI:
        return ".avi"
    default:
        return ".mp4"
    }
}