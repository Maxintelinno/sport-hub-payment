package main

import (
	"log"
	"sport-hub-payment/internal/database"
	"sport-hub-payment/internal/handler"
	"sport-hub-payment/internal/pkg/kbank"
	"sport-hub-payment/internal/repository"
	"sport-hub-payment/internal/service"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
)

func main() {
	// Initialize Viper Config
	viper.SetConfigFile("configs/config.yml")
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read config file: %v. Using defaults or env vars.", err)
	}

	// Initialize Database with GORM
	db, err := database.InitDB()
	if err != nil {
		log.Printf("Warning: Database connection failed: %v. Running without DB support.", err)
	}

	// Initialize KBank Client with configuration from viper
	kbankClient := kbank.NewClient()

	// Initialize Repository, Service and Handler
	paymentRepo := repository.NewPaymentRepository(db)
	paymentService := service.NewPaymentService(kbankClient, paymentRepo)
	paymentHandler := handler.NewPaymentHandler(paymentService)

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())

	// Setup Routes
	handler.SetupRoutes(e, paymentHandler)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
