package billing

import (
	"context"
	"log"
	"time"

	"github.com/dzhordano/ecom-thing/services/payment/internal/domain"
)

type StubBilling struct {
}

func NewStubBilling() domain.Billing {
	return &StubBilling{}
}

func (s *StubBilling) NewPayment(ctx context.Context, currency string, totalPrice float64, paymentData string) error {
	const op = "billing.StubBilling.NewPayment"

	log.Printf("creating payment with currency: %s, totalPrice: %f, paymentData: %s", currency, totalPrice, paymentData)
	time.Sleep(5 * time.Second)

	// if rand.IntN(50) > 0 {
	// 	return fmt.Errorf("%s: %w", op, domain.ErrPaymentFailed) // FIXME oh no oh no it failed...
	// }

	log.Printf("success creating payment with currency: %s, totalPrice: %f, paymentData: %s", currency, totalPrice, paymentData)

	return nil
}

func (s *StubBilling) CancelPayment(ctx context.Context, paymentId string) error {
	log.Printf("canceling payment with paymentId: %s", paymentId)
	return nil
}
