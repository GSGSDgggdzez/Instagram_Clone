package server

import (
	"github.com/gofiber/fiber/v2"

	"API/internal/database"
)

type FiberServer struct {
	*fiber.App

	db database.Service
}

func New() *FiberServer {
	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "API",
			AppName:      "API",
		}),

		db: database.New(),
	}

	return server
}
