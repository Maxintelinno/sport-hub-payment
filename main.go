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
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Explicitly bind environment variables to ensure they are picked up even if config is missing
	viper.BindEnv("omise.public_key", "OMISE_PUBLIC_KEY")
	viper.BindEnv("omise.secret_key", "OMISE_SECRET_KEY")

	// Diagnostics: Check direct os.Getenv visibility on Railway
	if os.Getenv("OMISE_SECRET_KEY") != "" {
		log.Printf("Debug: Direct os.Getenv(OMISE_SECRET_KEY) found: true")
	} else {
		log.Printf("Debug: Direct os.Getenv(OMISE_SECRET_KEY) found: false")
	}

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
		sKey := viper.GetString("omise.secret_key")
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
