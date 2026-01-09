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

// CreatePaymentSession creates a Stripe checkout session
// Best practice: Sets metadata on both session AND payment intent for reliable webhook matching
func (p *StripeProvider) CreatePaymentSession(req payment.CreatePaymentSessionRequest) (*payment.PaymentSessionResponse, error) {
	amountInCents := int64(req.Amount * 100)

	// Build metadata (will be set on both session and payment intent)
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
		// CRITICAL: Set metadata on PaymentIntent via payment_intent_data
		// This ensures payment_intent.succeeded events can match by metadata order_id
		PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
			Metadata: metadata,
		},
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

	log.Printf("✅ Checkout session created with metadata on both session and payment intent")
	log.Printf("   SessionID: %s, PaymentID: %s, Metadata: %v", session.ID, paymentID, metadata)

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

	// Use ConstructEventWithOptions to ignore API version mismatch
	// This allows webhooks with different API versions to be processed
	_, err := webhook.ConstructEventWithOptions(payload, signature, webhookSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		log.Printf("🚨 Webhook signature verification failed: %v", err)
		log.Printf("   Check that STRIPE_WEBHOOK_SECRET matches the signing secret in Stripe Dashboard")
		return err
	}

	return nil
}

// ProcessWebhook processes a Stripe webhook event
// Following Stripe's Go quickstart: https://docs.stripe.com/webhooks/quickstart?lang=go
func (p *StripeProvider) ProcessWebhook(payload []byte, signature string) (*payment.WebhookEvent, error) {
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		return nil, errors.New("STRIPE_WEBHOOK_SECRET not configured")
	}

	// ConstructEvent verifies the signature AND constructs the event in one step
	// Following Stripe's recommended pattern from their Go quickstart
	event, err := webhook.ConstructEventWithOptions(payload, signature, webhookSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		log.Printf("🚨 Webhook signature verification failed: %v", err)
		log.Printf("   Check that STRIPE_WEBHOOK_SECRET matches the signing secret in Stripe Dashboard")
		return nil, err
	}

	// ChatGPT Checklist #7: Verify webhook mode (test vs live) matches environment
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	isLiveMode := stripeKey != "" && !strings.HasPrefix(stripeKey, "sk_test_")
	eventIsLive := event.Livemode

	if isLiveMode != eventIsLive {
		log.Printf("⚠️  WARNING: Webhook mode mismatch - Event is from %s mode but environment is %s mode",
			map[bool]string{true: "LIVE", false: "TEST"}[eventIsLive],
			map[bool]string{true: "LIVE", false: "TEST"}[isLiveMode])
		log.Printf("   Event ID: %s", event.ID)
		log.Printf("   Check: Stripe Dashboard webhook configuration matches your STRIPE_SECRET_KEY mode")
	}

	webhookEvent := &payment.WebhookEvent{
		EventType: string(event.Type),
		Metadata:  make(map[string]string),
	}

	log.Printf("📋 Processing Stripe webhook event: %s (mode: %s)", event.Type, map[bool]string{true: "LIVE", false: "TEST"}[eventIsLive])

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
					log.Printf("   ✅ Parsed as CheckoutSession - SessionID: %s, PaymentID: %s, PaymentStatus: %s",
						webhookEvent.SessionID, webhookEvent.PaymentID, session.PaymentStatus)

					// CRITICAL: Verify session payment status is actually "paid" before marking as succeeded
					// ChatGPT Checklist #1: Payment may be successful on client but not finalized on Stripe
					if session.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
						log.Printf("   ⚠️  WARNING: checkout.session.completed event received but PaymentStatus is '%s', not 'paid'", session.PaymentStatus)
						log.Printf("   ⚠️  Payment may not be finalized. Only marking as succeeded if status is 'paid'")
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
						log.Printf("   ⚠️  Unknown payment status: %s", session.PaymentStatus)
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
					// PaymentIntent doesn't have SessionID - we need to get it from metadata or find the session
					log.Printf("   ✅ Parsed as PaymentIntent - PaymentID: %s, Status: %s", webhookEvent.PaymentID, paymentIntent.Status)

					// CRITICAL: Verify PaymentIntent is actually succeeded, not just requires_confirmation
					// ChatGPT Checklist #1: Payment may be successful on client but not finalized on Stripe
					if eventType == "payment_intent.succeeded" {
						if paymentIntent.Status != stripe.PaymentIntentStatusSucceeded {
							log.Printf("   ⚠️  WARNING: Received payment_intent.succeeded event but PaymentIntent status is '%s', not 'succeeded'", paymentIntent.Status)
							log.Printf("   ⚠️  This payment may not be finalized. Checking status from Stripe API...")

							// Fetch fresh PaymentIntent from Stripe to verify actual status
							freshPI, fetchErr := stripePaymentIntent.Get(paymentIntent.ID, nil)
							if fetchErr == nil {
								log.Printf("   📊 Fresh PaymentIntent status from Stripe: %s", freshPI.Status)
								if freshPI.Status != stripe.PaymentIntentStatusSucceeded {
									return nil, fmt.Errorf("payment intent %s is not actually succeeded (status: %s). Payment may not be finalized", paymentIntent.ID, freshPI.Status)
								}
							} else {
								log.Printf("   ⚠️  Could not fetch fresh PaymentIntent: %v", fetchErr)
								// Continue with event status but log warning
							}
						}

						// Additional check: Ensure PaymentIntent is not stuck in requires_confirmation
						if paymentIntent.Status == stripe.PaymentIntentStatusRequiresConfirmation {
							log.Printf("   ❌ PaymentIntent is in requires_confirmation state - payment not completed")
							return nil, errors.New("payment intent is in requires_confirmation state - manual confirmation may be required")
						}
					}

					switch paymentIntent.Status {
					case stripe.PaymentIntentStatusSucceeded:
						webhookEvent.Status = "succeeded"
					case stripe.PaymentIntentStatusProcessing:
						webhookEvent.Status = "pending"
					case stripe.PaymentIntentStatusRequiresConfirmation:
						// Should not reach here if we handled it above, but just in case
						webhookEvent.Status = "pending"
						log.Printf("   ⚠️  PaymentIntent requires confirmation - payment not finalized")
					default:
						webhookEvent.Status = string(paymentIntent.Status)
					}

					if paymentIntent.Amount > 0 {
						webhookEvent.Amount = float64(paymentIntent.Amount) / 100.0
						webhookEvent.Currency = string(paymentIntent.Currency)
					}

					// CRITICAL: Extract metadata from payment intent
					// With payment_intent_data.metadata set during session creation, this will contain order_id
					if len(paymentIntent.Metadata) > 0 {
						for k, v := range paymentIntent.Metadata {
							webhookEvent.Metadata[k] = v
						}
						log.Printf("   ✅ Extracted metadata from PaymentIntent: %v", paymentIntent.Metadata)
					} else {
						log.Printf("   ⚠️  PaymentIntent has no metadata - webhook matching will rely on PaymentID only")
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
					log.Printf("   ✅ Parsed as Charge - PaymentID: %s", webhookEvent.PaymentID)

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

	log.Printf("📤 Webhook event parsed - EventType: %s, PaymentID: %s, SessionID: %s, Metadata: %v",
		webhookEvent.EventType, webhookEvent.PaymentID, webhookEvent.SessionID, webhookEvent.Metadata)

	return webhookEvent, nil
}
