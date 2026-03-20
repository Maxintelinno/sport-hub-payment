package kbank

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/spf13/viper"
)

type Client struct {
	BaseURL       string
	AuthHeader    string
	PartnerId     string
	PartnerSecret string
	MerchantId    string
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type QRRequest struct {
	PartnerTxnUid   string `json:"partnerTxnUid"`
	PartnerId       string `json:"partnerId"`
	PartnerSecret   string `json:"partnerSecret"`
	RequestDt       string `json:"requestDt"`
	MerchantId      string `json:"merchantId"`
	TerminalId      string `json:"terminalId"`
	QrType          string `json:"qrType"`
	TxnAmount       string `json:"txnAmount"`
	TxnCurrencyCode string `json:"txnCurrencyCode"`
	Reference1      string `json:"reference1"`
	Reference2      string `json:"reference2"`
	Reference3      string `json:"reference3"`
	Reference4      string `json:"reference4"`
	Metadata        string `json:"metadata"`
}

type QRResponse struct {
	PartnerTxnUid string   `json:"partnerTxnUid"`
	PartnerId     string   `json:"partnerId"`
	StatusCode    string   `json:"statusCode"`
	ErrorCode     string   `json:"errorCode"`
	ErrorDesc     string   `json:"errorDesc"`
	AccountName   string   `json:"accountName"`
	QrCode        string   `json:"qrCode"`
	Sof           []string `json:"sof"`
}

func NewClient() *Client {
	return &Client{
		BaseURL:       viper.GetString("kbankPayment.baseUrl"),
		AuthHeader:    "Basic " + base64.StdEncoding.EncodeToString([]byte(viper.GetString("kbankPayment.consumerKey")+":"+viper.GetString("kbankPayment.consumerSecret"))),
		PartnerId:     viper.GetString("kbankPayment.partnerId"),
		PartnerSecret: viper.GetString("kbankPayment.partnerSecret"),
		MerchantId:    viper.GetString("kbankPayment.merchantId"),
	}
}

func (c *Client) GetAccessToken() (string, error) {
	u := fmt.Sprintf("%s/v2/oauth/token", c.BaseURL)
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", u, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", c.AuthHeader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("env-id", "OAUTH2")
	req.Header.Set("x-test-mode", "true")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get token: %s", string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

func (c *Client) GenerateThaiQR(requestUserID, bookingID, amount string, reference1, reference2 string) (*QRResponse, error) {
	token, err := c.GetAccessToken()
	if err != nil {
		return nil, err
	}

	u := fmt.Sprintf("%s/v1/qrpayment/request", c.BaseURL)
	//txnUid := fmt.Sprintf("PARTNERTEST%04d", time.Now().Unix()%10000)

	qrReq := QRRequest{
		PartnerTxnUid:   "PARTNERTEST0001",
		PartnerId:       c.PartnerId,
		PartnerSecret:   c.PartnerSecret,
		RequestDt:       time.Now().Format("2006-01-02T15:04:05-07:00"),
		MerchantId:      c.MerchantId,
		TerminalId:      viper.GetString("kbankPayment.terminalId"),
		QrType:          "3",
		TxnAmount:       amount,
		TxnCurrencyCode: "THB",
		Reference1:      "INV001",
		Reference2:      "HELLOWORLD",
		Reference3:      "INV001",
		Reference4:      "INV001",
		Metadata:        "test QR",
	}

	body, err := json.Marshal(qrReq)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", u, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-test-mode", "true")
	req.Header.Set("env-id", "QR002")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to generate QR: %s", string(respBody))
	}

	var qrResp QRResponse
	if err := json.NewDecoder(resp.Body).Decode(&qrResp); err != nil {
		return nil, err
	}

	return &qrResp, nil
}
