package services

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/utils"
	"github.com/sirupsen/logrus"
)

type MpesaService struct {
	BaseURL        string
	ConsumerKey    string
	ConsumerSecret string
	Shortcode      string
	Passkey        string
	CallbackURL    string
	Logger         *logrus.Logger
}

type STKPushRequest struct {
	BusinessShortCode string `json:"BusinessShortCode"`
	Password          string `json:"Password"`
	Timestamp         string `json:"Timestamp"`
	TransactionType   string `json:"TransactionType"`
	Amount            string `json:"Amount"`
	PartyA            string `json:"PartyA"`
	PartyB            string `json:"PartyB"`
	PhoneNumber       string `json:"PhoneNumber"`
	CallBackURL       string `json:"CallBackURL"`
	AccountReference  string `json:"AccountReference"`
	TransactionDesc   string `json:"TransactionDesc"`
}

type STKPushResponse struct {
	MerchantRequestID   string `json:"MerchantRequestID"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	CustomerMessage     string `json:"CustomerMessage"`
}

type OAuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
}

func NewMpesaService() *MpesaService {
	return &MpesaService{
		BaseURL:        "https://sandbox.safaricom.co.ke",
		ConsumerKey:    os.Getenv("MPESA_CONSUMER_KEY"),
		ConsumerSecret: os.Getenv("MPESA_CONSUMER_SECRET"),
		Shortcode:      os.Getenv("MPESA_SHORTCODE"),
		Passkey:        os.Getenv("MPESA_PASSKEY"),
		CallbackURL:    os.Getenv("MPESA_CALLBACK_URL"),
		Logger:         utils.Logger,
	}
}

func (m *MpesaService) GetAccessToken() (string, error) {
	url := fmt.Sprintf("%s/oauth/v1/generate?grant_type=client_credentials", m.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(m.ConsumerKey, m.ConsumerSecret)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("failed to get access token: status=%d body=%s", resp.StatusCode, string(body))
	}

	var oauthResp OAuthResponse
	if err := json.Unmarshal(body, &oauthResp); err != nil {
		return "", err
	}

	if oauthResp.AccessToken == "" {
		return "", fmt.Errorf("failed to get access token: %s", string(body))
	}

	return oauthResp.AccessToken, nil
}

func (m *MpesaService) GeneratePassword(timestamp string) string {
	data := m.Shortcode + m.Passkey + timestamp
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func (m *MpesaService) InitiateSTKPush(phoneNumber string, amount float64, accountReference string) (*STKPushResponse, error) {
	accessToken, err := m.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// for testing purposes, set amount to 5
	amount = 1

	timestamp := time.Now().Format("20060102150405")
	password := m.GeneratePassword(timestamp)

	url := fmt.Sprintf("%s/mpesa/stkpush/v1/processrequest", m.BaseURL)

	request := STKPushRequest{
		BusinessShortCode: m.Shortcode,
		Password:          password,
		Timestamp:         timestamp,
		TransactionType:   "CustomerPayBillOnline",
		Amount:            fmt.Sprintf("%.0f", amount),
		PartyA:            phoneNumber,
		PartyB:            m.Shortcode,
		PhoneNumber:       phoneNumber,
		CallBackURL:       m.CallbackURL,
		AccountReference:  accountReference,
		TransactionDesc:   "Payment for Smart Retail System",
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("stk push request failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var response STKPushResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	if response.ResponseCode != "0" && response.ResponseCode != "" {
		return nil, fmt.Errorf("stk push rejected: code=%s description=%s", response.ResponseCode, response.ResponseDescription)
	}

	m.Logger.WithFields(logrus.Fields{
		"checkout_request_id": response.CheckoutRequestID,
		"merchant_request_id": response.MerchantRequestID,
		"response_code":       response.ResponseCode,
		"phone":               phoneNumber,
		"amount":              amount,
	}).Info("STK Push initiated")

	return &response, nil
}

func (m *MpesaService) SimulatePayment(phoneNumber string, amount float64) error {
	accessToken, err := m.GetAccessToken()
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	url := fmt.Sprintf("%s/mpesa/stkpush/v1/processrequest", m.BaseURL)

	timestamp := time.Now().Format("20060102150405")
	password := m.GeneratePassword(timestamp)

	request := STKPushRequest{
		BusinessShortCode: "174379", // Default sandbox shortcode
		Password:          password,
		Timestamp:         timestamp,
		TransactionType:   "CustomerPayBillOnline",
		Amount:            fmt.Sprintf("%.0f", amount),
		PartyA:            phoneNumber,
		PartyB:            "174379",
		PhoneNumber:       phoneNumber,
		CallBackURL:       m.CallbackURL,
		AccountReference:  "SMART_RETAIL_" + generateRandomString(8),
		TransactionDesc:   "Smart Retail Payment",
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func generateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return strconv.FormatInt(int64(b[0]), 36)
}
