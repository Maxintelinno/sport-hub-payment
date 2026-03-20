package repository

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Payment struct {
	ID                    string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	BookingID             string         `gorm:"type:uuid;not null"`
	PaymentNo             string         `gorm:"type:varchar(30);not null;unique"`
	Provider              string         `gorm:"type:varchar(50);not null"`
	Method                string         `gorm:"type:varchar(30);not null"`
	Amount                float64        `gorm:"type:numeric(10,2);not null"`
	Currency              string         `gorm:"type:varchar(10);not null;default:'THB'"`
	Status                string         `gorm:"type:varchar(20);not null;default:'pending'"`
	ProviderPaymentID     *string        `gorm:"type:varchar(100)"`
	ProviderTransactionID *string        `gorm:"type:varchar(150)"`
	ProviderReference     *string        `gorm:"type:varchar(150)"`
	QRPayload             *string        `gorm:"type:text"`
	QRImageURL            *string        `gorm:"type:text"`
	ExpiresAt             time.Time      `gorm:"not null"`
	PaidAt                *time.Time     
	FailedAt              *time.Time     
	FailureReason         *string        `gorm:"type:text"`
	Metadata              json.RawMessage `gorm:"type:jsonb"`
	CreatedAt             time.Time      `gorm:"not null;default:now()"`
	UpdatedAt             time.Time      `gorm:"not null;default:now()"`
}

type PaymentWebhookLog struct {
	ID             string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Provider       string          `gorm:"type:varchar(50);not null"`
	EventType      string          `gorm:"type:varchar(100)"`
	EventID        string          `gorm:"type:varchar(150)"`
	Signature      string          `gorm:"type:text"`
	RequestHeaders json.RawMessage `gorm:"type:jsonb"`
	RequestBody    json.RawMessage `gorm:"type:jsonb;not null"`
	ReceivedAt     time.Time       `gorm:"not null;default:now()"`
	ProcessedAt    *time.Time
	ProcessStatus  string          `gorm:"type:varchar(30);not null;default:'received'"`
	ProcessError   *string         `gorm:"type:text"`
}

type PaymentEvent struct {
	ID          string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Provider    string          `gorm:"type:varchar(50);not null"`
	EventID     string          `gorm:"type:varchar(150);not null;uniqueIndex:idx_provider_event"`
	PaymentID   *string         `gorm:"type:uuid"`
	BookingID   *string         `gorm:"type:uuid"`
	EventType   string          `gorm:"type:varchar(100);not null"`
	EventStatus string          `gorm:"type:varchar(50)"`
	Payload     json.RawMessage `gorm:"type:jsonb;not null"`
	CreatedAt   time.Time       `gorm:"not null;default:now()"`
}

type PaymentRepository interface {
	Save(payment *Payment) error
	SavePaymentAndUpdateBooking(payment *Payment) error
	GetBookingOwner(bookingID string) (string, error)

	// Webhook & Events
	SaveWebhookLog(log *PaymentWebhookLog) error
	IsEventProcessed(provider, eventID string) (bool, error)
	GetPaymentByProviderRef(provider, ref string) (*Payment, error)
	ProcessKBankQRSuccess(payment *Payment, event *PaymentEvent, webhookLogID string) error

	// Polling
	GetPaymentByID(id string) (*Payment, error)
	VerifyPaymentOwner(paymentID, userID string) (bool, error)
}

type gormPaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &gormPaymentRepository{db: db}
}

func (r *gormPaymentRepository) Save(payment *Payment) error {
	return r.db.Create(payment).Error
}

func (r *gormPaymentRepository) SavePaymentAndUpdateBooking(payment *Payment) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Save Payment
		if err := tx.Create(payment).Error; err != nil {
			return err
		}

		// 2. Update Booking
		// Note: Using a map to update specific fields without needing the full Booking model
		result := tx.Table("bookings").Where("id = ?", payment.BookingID).Updates(map[string]interface{}{
			"status":         "pending_payment",
			"payment_status": "unpaid",
			"expires_at":     gorm.Expr("NOW() + INTERVAL '10 minutes'"),
			"updated_at":     gorm.Expr("NOW()"),
		})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return fmt.Errorf("booking not found: %s", payment.BookingID)
		}

		return nil
	})
}

func (r *gormPaymentRepository) GetBookingOwner(bookingID string) (string, error) {
	var userID string
	err := r.db.Table("bookings").Select("user_id").Where("id = ?", bookingID).Scan(&userID).Error
	return userID, err
}

func (r *gormPaymentRepository) SaveWebhookLog(log *PaymentWebhookLog) error {
	return r.db.Create(log).Error
}

func (r *gormPaymentRepository) IsEventProcessed(provider, eventID string) (bool, error) {
	var count int64
	err := r.db.Model(&PaymentEvent{}).Where("provider = ? AND event_id = ?", provider, eventID).Count(&count).Error
	return count > 0, err
}

func (r *gormPaymentRepository) GetPaymentByProviderRef(provider, ref string) (*Payment, error) {
	var payment Payment
	err := r.db.Where("provider = ? AND provider_reference = ?", provider, ref).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *gormPaymentRepository) ProcessKBankQRSuccess(payment *Payment, event *PaymentEvent, webhookLogID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Create Payment Event
		if err := tx.Create(event).Error; err != nil {
			return err
		}

		// 2. Update Payment
		now := time.Now()
		payment.Status = "paid"
		payment.PaidAt = &now
		
		if err := tx.Save(payment).Error; err != nil {
			return err
		}

		// 3. Update Booking
		result := tx.Table("bookings").Where("id = ?", payment.BookingID).Updates(map[string]interface{}{
			"status":         "confirmed",
			"payment_status": "paid",
			"updated_at":     now,
		})
		if result.Error != nil {
			return result.Error
		}

		// 4. Update Webhook Log Status
		if err := tx.Model(&PaymentWebhookLog{}).Where("id = ?", webhookLogID).Updates(map[string]interface{}{
			"processed_at":   now,
			"process_status": "processed",
		}).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *gormPaymentRepository) GetPaymentByID(id string) (*Payment, error) {
	var payment Payment
	err := r.db.Where("id = ?", id).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *gormPaymentRepository) VerifyPaymentOwner(paymentID, userID string) (bool, error) {
	var count int64
	err := r.db.Table("payments").
		Joins("JOIN bookings ON bookings.id = payments.booking_id").
		Where("payments.id = ? AND bookings.user_id = ?", paymentID, userID).
		Count(&count).Error
	
	return count > 0, err
}
