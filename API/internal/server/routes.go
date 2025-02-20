package server

import (
	"API/internal/controllers"
	"API/internal/middleware"
	"API/internal/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func NewServer() *fiber.App {
	app := fiber.New(fiber.Config{
		Prefork:              false,
		Concurrency:          256,
		ReadTimeout:          10 * time.Second,
		WriteTimeout:         10 * time.Second,
		IdleTimeout:          20 * time.Second,
		BodyLimit:            2 * 1024 * 1024,
		DisableKeepalive:     true,
		CompressedFileSuffix: ".fiber.gz",
		EnablePrintRoutes:    true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Server error", err.Error())
		},
	})

	return app
}

func (s *FiberServer) RegisterFiberRoutes() {

	// Apply CORS middleware
	s.App.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Accept,Authorization,Content-Type",
		AllowCredentials: false,
		MaxAge:           300,
	}))

	authController := controllers.NewAuthController(s.db)

	// Public routes
	auth := s.App.Group("/api/v1/auth")
	auth.Post("/register", authController.Register)
	auth.Post("/login", authController.Login)
	auth.Post("/forgot-password", authController.ForgotPassword)
	auth.Get("/verify/:token", authController.VerifyEmail)
	auth.Get("/reset-password/:Token", authController.RestPassword)

	// Protected routes
	protected := s.App.Group("/api/v1", middleware.AuthRequired())
	protected.Delete("/user/:ID", authController.DeleteUser)
	protected.Put("/user/:ID", authController.EditUser)

	// Health check
	s.App.Get("/api/health", s.healthHandler)
	s.App.Get("/api/hello", s.healthHandler)
}

func (s *FiberServer) HelloWorldHandler(c *fiber.Ctx) error {
	resp := fiber.Map{
		"message": "Hello World",
	}

	return c.JSON(resp)
}

func (s *FiberServer) healthHandler(c *fiber.Ctx) error {
	return c.JSON(s.db.Health())
}
