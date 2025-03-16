package billing

import (
	"context"
	"errors"
	"log"
	"math/rand/v2"
	"time"
)

type StubBilling struct {
}

func NewStubBilling() Billing {
	return &StubBilling{}
}

func (s *StubBilling) CreatePayment(ctx context.Context, currency, amount, redirectURL, paymentData string) error {
	log.Printf("creating payment with currency: %s, amount: %s, redirectURL: %s, paymentData: %s", currency, amount, redirectURL, paymentData)
	time.Sleep(10 * time.Second)

	if rand.IntN(50) > 40 {
		return errors.New("payment failed") // experimental
	}

	log.Printf("success creating payment with currency: %s, amount: %s, redirectURL: %s, paymentData: %s", currency, amount, redirectURL, paymentData)

	return nil
}

func (s *StubBilling) CancelPayment(ctx context.Context, paymentId string) error {
	log.Printf("canceling payment with paymentId: %s", paymentId)
	return nil
}
