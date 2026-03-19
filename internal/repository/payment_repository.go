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

type PaymentRepository interface {
	Save(payment *Payment) error
	SavePaymentAndUpdateBooking(payment *Payment) error
	GetBookingOwner(bookingID string) (string, error)
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
