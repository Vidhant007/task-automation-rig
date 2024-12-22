package routes

import (
    "github.com/gofiber/fiber/v2"
    "task-automation-rig/controllers"
)

func SetupRoutes(app *fiber.App) {
    // Initialize controllers
    backupController := controllers.NewBackupController()
    mediaController := controllers.NewMediaController()

    // Backup routes
    backup := app.Group("/api/backups")
    backup.Post("/", backupController.CreateBackup)
    backup.Get("/", backupController.ListBackups)
    backup.Get("/:id", backupController.GetBackup)

    // Media processing routes
    media := app.Group("/api/media")
    media.Post("/", mediaController.CreateMediaJob)
    media.Get("/", mediaController.ListMediaJobs)
    media.Get("/:id", mediaController.GetMediaJob)
}