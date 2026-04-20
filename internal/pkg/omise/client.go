package omisepkg

import (
	"fmt"
	"os"

	"github.com/omise/omise-go"
	"github.com/omise/omise-go/operations"
)

type Client struct {
	BaseClient *omise.Client
}

func NewClient() (*Client, error) {
	publicKey := os.Getenv("OMISE_PUBLIC_KEY")
	secretKey := os.Getenv("OMISE_SECRET_KEY")

	if secretKey == "" {
		return nil, fmt.Errorf("OMISE_SECRET_KEY is required")
	}

	client, err := omise.NewClient(publicKey, secretKey)
	if err != nil {
		return nil, err
	}

	return &Client{BaseClient: client}, nil
}

func (c *Client) CreatePromptPayCharge(amount int64, bookingID string) (*omise.Charge, error) {
	if c == nil || c.BaseClient == nil {
		return nil, fmt.Errorf("omisepkg client is not initialized")
	}

	// 1. Create Source
	source := &omise.Source{}
	createSource := &operations.CreateSource{
		Amount:   amount,
		Currency: "thb",
		Type:     "promptpay",
	}
	if err := c.BaseClient.Do(source, createSource); err != nil {
		return nil, fmt.Errorf("failed to create source: %w", err)
	}

	// 2. Create Charge using the Source
	charge := &omise.Charge{}
	createCharge := &operations.CreateCharge{
		Amount:   amount,
		Currency: "thb",
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
