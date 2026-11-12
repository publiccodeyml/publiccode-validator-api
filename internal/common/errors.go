package common

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v3"
)

func InternalServerError(title string) ProblemJSONError {
	return Error(fiber.StatusInternalServerError, title, fiber.ErrInternalServerError.Message)
}

func Error(status int, title string, detail string) ProblemJSONError {
	return ProblemJSONError{Title: title, Detail: detail, Status: status}
}

func CustomErrorHandler(ctx fiber.Ctx, err error) error {
	var problemJSON *ProblemJSONError

	// Retrieve the custom status code if it's a fiber.Error.
	fiberError, ok := errors.AsType[*fiber.Error](err)
	if ok {
		problemJSON = &ProblemJSONError{
			Code:   "",
			Title:  fiberError.Message,
			Detail: "",
			Status: fiberError.Code,
		}
	}

	if problemJSON == nil {
		//nolint:errorlint
		switch e := err.(type) {
		case ProblemJSONError:
			problemJSON = &e
		default:
			log.Printf("Unhandled request error: %v", err)

			problemJSON = &ProblemJSONError{
				Status: fiber.StatusInternalServerError,
				Title:  fiber.ErrInternalServerError.Message,
			}
		}
	}

	ctx.Status(problemJSON.Status)

	return ctx.JSON(problemJSON, "application/problem+json")
}
