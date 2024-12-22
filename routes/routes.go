package routes

import (
    "github.com/gofiber/fiber/v2"
    "task-automation-rig/controllers"
)

func SetupRoutes(app *fiber.App) {
    // Initialize controllers
    backupController := controllers.NewBackupController()

    // Backup routes
    backup := app.Group("/api/backups")
    backup.Post("/", backupController.CreateBackup)
    backup.Get("/", backupController.ListBackups)
    backup.Get("/:id", backupController.GetBackup)
}