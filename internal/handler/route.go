package handler

import (
	"sport-hub-payment/internal/middleware"
	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo, ph *PaymentHandler) {
	e.GET("/", ph.Hello)
	
	v1 := e.Group("/v1")
	v1.Use(middleware.Auth)

	v1.POST("/payments/webhooks/kbank", ph.HandleKBankWebhook)

	payment := v1.Group("/payment")
	{
		payment.POST("/qr/generate", ph.GenerateThaiQR)

		// [DIAGNOSTIC] Check if the path is reachable via GET
		payment.GET("/qr/generate", func(c echo.Context) error {
			return c.String(200, "Reached: GET /v1/payment/qr/generate")
		})
	}

	// [DIAGNOSTIC] Check if the v1 group and middleware are working
	v1.GET("/test", func(c echo.Context) error {
		return c.String(200, "Reached: GET /v1/test")
	})
}
