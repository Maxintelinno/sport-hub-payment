package service

import (
	"encoding/json"
	"fmt"
	"sport-hub-payment/internal/model"
	"sport-hub-payment/internal/repository"
)

type WebhookService interface {
	HandleKBankWebhook(headers map[string]string, body []byte) error
	HandleOmiseWebhook(body []byte) error
}

type webhookService struct {
	repo repository.PaymentRepository
}

func NewWebhookService(repo repository.PaymentRepository) WebhookService {
	return &webhookService{repo: repo}
}

func (s *webhookService) HandleKBankWebhook(headers map[string]string, body []byte) error {
	// 1. Parse Request Body
	var req model.KBankWebhookRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return fmt.Errorf("failed to unmarshal webhook body: %w", err)
	}

	// 2. Prepare Log
	headersJSON, _ := json.Marshal(headers)
	webhookLog := &repository.PaymentWebhookLog{
		Provider:       "kbank",
		EventType:      req.EventType,
		EventID:        req.EventID,
		Signature:      headers["x-signature"],
		RequestHeaders: headersJSON,
		RequestBody:    body,
		ProcessStatus:  "received",
	}

	// 3. Save Raw Log
	if err := s.repo.SaveWebhookLog(webhookLog); err != nil {
		return fmt.Errorf("failed to save webhook log: %w", err)
	}

	// 4. Check Idempotency
	isProcessed, err := s.repo.IsEventProcessed("kbank", req.EventID)
	if err != nil {
		return s.markLogFailed(webhookLog.ID, fmt.Sprintf("idempotency check failed: %v", err))
	}
	if isProcessed {
		return fmt.Errorf("event already processed (eventId: %s)", req.EventID)
	}

	// 5. Look up Payment by partnerTxnUid (reference)
	payment, err := s.repo.GetPaymentByProviderRef("kbank", req.PartnerTxnUid)
	if err != nil {
		return s.markLogFailed(webhookLog.ID, fmt.Sprintf("payment not found for ref %s: %v", req.PartnerTxnUid, err))
	}

	// 6. Process based on Status
	if req.Status == "SUCCESS" {
		event := &repository.PaymentEvent{
			Provider:    "kbank",
			EventID:     req.EventID,
			PaymentID:   &payment.ID,
			BookingID:   &payment.BookingID,
			EventType:   req.EventType,
			EventStatus: req.Status,
			Payload:     body,
		}

		// Assign actual Provider Transaction ID from webhook
		payment.ProviderTransactionID = &req.TransactionID

		if err := s.repo.ProcessKBankQRSuccess(payment, event, webhookLog.ID); err != nil {
			return s.markLogFailed(webhookLog.ID, fmt.Sprintf("failed to process success event: %v", err))
		}
	} else {
		return fmt.Errorf("ignored status: %s (only SUCCESS is processed)", req.Status)
	}

	return nil
}

func (s *webhookService) HandleOmiseWebhook(body []byte) error {
	// 1. Parse Event
	var event model.OmiseWebhookRequest
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("failed to unmarshal omise webhook: %w", err)
	}

	// 2. Prepare Log
	webhookLog := &repository.PaymentWebhookLog{
		Provider:      "omise",
		EventType:     event.Key,
		EventID:       event.ID,
		RequestBody:   body,
		ProcessStatus: "received",
	}

	if err := s.repo.SaveWebhookLog(webhookLog); err != nil {
		return fmt.Errorf("failed to save webhook log: %w", err)
	}

	// 3. Process Charge Event
	if event.Key == "charge.complete" {
		var charge struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}
		if err := json.Unmarshal(event.Data, &charge); err != nil {
			return s.markLogFailed(webhookLog.ID, fmt.Sprintf("failed to parse charge data: %v", err))
		}

		// Look up payment
		payment, err := s.repo.GetPaymentByProviderPaymentID("omise", charge.ID)
		if err != nil {
			return s.markLogFailed(webhookLog.ID, fmt.Sprintf("payment not found for charge %s: %v", charge.ID, err))
		}

		if charge.Status == "successful" {
			paymentEvent := &repository.PaymentEvent{
				Provider:    "omise",
				EventID:     event.ID,
				PaymentID:   &payment.ID,
				BookingID:   &payment.BookingID,
				EventType:   event.Key,
				EventStatus: charge.Status,
				Payload:     body,
			}

			if err := s.repo.ProcessOmiseSuccess(payment, paymentEvent, webhookLog.ID); err != nil {
				return s.markLogFailed(webhookLog.ID, fmt.Sprintf("failed to process success event: %v", err))
			}
		} else {
			// Handle failed charge if needed
			return s.markLogFailed(webhookLog.ID, fmt.Sprintf("charge status ignored: %s", charge.Status))
		}
	}

	return nil
}

func (s *webhookService) markLogFailed(id, reason string) error {
	// Implementation to update log status to failed
	// For now, returning the error to the handler
	return fmt.Errorf("%s", reason)
}

func (s *webhookService) markLogSuccess(id, note string) error {
	// Implementation to update log status to success/skipped
	return nil
}
