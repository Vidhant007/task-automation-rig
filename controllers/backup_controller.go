package controllers

import (
    "fmt"
    "os/exec"
    "path/filepath"
    "time"
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "task-automation-rig/models"
)

type BackupController struct {
    backups map[string]*models.Backup
}

func NewBackupController() *BackupController {
    return &BackupController{
        backups: make(map[string]*models.Backup),
    }
}

// CreateBackup initiates a new backup job
func (c *BackupController) CreateBackup(ctx *fiber.Ctx) error {
    var request models.BackupRequest
    if err := ctx.BodyParser(&request); err != nil {
        return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request body",
        })
    }

    // Validate request
    if len(request.Paths) == 0 || request.DestinationPath == "" {
        return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Paths and destination path are required",
        })
    }

    // Generate timestamp for the backup filename
    timestamp := time.Now().Format("2006-01-02_15-04-05")
    
    // Get the directory and original filename
    destDir := filepath.Dir(request.DestinationPath)
    
    // Create the new filename with timestamp
    var newFilename string
    switch request.CompressionType {
    case models.Tar:
        newFilename = fmt.Sprintf("backup_%s.tar.gz", timestamp)
    case models.Zip:
        newFilename = fmt.Sprintf("backup_%s.zip", timestamp)
    default:
        newFilename = fmt.Sprintf("backup_%s.tar.gz", timestamp)
    }

    // Combine the directory with the new filename
    request.DestinationPath = filepath.Join(destDir, newFilename)

    // Create backup record
    backup := &models.Backup{
        ID:              uuid.New().String(),
        Paths:           request.Paths,
        DestinationPath: request.DestinationPath,
        CompressionType: request.CompressionType,
        Status:          "pending",
        StartTime:       time.Now(),
    }

    // Store backup record
    c.backups[backup.ID] = backup

    // Start backup process asynchronously
    go c.processBackup(backup)

    return ctx.Status(fiber.StatusAccepted).JSON(backup)
}

// GetBackup returns the status of a specific backup job
func (c *BackupController) GetBackup(ctx *fiber.Ctx) error {
    id := ctx.Params("id")
    backup, exists := c.backups[id]
    if !exists {
        return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "Backup not found",
        })
    }

    return ctx.JSON(backup)
}

// ListBackups returns all backup jobs
func (c *BackupController) ListBackups(ctx *fiber.Ctx) error {
    backupList := make([]*models.Backup, 0, len(c.backups))
    for _, backup := range c.backups {
        backupList = append(backupList, backup)
    }
    return ctx.JSON(backupList)
}

func (c *BackupController) processBackup(backup *models.Backup) {
    backup.Status = "in_progress"

    // Create destination directory if it doesn't exist
    destDir := filepath.Dir(backup.DestinationPath)
    if err := exec.Command("mkdir", "-p", destDir).Run(); err != nil {
        backup.Status = "failed"
        backup.Error = err.Error()
        backup.EndTime = time.Now()
        return
    }

    // Prepare compression command based on type
    var cmd *exec.Cmd
    switch backup.CompressionType {
    case models.Tar:
        args := append([]string{"-czf", backup.DestinationPath}, backup.Paths...)
        cmd = exec.Command("tar", args...)
    case models.Zip:
        args := append([]string{"-r", backup.DestinationPath}, backup.Paths...)
        cmd = exec.Command("zip", args...)
    default:
        backup.Status = "failed"
        backup.Error = "unsupported compression type"
        backup.EndTime = time.Now()
        return
    }

    // Execute compression command
    if err := cmd.Run(); err != nil {
        backup.Status = "failed"
        backup.Error = err.Error()
    } else {
        backup.Status = "completed"
    }

    backup.EndTime = time.Now()
}