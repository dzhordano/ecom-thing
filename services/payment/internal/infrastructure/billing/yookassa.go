package billing

import (
	"context"
	"time"

	"github.com/dzhordano/ecom-thing/services/payment/internal/domain"
	yk "github.com/rvinnie/yookassa-sdk-go/yookassa"
	yoocommon "github.com/rvinnie/yookassa-sdk-go/yookassa/common"
	yoopayment "github.com/rvinnie/yookassa-sdk-go/yookassa/payment"
)

type Client struct {
	ph *yk.PaymentHandler
}

// FIXME Из окружения
func NewClient(accountId, secretKey string) Billing {
	yclient := yk.NewClient(
		accountId,
		secretKey,
	)

	ph := yk.NewPaymentHandler(yclient)

	return &Client{
		ph: ph,
	}
}

// CreatePayment implements Billing.
//
// Is blocking so consider running in goroutine.
func (c *Client) CreatePayment(ctx context.Context, currency, amount, redirectURL, paymentData string) error {
	p, err := c.ph.CreatePayment(&yoopayment.Payment{
		Amount: &yoocommon.Amount{
			Value:    currency, // Извне тоже
			Currency: amount,   // Извне тоже
		},
		PaymentMethod: yoopayment.PaymentMethodType("bank_card"),
		Confirmation: yoopayment.Redirect{
			Type:      "redirect",
			ReturnURL: redirectURL, // Тут извне получение (сделать HTML документ этаки)
		},
		Description: paymentData,
	})
	if err != nil {
		return err
	}

	for {
		time.Sleep(30 * time.Second)
		// FIXME Правильно ли?
		if ctx.Err() != nil {
			return ctx.Err()
		}

		crpt, err := c.ph.CapturePayment(p)
		if err != nil {
			return err
		}

		switch crpt.Status {
		case "waiting_for_capture":
			continue
		case "pending":
			// ???
			return nil
		case "succeeded":
			// Something else?
			return nil
		case "canceled":
			return domain.ErrPaymentCancelled
		}
	}
}

// CancelPayment implements Billing.
func (c *Client) CancelPayment(ctx context.Context, paymentId string) error {
	_, err := c.ph.CancelPayment(paymentId)
	if err != nil {
		// FIXME Failed to cancel
		return err
	}

	return nil
}
