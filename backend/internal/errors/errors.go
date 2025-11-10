package errors

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

// Helpers

func BadRequest(msg string) *AppError {
	return &AppError{Code: fiber.StatusBadRequest, Message: msg}
}

func Unprocessable(msg string) *AppError {
	return &AppError{Code: fiber.StatusUnprocessableEntity, Message: msg}
}

func NotFound(msg string) *AppError {
	return &AppError{Code: fiber.StatusNotFound, Message: msg}
}

func Conflict(msg string) *AppError {
	return &AppError{Code: fiber.StatusConflict, Message: msg}
}

func Unauthorized(msg string) *AppError {
	return &AppError{Code: fiber.StatusUnauthorized, Message: msg}
}

func Forbidden(msg string) *AppError {
	return &AppError{Code: fiber.StatusForbidden, Message: msg}
}

func Internal(msg string) *AppError {
	return &AppError{Code: fiber.StatusInternalServerError, Message: msg}
}

func HandleError(c *fiber.Ctx, err error) error {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return c.Status(appErr.Code).JSON(fiber.Map{"error": appErr.Message})
	}

	// fallback: erreur inattendue
	log.Errorf("Unexpected error: %v", err)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": "internal server error",
	})
}

func HandleBodyParserError(c *fiber.Ctx, err error) error {
	return HandleError(c, BadRequest(fmt.Sprintf("invalid request body: %v", err.Error())))

}
