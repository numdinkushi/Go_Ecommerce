package stripe

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"

	"go-ecommerce-app/pkg/payment"

	"github.com/stripe/stripe-go/v84"
	stripeSession "github.com/stripe/stripe-go/v84/checkout/session"
	"github.com/stripe/stripe-go/v84/webhook"
)

// StripeProvider implements the PaymentProvider interface for Stripe
type StripeProvider struct {
	client    *stripe.Client
	secretKey string
}

// NewStripeProvider creates a new Stripe payment provider
func NewStripeProvider(secretKey string) payment.PaymentProvider {
	// Set the global Stripe API key for package-level functions
	stripe.Key = secretKey
	return &StripeProvider{
		client:    stripe.NewClient(secretKey),
		secretKey: secretKey,
	}
}

// GetProviderName returns the provider identifier
func (p *StripeProvider) GetProviderName() string {
	return "stripe"
}

// CreatePaymentSession creates a Stripe checkout session
func (p *StripeProvider) CreatePaymentSession(req payment.CreatePaymentSessionRequest) (*payment.PaymentSessionResponse, error) {
	amountInCents := int64(req.Amount * 100)

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(req.Currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Order #" + strconv.Itoa(int(req.OrderID))),
					},
					UnitAmount: stripe.Int64(amountInCents),
				},
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(req.SuccessURL),
		CancelURL:  stripe.String(req.CancelURL),
		Metadata: map[string]string{
			"user_id":  strconv.Itoa(int(req.UserID)),
			"order_id": strconv.Itoa(int(req.OrderID)),
		},
	}

	if req.Metadata != nil {
		for k, v := range req.Metadata {
			params.Metadata[k] = v
		}
	}

	// Use package-level function which uses the global API key set in NewStripeProvider
	session, err := stripeSession.New(params)
	if err != nil {
		return nil, err
	}

	paymentID := ""
	if session.PaymentIntent != nil {
		paymentID = session.PaymentIntent.ID
	}

	return &payment.PaymentSessionResponse{
		SessionID:   session.ID,
		CheckoutURL: session.URL,
		PaymentID:   paymentID,
	}, nil
}

// GetPaymentStatus retrieves the status of a Stripe checkout session
func (p *StripeProvider) GetPaymentStatus(sessionID string) (*payment.PaymentStatusResponse, error) {
	// Use package-level function which uses the global API key set in NewStripeProvider
	session, err := stripeSession.Get(sessionID, nil)
	if err != nil {
		return nil, err
	}

	status := "pending"
	if session.PaymentStatus == stripe.CheckoutSessionPaymentStatusPaid {
		status = "succeeded"
	} else if session.PaymentStatus == stripe.CheckoutSessionPaymentStatusUnpaid {
		status = "pending"
	} else if session.PaymentStatus == stripe.CheckoutSessionPaymentStatusNoPaymentRequired {
		status = "succeeded"
	}

	amount := float64(0)
	currency := ""
	if session.AmountTotal > 0 {
		amount = float64(session.AmountTotal) / 100.0
		currency = string(session.Currency)
	}

	paymentID := ""
	if session.PaymentIntent != nil {
		paymentID = session.PaymentIntent.ID
	}

	metadata := make(map[string]string)
	if session.Metadata != nil {
		for k, v := range session.Metadata {
			metadata[k] = v
		}
	}

	return &payment.PaymentStatusResponse{
		Status:    status,
		PaymentID: paymentID,
		Amount:    amount,
		Currency:  currency,
		Metadata:  metadata,
	}, nil
}

// VerifyWebhook verifies the authenticity of a Stripe webhook
func (p *StripeProvider) VerifyWebhook(payload []byte, signature string) error {
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		return errors.New("STRIPE_WEBHOOK_SECRET not configured")
	}

	_, err := webhook.ConstructEvent(payload, signature, webhookSecret)
	if err != nil {
		return err
	}

	return nil
}

// ProcessWebhook processes a Stripe webhook event
func (p *StripeProvider) ProcessWebhook(payload []byte, signature string) (*payment.WebhookEvent, error) {
	if err := p.VerifyWebhook(payload, signature); err != nil {
		return nil, err
	}

	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		return nil, errors.New("STRIPE_WEBHOOK_SECRET not configured")
	}

	event, err := webhook.ConstructEvent(payload, signature, webhookSecret)
	if err != nil {
		return nil, err
	}

	webhookEvent := &payment.WebhookEvent{
		EventType: string(event.Type),
		Metadata:  make(map[string]string),
	}

	if event.Data != nil && event.Data.Object != nil {
		objectBytes, err := json.Marshal(event.Data.Object)
		if err == nil {
			var session stripe.CheckoutSession
			if err := json.Unmarshal(objectBytes, &session); err == nil {
				webhookEvent.SessionID = session.ID
				if session.PaymentIntent != nil {
					webhookEvent.PaymentID = session.PaymentIntent.ID
				}

				switch session.PaymentStatus {
				case stripe.CheckoutSessionPaymentStatusPaid:
					webhookEvent.Status = "succeeded"
				case stripe.CheckoutSessionPaymentStatusUnpaid:
					webhookEvent.Status = "pending"
				}

				if session.AmountTotal > 0 {
					webhookEvent.Amount = float64(session.AmountTotal) / 100.0
					webhookEvent.Currency = string(session.Currency)
				}

				if session.Metadata != nil {
					for k, v := range session.Metadata {
						webhookEvent.Metadata[k] = v
					}
				}
			} else {
				var paymentIntent stripe.PaymentIntent
				if err := json.Unmarshal(objectBytes, &paymentIntent); err == nil {
					webhookEvent.PaymentID = paymentIntent.ID
					switch paymentIntent.Status {
					case stripe.PaymentIntentStatusSucceeded:
						webhookEvent.Status = "succeeded"
					case stripe.PaymentIntentStatusProcessing:
						webhookEvent.Status = "pending"
					default:
						webhookEvent.Status = string(paymentIntent.Status)
					}

					if paymentIntent.Amount > 0 {
						webhookEvent.Amount = float64(paymentIntent.Amount) / 100.0
						webhookEvent.Currency = string(paymentIntent.Currency)
					}
				}
			}
		}
	}

	return webhookEvent, nil
}
