package stripe

import (
	"fmt"
	"os"

	stripeSDK "github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/checkout/session"
)

type Service struct {
	secretKey string
	priceID   string
}

func NewService(secretKey, priceID string) *Service {
	stripeSDK.Key = secretKey
	return &Service{
		secretKey: secretKey,
		priceID:   priceID,
	}
}

func (s *Service) CreateCheckoutSession(customerID, email string) (string, error) {
	domain := os.Getenv("BASE_DOMAIN")
	if domain == "" {
		domain = "localhost:8080"
	}

	params := &stripeSDK.CheckoutSessionParams{
		LineItems: []*stripeSDK.CheckoutSessionLineItemParams{
			{
				Price:    stripeSDK.String(s.priceID),
				Quantity: stripeSDK.Int64(1),
			},
		},
		Mode:          stripeSDK.String(string(stripeSDK.CheckoutSessionModeSubscription)),
		SuccessURL:    stripeSDK.String(fmt.Sprintf("https://%s/success?customer_id=%s", domain, customerID)),
		CancelURL:     stripeSDK.String(fmt.Sprintf("https://%s/configure", domain)),
		CustomerEmail: stripeSDK.String(email),
		Metadata: map[string]string{
			"customer_id": customerID,
		},
	}

	result, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("create checkout session: %w", err)
	}

	return result.URL, nil
}
