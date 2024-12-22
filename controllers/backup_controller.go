package controllers

import (
    "fmt"
    "os/exec"
    "path/filepath"
    "time"
    "log"
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
        newFilename = fmt.Sprintf("backup_%s.tar", timestamp)
    case models.TarGz:
        newFilename = fmt.Sprintf("backup_%s.tar.gz", timestamp)
    case models.TarBz2:
        newFilename = fmt.Sprintf("backup_%s.tar.bz2", timestamp)
    case models.TarXz:
        newFilename = fmt.Sprintf("backup_%s.tar.xz", timestamp)
    case models.Zip:
        newFilename = fmt.Sprintf("backup_%s.zip", timestamp)
    case models.SevenZ:
        newFilename = fmt.Sprintf("backup_%s.7z", timestamp)
    case models.Rar:
        newFilename = fmt.Sprintf("backup_%s.rar", timestamp)
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
    log.Printf("Starting backup process for ID: %s\n", backup.ID)
    log.Printf("Source paths: %v\n", backup.Paths)
    log.Printf("Destination: %s\n", backup.DestinationPath)

    // Create destination directory if it doesn't exist
    destDir := filepath.Dir(backup.DestinationPath)
    if err := exec.Command("mkdir", "-p", destDir).Run(); err != nil {
        log.Printf("Failed to create directory: %s\n", err)
        backup.Status = "failed"
        backup.Error = err.Error()
        backup.EndTime = time.Now()
        return
    }

    // Prepare compression command based on type
    var cmd *exec.Cmd
    var cmdStr string
    
    switch backup.CompressionType {
    case models.Tar:
        args := append([]string{"-cvf", backup.DestinationPath}, backup.Paths...)
        cmd = exec.Command("tar", args...)
        cmdStr = fmt.Sprintf("tar -cvf %s %v", backup.DestinationPath, backup.Paths)
    
    case models.TarGz:
        args := append([]string{"-czf", backup.DestinationPath}, backup.Paths...)
        cmd = exec.Command("tar", args...)
        cmdStr = fmt.Sprintf("tar -czf %s %v", backup.DestinationPath, backup.Paths)
    
    case models.TarBz2:
        args := append([]string{"-cjf", backup.DestinationPath}, backup.Paths...)
        cmd = exec.Command("tar", args...)
        cmdStr = fmt.Sprintf("tar -cjf %s %v", backup.DestinationPath, backup.Paths)
    
    case models.TarXz:
        args := append([]string{"-cJf", backup.DestinationPath}, backup.Paths...)
        cmd = exec.Command("tar", args...)
        cmdStr = fmt.Sprintf("tar -cJf %s %v", backup.DestinationPath, backup.Paths)
    
    case models.Zip:
        args := append([]string{"-r", backup.DestinationPath}, backup.Paths...)
        cmd = exec.Command("zip", args...)
        cmdStr = fmt.Sprintf("zip -r %s %v", backup.DestinationPath, backup.Paths)
    
    case models.SevenZ:
        args := append([]string{"a", backup.DestinationPath}, backup.Paths...)
        cmd = exec.Command("7z", args...)
        cmdStr = fmt.Sprintf("7z a %s %v", backup.DestinationPath, backup.Paths)
    
    case models.Rar:
        args := append([]string{"a", backup.DestinationPath}, backup.Paths...)
        cmd = exec.Command("rar", args...)
        cmdStr = fmt.Sprintf("rar a %s %v", backup.DestinationPath, backup.Paths)
    
    default:
        log.Printf("Unsupported compression type: %s\n", backup.CompressionType)
        backup.Status = "failed"
        backup.Error = "unsupported compression type"
        backup.EndTime = time.Now()
        return
    }

    log.Printf("Executing command: %s\n", cmdStr)

    // Capture command output
    output, err := cmd.CombinedOutput()
    if err != nil {
        log.Printf("Command failed: %s\n", err)
        log.Printf("Command output: %s\n", string(output))
        backup.Status = "failed"
        backup.Error = fmt.Sprintf("Command failed: %s. Output: %s", err, string(output))
    } else {
        log.Printf("Command completed successfully\n")
        log.Printf("Command output: %s\n", string(output))
        backup.Status = "completed"
    }

    backup.EndTime = time.Now()
}