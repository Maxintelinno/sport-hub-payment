package handler

import (
	"io"
	"net/http"
	"sport-hub-payment/internal/model"
	"sport-hub-payment/internal/service"

	"github.com/labstack/echo/v4"
)

type PaymentHandler struct {
	paymentService service.PaymentService
	webhookService service.WebhookService
}

func NewPaymentHandler(ps service.PaymentService, ws service.WebhookService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: ps,
		webhookService: ws,
	}
}

func (h *PaymentHandler) GenerateThaiQR(c echo.Context) error {
	req := new(model.GenerateQRRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	if req.Amount == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "amount is required",
		})
	}

	if req.BookingID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "booking_id is required",
		})
	}

	// Extract userID from context (set by middleware)
	requestUserID := c.Get("user_id").(string)

	resp, err := h.paymentService.GenerateThaiQR(requestUserID, req.BookingID, req.Amount, req.Reference1, req.Reference2)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *PaymentHandler) HandleKBankWebhook(c echo.Context) error {
	// 1. Capture Headers for Logging/Metadata
	headers := make(map[string]string)
	for k, v := range c.Request().Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	// 2. Read Raw Body
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "failed to read body"})
	}

	// 3. Process Webhook
	if err := h.webhookService.HandleKBankWebhook(headers, body); err != nil {
		// Even if processing fails (e.g. payment not found), we should log it
		// and potentially return 200/400 based on KBank requirements.
		// KBank usually expects 200 if the "delivery" was successful.
		// We'll return 200 but keep internal logs for debugging.
		return c.JSON(http.StatusOK, map[string]string{"status": "received_with_error", "details": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "success"})
}

func (h *PaymentHandler) Hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
