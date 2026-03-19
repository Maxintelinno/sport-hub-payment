package model

type GenerateQRRequest struct {
	BookingID  string `json:"booking_id"`
	Amount     string `json:"amount"`
	Reference1 string `json:"reference1"`
	Reference2 string `json:"reference2"`
}

type QRResponse struct {
	QrCode string `json:"qrCode"`
	Status string `json:"status"`
}
