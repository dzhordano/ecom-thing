package interfaces

import (
	"context"
	"github.com/google/uuid"
)

type ProductService interface {
	GetProductInfo(ctx context.Context, orderId uuid.UUID) (float64, bool, error)
}
