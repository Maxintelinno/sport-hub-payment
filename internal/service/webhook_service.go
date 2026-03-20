package service

import (
	"encoding/json"
	"fmt"
	"sport-hub-payment/internal/model"
	"sport-hub-payment/internal/repository"
)

type WebhookService interface {
	HandleKBankWebhook(headers map[string]string, body []byte) error
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
		return s.markLogSuccess(webhookLog.ID, "event already processed")
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

		if err := s.repo.ProcessKBankQRSuccess(payment, event, webhookLog.ID); err != nil {
			return s.markLogFailed(webhookLog.ID, fmt.Sprintf("failed to process success event: %v", err))
		}
	} else {
		return s.markLogSuccess(webhookLog.ID, fmt.Sprintf("ignored status: %s", req.Status))
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
