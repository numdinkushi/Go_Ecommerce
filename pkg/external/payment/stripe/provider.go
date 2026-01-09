package stripe

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"go-ecommerce-app/pkg/payment"

	"github.com/stripe/stripe-go/v84"
	stripeSession "github.com/stripe/stripe-go/v84/checkout/session"
	stripePaymentIntent "github.com/stripe/stripe-go/v84/paymentintent"
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

// CreatePaymentSession creates a Stripe checkout session with metadata on both session and payment intent
func (p *StripeProvider) CreatePaymentSession(req payment.CreatePaymentSessionRequest) (*payment.PaymentSessionResponse, error) {
	amountInCents := int64(req.Amount * 100)

	metadata := map[string]string{
		"user_id":  strconv.Itoa(int(req.UserID)),
		"order_id": strconv.Itoa(int(req.OrderID)),
	}

	// Add any additional metadata from request
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			metadata[k] = v
		}
	}

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
		Metadata:   metadata,
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: metadata,
		},
	}

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

// GetPaymentIntentMetadata fetches a payment intent by ID to get its metadata
func (p *StripeProvider) GetPaymentIntentMetadata(paymentIntentID string) (map[string]string, error) {
	pi, err := stripePaymentIntent.Get(paymentIntentID, nil)
	if err != nil {
		return nil, err
	}

	metadata := make(map[string]string)
	if pi.Metadata != nil {
		for k, v := range pi.Metadata {
			metadata[k] = v
		}
	}

	return metadata, nil
}

// VerifyWebhook verifies the authenticity of a Stripe webhook
func (p *StripeProvider) VerifyWebhook(payload []byte, signature string) error {
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		return errors.New("STRIPE_WEBHOOK_SECRET not configured")
	}

	_, err := webhook.ConstructEventWithOptions(payload, signature, webhookSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		log.Printf("Webhook signature verification failed: %v", err)
		return err
	}

	return nil
}

// ProcessWebhook processes a Stripe webhook event
func (p *StripeProvider) ProcessWebhook(payload []byte, signature string) (*payment.WebhookEvent, error) {
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		return nil, errors.New("STRIPE_WEBHOOK_SECRET not configured")
	}

	event, err := webhook.ConstructEventWithOptions(payload, signature, webhookSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		log.Printf("Webhook signature verification failed: %v", err)
		return nil, err
	}

	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	isLiveMode := stripeKey != "" && !strings.HasPrefix(stripeKey, "sk_test_")
	eventIsLive := event.Livemode

	if isLiveMode != eventIsLive {
		log.Printf("Webhook mode mismatch - Event is from %s mode but environment is %s mode, Event ID: %s",
			map[bool]string{true: "LIVE", false: "TEST"}[eventIsLive],
			map[bool]string{true: "LIVE", false: "TEST"}[isLiveMode],
			event.ID)
	}

	webhookEvent := &payment.WebhookEvent{
		EventType: string(event.Type),
		Metadata:  make(map[string]string),
	}

	if event.Data != nil && event.Data.Object != nil {
		objectBytes, err := json.Marshal(event.Data.Object)
		if err == nil {
			// Determine event type and parse accordingly
			eventType := string(event.Type)

			// checkout.session.completed events contain CheckoutSession
			if eventType == "checkout.session.completed" {
				var session stripe.CheckoutSession
				if err := json.Unmarshal(objectBytes, &session); err == nil && session.ID != "" {
					webhookEvent.SessionID = session.ID
					if session.PaymentIntent != nil {
						webhookEvent.PaymentID = session.PaymentIntent.ID
					}

					// Verify payment status is actually "paid" before marking as succeeded
					if session.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid && session.PaymentStatus != "" {
						log.Printf("Checkout session completed but payment status is '%s', not 'paid'", session.PaymentStatus)
					}

					switch session.PaymentStatus {
					case stripe.CheckoutSessionPaymentStatusPaid:
						webhookEvent.Status = "succeeded"
					case stripe.CheckoutSessionPaymentStatusUnpaid:
						webhookEvent.Status = "pending"
					case stripe.CheckoutSessionPaymentStatusNoPaymentRequired:
						webhookEvent.Status = "succeeded"
					default:
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
				}
			} else if eventType == "payment_intent.succeeded" || eventType == "payment_intent.created" || eventType == "payment_intent.payment_failed" {
				// payment_intent.* events contain PaymentIntent
				var paymentIntent stripe.PaymentIntent
				if err := json.Unmarshal(objectBytes, &paymentIntent); err == nil && paymentIntent.ID != "" {
					webhookEvent.PaymentID = paymentIntent.ID

					// Verify PaymentIntent status matches event type
					if eventType == "payment_intent.succeeded" {
						if paymentIntent.Status != stripe.PaymentIntentStatusSucceeded {
							// Fetch fresh status from Stripe to verify
							freshPI, fetchErr := stripePaymentIntent.Get(paymentIntent.ID, nil)
							if fetchErr == nil {
								if freshPI.Status != stripe.PaymentIntentStatusSucceeded {
									return nil, fmt.Errorf("payment intent %s is not actually succeeded (status: %s)", paymentIntent.ID, freshPI.Status)
								}
							}
						}

						if paymentIntent.Status == stripe.PaymentIntentStatusRequiresConfirmation {
							return nil, errors.New("payment intent is in requires_confirmation state - manual confirmation may be required")
						}
					}

					switch paymentIntent.Status {
					case stripe.PaymentIntentStatusSucceeded:
						webhookEvent.Status = "succeeded"
					case stripe.PaymentIntentStatusProcessing:
						webhookEvent.Status = "pending"
					case stripe.PaymentIntentStatusRequiresConfirmation:
						webhookEvent.Status = "pending"
					default:
						webhookEvent.Status = string(paymentIntent.Status)
					}

					if paymentIntent.Amount > 0 {
						webhookEvent.Amount = float64(paymentIntent.Amount) / 100.0
						webhookEvent.Currency = string(paymentIntent.Currency)
					}

					if len(paymentIntent.Metadata) > 0 {
						for k, v := range paymentIntent.Metadata {
							webhookEvent.Metadata[k] = v
						}
					}
				}
			} else if eventType == "charge.succeeded" || eventType == "charge.updated" || eventType == "charge.failed" {
				// charge.* events contain Charge
				var charge stripe.Charge
				if err := json.Unmarshal(objectBytes, &charge); err == nil && charge.ID != "" {
					// Charge events don't have SessionID, only PaymentIntent
					if charge.PaymentIntent != nil {
						webhookEvent.PaymentID = charge.PaymentIntent.ID
					}

					switch charge.Status {
					case stripe.ChargeStatusSucceeded:
						webhookEvent.Status = "succeeded"
					case stripe.ChargeStatusPending:
						webhookEvent.Status = "pending"
					case stripe.ChargeStatusFailed:
						webhookEvent.Status = "failed"
					default:
						webhookEvent.Status = string(charge.Status)
					}

					if charge.Amount > 0 {
						webhookEvent.Amount = float64(charge.Amount) / 100.0
						webhookEvent.Currency = string(charge.Currency)
					}

					// Extract metadata from charge
					if charge.Metadata != nil {
						for k, v := range charge.Metadata {
							webhookEvent.Metadata[k] = v
						}
					}
				}
			}
		}
	}

	return webhookEvent, nil
}
