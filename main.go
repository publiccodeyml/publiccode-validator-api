package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/italia/publiccode-validator-api/internal/common"
	"github.com/italia/publiccode-validator-api/internal/handlers"
	"github.com/italia/publiccode-validator-api/internal/jsondecoder"
)

var (
	Version = "dev" //nolint:gochecknoglobals // We need this to be set at build
	Commit  = "-"   //nolint:gochecknoglobals // We need this to be set at build
)

func main() {
	app := Setup()
	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}

func Setup() *fiber.App {
	methods := append(fiber.DefaultMethods, "QUERY") //nolint:gocritic // We want a new slice here

	app := fiber.New(fiber.Config{
		ErrorHandler: common.CustomErrorHandler,
		// Allow 30 seconds to upload the default maximum body of 4 MiB.
		ReadTimeout: 30 * time.Second,
		// This bounds response writes after the handler, including external checks, returns.
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
		// Fiber doesn't set DisallowUnknownFields by default
		// (https://github.com/gofiber/fiber/issues/2601)
		JSONDecoder:    jsondecoder.UnmarshalDisallowUnknownFields,
		RequestMethods: methods,
	})

	// Automatically recover panics in handlers
	app.Use(recover.New())

	app.Use(cors.New(cors.Config{
		AllowMethods: []string{"GET", "POST", "HEAD", "PUT", "DELETE", "PATCH", "QUERY"},
	}))

	setupHandlers(app)

	return app
}

func setupHandlers(app *fiber.App) {
	validateHandler := handlers.NewPubliccodeymlValidatorHandler()
	statusHandler := handlers.NewStatus(Version, Commit)

	v1 := app.Group("/v1")

	v1.Get("/status", statusHandler.GetStatus)
	v1.Add([]string{"QUERY"}, "/validate", validateHandler.Query)
	v1.Add([]string{"POST"}, "/validate", validateHandler.Query)
}
