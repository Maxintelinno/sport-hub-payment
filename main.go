package main

import (
	"log"
	"os"
	"sport-hub-payment/internal/database"
	"sport-hub-payment/internal/handler"
	"sport-hub-payment/internal/pkg/kbank"
	omisepkg "sport-hub-payment/internal/pkg/omise"
	"sport-hub-payment/internal/repository"
	"sport-hub-payment/internal/service"
	"strings"

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

	// Support Environment Variables
	viper.AutomaticEnv()
	// Replace dots with underscores (e.g., omise.secretKey -> OMISE_SECRETKEY or OMISE_SECRET_KEY if specifically mapped)
	// We will also use strings.ToUpper if we want to be strict, but AutomaticEnv handles case-insensitivity in some environments.
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Initialize Database with GORM
	db, err := database.InitDB()
	if err != nil {
		log.Printf("Warning: Database connection failed: %v. Running without DB support.", err)
	}

	// Initialize KBank Client with configuration from viper
	kbankClient := kbank.NewClient()

	// Initialize Omise Client
	omiseClient, err := omisepkg.NewClient()
	if err != nil {
		log.Printf("Warning: Omise client initialization failed: %v", err)
	} else {
		sKey := viper.GetString("omise.secretKey")
		if sKey != "" {
			masked := "********"
			if len(sKey) > 4 {
				masked = "..." + sKey[len(sKey)-4:]
			}
			log.Printf("Info: Omise client initialized with secret key ending in %s", masked)
		}
	}

	// Initialize Repository, Service and Handler
	paymentRepo := repository.NewPaymentRepository(db)
	paymentService := service.NewPaymentService(kbankClient, omiseClient, paymentRepo)
	webhookService := service.NewWebhookService(paymentRepo)
	paymentHandler := handler.NewPaymentHandler(paymentService, webhookService)

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())

	// Setup Routes
	handler.SetupRoutes(e, paymentHandler)

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e.Logger.Fatal(e.Start(":" + port))
}
