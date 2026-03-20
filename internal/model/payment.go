package model

type GenerateQRRequest struct {
	BookingID  string `json:"booking_id"`
	Amount     string `json:"amount"`
	Reference1 string `json:"reference1"`
	Reference2 string `json:"reference2"`
}

type QRResponse struct {
	PaymentID string `json:"paymentId"`
	QrCode    string `json:"qrCode"`
	Status    string `json:"status"`
}

type KBankWebhookRequest struct {
	EventType       string `json:"eventType"`
	EventID         string `json:"eventId"`
	Status          string `json:"status"`
	PartnerTxnUid   string `json:"partnerTxnUid"`
	TransactionID   string `json:"transactionId"`
	Amount          string `json:"amount"`
	Currency        string `json:"currency"`
	PaymentDateTime string `json:"paymentDateTime"`
}
