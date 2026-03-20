package service

import (
	"encoding/json"
	"fmt"
	"sport-hub-payment/internal/model"
	"sport-hub-payment/internal/pkg/kbank"
	"sport-hub-payment/internal/repository"
	"strconv"
	"time"
)

type PaymentService interface {
	GenerateThaiQR(requestUserID, bookingID, amount, ref1, ref2 string) (*model.QRResponse, error)
	GetPaymentStatus(paymentID, userID string) (*repository.Payment, error)
}

type paymentService struct {
	kbankClient *kbank.Client
	repo        repository.PaymentRepository
}

func NewPaymentService(client *kbank.Client, repo repository.PaymentRepository) PaymentService {
	return &paymentService{
		kbankClient: client,
		repo:        repo,
	}
}

func (s *paymentService) GenerateThaiQR(requestUserID, bookingID, amount, ref1, ref2 string) (*model.QRResponse, error) {
	// 1. Validate Ownership: Check if the requestUserID matches the booking owner
	bookingOwner, err := s.repo.GetBookingOwner(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify booking owner: %w", err)
	}

	if bookingOwner == "" {
		return nil, fmt.Errorf("booking not found: %s", bookingID)
	}

	if bookingOwner != requestUserID {
		return nil, fmt.Errorf("unauthorized: user %s does not own booking %s", requestUserID, bookingID)
	}

	// 2. Generate QR Code from KBank
	resp, err := s.kbankClient.GenerateThaiQR(requestUserID, bookingID, amount, ref1, ref2)
	if err != nil {
		return nil, err
	}

	// 2. Prepare payment record for GORM
	paymentNo := fmt.Sprintf("PAY-%d", time.Now().UnixNano()/1e6)

	amountF, _ := strconv.ParseFloat(amount, 64)

	metadata := map[string]interface{}{
		"partnerId":   resp.PartnerId,
		"merchantId":  s.kbankClient.MerchantId,
		"terminalId":  "term1",
		"accountName": resp.AccountName,
		"sof":         resp.Sof,
	}
	metadataJSON, _ := json.Marshal(metadata)

	payment := &repository.Payment{
		BookingID:         bookingID,
		PaymentNo:         paymentNo,
		Provider:          "kbank",
		Method:            "promptpay_qr",
		Amount:            amountF,
		Currency:          "THB",
		Status:            "pending",
		ProviderReference: &resp.PartnerTxnUid,
		QRPayload:         &resp.QrCode,
		ExpiresAt:         time.Now().Add(10 * time.Minute),
		Metadata:          metadataJSON,
	}

	// 3. Save to database using transactional repository method
	if err := s.repo.SavePaymentAndUpdateBooking(payment); err != nil {
		return nil, fmt.Errorf("failed to save payment and update booking: %w", err)
	}

	return &model.QRResponse{
		PaymentID: payment.ID,
		QrCode:    resp.QrCode,
		Status:    "pending", // Initial status
	}, nil
}

func (s *paymentService) GetPaymentStatus(paymentID, userID string) (*repository.Payment, error) {
	// 1. Get Payment first to see if it exists
	payment, err := s.repo.GetPaymentByID(paymentID)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}

	// 2. Verify Ownership
	isOwner, err := s.repo.VerifyPaymentOwner(paymentID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify payment owner: %w", err)
	}

	if !isOwner {
		return nil, fmt.Errorf("unauthorized: user %s does not own payment %s", userID, paymentID)
	}

	return payment, nil
}
