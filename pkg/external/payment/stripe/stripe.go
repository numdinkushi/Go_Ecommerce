package stripe

import (
	"os"

	"github.com/stripe/stripe-go/v84"
)

type StripeClient struct {
	Client *stripe.Client
}

func NewStripeClient() *StripeClient {
	return &StripeClient{
		Client: stripe.NewClient(os.Getenv("STRIPE_SECRET_KEY")),
	}
}
