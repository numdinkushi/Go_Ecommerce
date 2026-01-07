package payment

// PaymentProvider defines the interface that all payment providers must implement
type PaymentProvider interface {
	// CreatePaymentSession creates a payment session/checkout session
	// Returns a session ID and redirect URL
	CreatePaymentSession(req CreatePaymentSessionRequest) (*PaymentSessionResponse, error)

	// GetPaymentStatus retrieves the current status of a payment
	GetPaymentStatus(sessionID string) (*PaymentStatusResponse, error)

	// VerifyWebhook verifies the authenticity of a webhook payload
	VerifyWebhook(payload []byte, signature string) error

	// ProcessWebhook processes a webhook event and returns standardized payment info
	ProcessWebhook(payload []byte, signature string) (*WebhookEvent, error)

	// GetProviderName returns the provider identifier (e.g., "stripe", "flutterwave")
	GetProviderName() string
}

// CreatePaymentSessionRequest represents a standardized payment session request
type CreatePaymentSessionRequest struct {
	Amount     float64
	Currency   string
	UserID     uint
	OrderID    uint
	SuccessURL string
	FailureURL string
	Metadata   map[string]string
}

// PaymentSessionResponse represents a standardized payment session response
type PaymentSessionResponse struct {
	SessionID   string
	CheckoutURL string
	PaymentID   string
}

// PaymentStatusResponse represents a standardized payment status response
type PaymentStatusResponse struct {
	Status    string
	PaymentID string
	Amount    float64
	Currency  string
	Metadata  map[string]string
}

// WebhookEvent represents a standardized webhook event
type WebhookEvent struct {
	EventType string
	PaymentID string
	SessionID string
	Status    string
	Amount    float64
	Currency  string
	Metadata  map[string]string
}
