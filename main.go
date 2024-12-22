package main

import (
    "log"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "task-automation-rig/routes"
)

func main() {
    app := fiber.New(fiber.Config{
        ErrorHandler: func(c *fiber.Ctx, err error) error {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "error": err.Error(),
            })
        },
    })

    // Middleware
    app.Use(logger.New())

    // Setup routes
    routes.SetupRoutes(app)

    // Start server
    log.Fatal(app.Listen(":3000"))
}