package controllers

import (
    "github.com/gofiber/fiber/v2"
    "task-automation-rig/models"
)

type UserController struct {
    users map[string]*models.User
}

func NewUserController() *UserController {
    return &UserController{
        users: make(map[string]*models.User),
    }
}

func (c *UserController) GetUser(ctx *fiber.Ctx) error {
    id := ctx.Params("id")
    user, exists := c.users[id]
    if !exists {
        return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
            "error": "User not found",
        })
    }
    return ctx.JSON(user)
}