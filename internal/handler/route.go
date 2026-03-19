package handler

import (
	"sport-hub-payment/internal/middleware"
	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo, ph *PaymentHandler) {
	e.GET("/", ph.Hello)
	
	v1 := e.Group("/v1")
	v1.Use(middleware.Auth)
	
	payment := v1.Group("/payment")
	{
		payment.POST("/qr/generate", ph.GenerateThaiQR)
	}
}
