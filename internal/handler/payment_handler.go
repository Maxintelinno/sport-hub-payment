package handler

import (
	"net/http"
	"sport-hub-payment/internal/model"
	"sport-hub-payment/internal/service"

	"github.com/labstack/echo/v4"
)

type PaymentHandler struct {
	paymentService service.PaymentService
}

func NewPaymentHandler(ps service.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: ps,
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

func (h *PaymentHandler) Hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
