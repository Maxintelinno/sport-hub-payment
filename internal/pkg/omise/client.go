package omise

import (
	"fmt"
	"github.com/omise/omise-go"
	"github.com/omise/omise-go/operations"
	"github.com/spf13/viper"
)

type Client struct {
	BaseClient *omise.Client
}

func NewClient() (*Client, error) {
	publicKey := viper.GetString("omise.publicKey")
	secretKey := viper.GetString("omise.secretKey")

	if secretKey == "" {
		return nil, fmt.Errorf("omise.secretKey is required")
	}

	client, err := omise.NewClient(publicKey, secretKey)
	if err != nil {
		return nil, err
	}

	return &Client{BaseClient: client}, nil
}

func (c *Client) CreatePromptPayCharge(amount int64, bookingID string) (*omise.Charge, error) {
	// 1. Create Source
	source := &omise.Source{}
	createSource := &operations.CreateSource{
		Amount:   amount,
		Currency: "THB",
		Type:     "promptpay",
	}
	if err := c.BaseClient.Do(source, createSource); err != nil {
		return nil, fmt.Errorf("failed to create source: %w", err)
	}

	// 2. Create Charge using the Source
	charge := &omise.Charge{}
	createCharge := &operations.CreateCharge{
		Amount:   amount,
		Currency: "THB",
		Source:   source.ID,
		Metadata: map[string]interface{}{
			"booking_id": bookingID,
		},
	}
	if err := c.BaseClient.Do(charge, createCharge); err != nil {
		return nil, fmt.Errorf("failed to create charge: %w", err)
	}

	return charge, nil
}
