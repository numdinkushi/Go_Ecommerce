package flutterwave

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	flutterwaveClient "go-ecommerce-app/pkg/external/flutterwave"
	"go-ecommerce-app/pkg/payment"
)

const (
	paymentBaseURL = "https://api.flutterwave.com/v3"
)

// FlutterwaveProvider implements the PaymentProvider interface for Flutterwave
type FlutterwaveProvider struct {
	client     *flutterwaveClient.Client
	secretKey  string
	httpClient *http.Client
}

// NewFlutterwaveProvider creates a new Flutterwave payment provider
func NewFlutterwaveProvider(secretKey string) payment.PaymentProvider {
	return &FlutterwaveProvider{
		client:     flutterwaveClient.NewClient(secretKey),
		secretKey:  secretKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetProviderName returns the provider identifier
func (p *FlutterwaveProvider) GetProviderName() string {
	return "flutterwave"
}

// CreatePaymentRequest represents Flutterwave payment initialization request
type CreatePaymentRequest struct {
	TxRef          string                 `json:"tx_ref"`
	Amount         float64                `json:"amount"`
	Currency       string                 `json:"currency"`
	RedirectURL    string                 `json:"redirect_url"`
	PaymentOptions string                 `json:"payment_options"`
	Customer       map[string]interface{} `json:"customer"`
	Customizations map[string]interface{} `json:"customizations"`
	Meta           map[string]interface{} `json:"meta"`
}

// CreatePaymentResponse represents Flutterwave payment initialization response
type CreatePaymentResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Link string `json:"link"`
	} `json:"data"`
}

// CreatePaymentSession creates a Flutterwave payment session
func (p *FlutterwaveProvider) CreatePaymentSession(req payment.CreatePaymentSessionRequest) (*payment.PaymentSessionResponse, error) {
	txRef := fmt.Sprintf("order_%d_%d_%d", req.OrderID, req.UserID, time.Now().Unix())

	paymentReq := CreatePaymentRequest{
		TxRef:          txRef,
		Amount:         req.Amount,
		Currency:       req.Currency,
		RedirectURL:    req.SuccessURL,
		PaymentOptions: "card,account,ussd,transfer",
		Customer: map[string]interface{}{
			"email": fmt.Sprintf("user_%d@example.com", req.UserID),
		},
		Customizations: map[string]interface{}{
			"title": fmt.Sprintf("Order #%d", req.OrderID),
		},
		Meta: map[string]interface{}{
			"user_id":  req.UserID,
			"order_id": req.OrderID,
		},
	}

	if req.Metadata != nil {
		for k, v := range req.Metadata {
			paymentReq.Meta[k] = v
		}
	}

	jsonData, err := json.Marshal(paymentReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payment request: %w", err)
	}

	url := fmt.Sprintf("%s/payments", paymentBaseURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.secretKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp flutterwaveClient.ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, errorResp.Message)
		}
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var paymentResp CreatePaymentResponse
	if err := json.Unmarshal(body, &paymentResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if paymentResp.Status != "success" {
		return nil, fmt.Errorf("payment creation failed: %s", paymentResp.Message)
	}

	return &payment.PaymentSessionResponse{
		SessionID:   txRef,
		CheckoutURL: paymentResp.Data.Link,
		PaymentID:   txRef,
	}, nil
}

// VerifyTransactionResponse represents Flutterwave transaction verification response
type VerifyTransactionResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID          int                    `json:"id"`
		TxRef       string                 `json:"tx_ref"`
		FlwRef      string                 `json:"flw_ref"`
		Amount      float64                `json:"amount"`
		Currency    string                 `json:"currency"`
		Status      string                 `json:"status"`
		PaymentType string                 `json:"payment_type"`
		CreatedAt   string                 `json:"created_at"`
		Customer    map[string]interface{} `json:"customer"`
		Meta        map[string]interface{} `json:"meta"`
	} `json:"data"`
}

// GetPaymentStatus retrieves the status of a Flutterwave transaction
func (p *FlutterwaveProvider) GetPaymentStatus(txRef string) (*payment.PaymentStatusResponse, error) {
	url := fmt.Sprintf("%s/transactions/verify_by_reference?tx_ref=%s", paymentBaseURL, txRef)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.secretKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp flutterwaveClient.ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, errorResp.Message)
		}
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var verifyResp VerifyTransactionResponse
	if err := json.Unmarshal(body, &verifyResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if verifyResp.Status != "success" {
		return nil, fmt.Errorf("verification failed: %s", verifyResp.Message)
	}

	status := "pending"
	if verifyResp.Data.Status == "successful" {
		status = "succeeded"
	} else if verifyResp.Data.Status == "failed" {
		status = "failed"
	} else if verifyResp.Data.Status == "pending" {
		status = "pending"
	}

	metadata := make(map[string]string)
	if verifyResp.Data.Meta != nil {
		for k, v := range verifyResp.Data.Meta {
			metadata[k] = strconv.FormatFloat(v.(float64), 'f', -1, 64)
			if str, ok := v.(string); ok {
				metadata[k] = str
			}
		}
	}

	return &payment.PaymentStatusResponse{
		Status:    status,
		PaymentID: verifyResp.Data.FlwRef,
		Amount:    verifyResp.Data.Amount,
		Currency:  verifyResp.Data.Currency,
		Metadata:  metadata,
	}, nil
}

// VerifyWebhook verifies the authenticity of a Flutterwave webhook
func (p *FlutterwaveProvider) VerifyWebhook(payload []byte, signature string) error {
	webhookSecret := os.Getenv("FLUTTERWAVE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		return fmt.Errorf("FLUTTERWAVE_WEBHOOK_SECRET not configured")
	}

	hash := sha256.Sum256(payload)
	expectedSignature := hex.EncodeToString(hash[:])

	if signature != expectedSignature {
		return fmt.Errorf("invalid webhook signature")
	}

	return nil
}

// WebhookPayload represents Flutterwave webhook payload
type WebhookPayload struct {
	Event string `json:"event"`
	Data  struct {
		ID          int                    `json:"id"`
		TxRef       string                 `json:"tx_ref"`
		FlwRef      string                 `json:"flw_ref"`
		Amount      float64                `json:"amount"`
		Currency    string                 `json:"currency"`
		Status      string                 `json:"status"`
		PaymentType string                 `json:"payment_type"`
		Meta        map[string]interface{} `json:"meta"`
	} `json:"data"`
}

// ProcessWebhook processes a Flutterwave webhook event
func (p *FlutterwaveProvider) ProcessWebhook(payload []byte, signature string) (*payment.WebhookEvent, error) {
	if err := p.VerifyWebhook(payload, signature); err != nil {
		return nil, err
	}

	var webhookPayload WebhookPayload
	if err := json.Unmarshal(payload, &webhookPayload); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	status := "pending"
	if webhookPayload.Data.Status == "successful" {
		status = "succeeded"
	} else if webhookPayload.Data.Status == "failed" {
		status = "failed"
	}

	metadata := make(map[string]string)
	if webhookPayload.Data.Meta != nil {
		for k, v := range webhookPayload.Data.Meta {
			if str, ok := v.(string); ok {
				metadata[k] = str
			} else {
				metadata[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	return &payment.WebhookEvent{
		EventType: webhookPayload.Event,
		PaymentID: webhookPayload.Data.FlwRef,
		SessionID: webhookPayload.Data.TxRef,
		Status:    status,
		Amount:    webhookPayload.Data.Amount,
		Currency:  webhookPayload.Data.Currency,
		Metadata:  metadata,
	}, nil
}
