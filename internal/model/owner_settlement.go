package model

import (
	"time"
)

type OwnerSettlement struct {
	ID             string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BookingID      string     `gorm:"type:uuid;not null;unique" json:"bookingId"`
	OwnerID        string     `gorm:"type:uuid;not null" json:"ownerId"`
	GrossAmount    float64    `gorm:"type:numeric(10,2);not null" json:"grossAmount"`
	PlatformFee    float64    `gorm:"type:numeric(10,2);not null;default:0" json:"platformFee"`
	DiscountAmount float64    `gorm:"type:numeric(10,2);not null;default:0" json:"discountAmount"`
	NetAmount      float64    `gorm:"type:numeric(10,2);not null" json:"netAmount"`
	Status         string     `gorm:"type:varchar(30);not null;default:'pending'" json:"status"`
	AvailableAt    *time.Time `json:"availableAt"`
	PaidAt         *time.Time `json:"paidAt"`
	CreatedAt      time.Time  `gorm:"not null;default:now()" json:"createdAt"`
	UpdatedAt      time.Time  `gorm:"not null;default:now()" json:"updatedAt"`
}
