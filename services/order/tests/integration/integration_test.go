package integration

import (
	"context"
	"github.com/dzhordano/ecom-thing/services/order/internal/application/dto"
	"github.com/dzhordano/ecom-thing/services/order/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/order/internal/application/service"
	"github.com/dzhordano/ecom-thing/services/order/internal/domain"
	"github.com/dzhordano/ecom-thing/services/order/internal/domain/repository"
	"github.com/dzhordano/ecom-thing/services/order/internal/infrastructure/repository/pg"
	"github.com/dzhordano/ecom-thing/services/order/pkg/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
	"unsafe"
)

type stubInventoryService struct {
	RetReservable bool
	RetErr        error
}

func (s *stubInventoryService) IsReservable(ctx context.Context, items map[string]uint64) (bool, error) {
	return s.RetReservable, s.RetErr
}

type stubProductService struct {
	RetPrice float64
	RetValid bool
	RetErr   error
}

func (s *stubProductService) GetProductInfo(ctx context.Context, orderId uuid.UUID) (float64, bool, error) {
	return s.RetPrice, s.RetValid, s.RetErr
}

type Suite struct {
	suite.Suite

	db       *pgxpool.Pool
	repo     repository.OrderRepository
	orderSvc interfaces.OrderService
	invSvc   unsafe.Pointer
	prodSvc  unsafe.Pointer

	testOrder *domain.Order
}

func (s *Suite) SetupSuite() {
	// Get current file location
	_, currentFile, _, _ := runtime.Caller(0)

	// Get current dir from the current file path
	currDir := filepath.Dir(currentFile)

	// Go two folders up
	projectDir := filepath.Join(currDir, "..", "..")

	// Specify .env file
	envPath := filepath.Join(projectDir, ".env")
	// Specify folder containing migrations
	migrationsPath := filepath.Join(projectDir, "migrations")

	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("suite: error loading .env file: %v", err)
	}

	dsn := os.Getenv("PG_TEST_URL")

	if dsn == "" {
		log.Fatal("PG_TEST_URL not specified in .env")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	timeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(timeout); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	m, err := migrate.New("file://"+migrationsPath, dsn)
	if err != nil {
		log.Fatalf("failed to create migrate instance: %v", err)
	}

	if err = m.Up(); err != nil {
		panic(err)
	}

	testLogger := logger.MustInit(logger.LevelDebug, "order-test.log", "json", false)

	s.db = pool
	s.repo = pg.NewOrderRepository(s.db)

	s.orderSvc = service.NewOrderService(
		testLogger,
		&stubProductService{
			RetPrice: 99.99,
			RetValid: true,
			RetErr:   nil,
		},
		&stubInventoryService{
			RetReservable: true,
			RetErr:        nil,
		},
		s.repo)
}

func (s *Suite) TearDownSuite() {
	s.db.Close()
}

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	suite.Run(t, new(Suite))
}

func (s *Suite) SetupTest() {

	s.testOrder = &domain.Order{
		ID:              uuid.New(),
		UserID:          uuid.New(),
		Description:     "Test Description",
		Status:          domain.OrderPending,
		Currency:        domain.USD,
		TotalPrice:      99.99,
		PaymentMethod:   domain.Cash,
		DeliveryMethod:  domain.Pickup,
		DeliveryAddress: "Test Address",
		DeliveryDate:    time.Now().Add(time.Hour),
		Items: []domain.Item{
			{
				ProductID: uuid.New(),
				Quantity:  1,
			},
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	err := s.repo.Save(context.Background(), s.testOrder)
	s.NoError(err)
}

func (s *Suite) TearDownTest() {
	err := s.repo.Delete(context.Background(), s.testOrder.ID.String())
	s.NoError(err)
}

func (s *Suite) Test_CreateOrder() {
	info := dto.CreateOrderRequest{
		Description:     "TestDescription",
		Currency:        domain.USD.String(),
		Coupon:          "",
		PaymentMethod:   domain.Cash.String(),
		DeliveryMethod:  domain.Pickup.String(),
		DeliveryAddress: "TestAddress",
		DeliveryDate:    time.Now().Add(time.Hour).UTC(),
		Items:           []domain.Item{{ProductID: uuid.New(), Quantity: 1}},
	}

	o, err := s.orderSvc.CreateOrder(context.Background(), info)
	s.NoError(err)

	ro, err := s.repo.GetById(context.Background(), o.ID.String())
	s.NoError(err)

	s.Equal(o.ID.String(), ro.ID.String())
	s.Equal(o.UserID.String(), ro.UserID.String())
	s.Equal(o.Description, ro.Description)
	s.Equal(o.Status.String(), ro.Status.String())
	s.Equal(o.Currency.String(), ro.Currency.String())
	s.Equal(o.TotalPrice, ro.TotalPrice)
	s.Equal(o.PaymentMethod.String(), ro.PaymentMethod.String())
	s.Equal(o.DeliveryMethod.String(), ro.DeliveryMethod.String())
	s.Equal(o.DeliveryAddress, ro.DeliveryAddress)
	s.Equal(o.DeliveryDate.Unix(), ro.DeliveryDate.Unix())
	s.Equal(o.Items, ro.Items)
}

func (s *Suite) Test_CancelOrder() {
	err := s.orderSvc.CancelOrder(context.Background(), s.testOrder.ID)
	s.NoError(err)

	ro, err := s.repo.GetById(context.Background(), s.testOrder.ID.String())
	s.NoError(err)
	s.Equal(domain.OrderCancelled, ro.Status)
}

func (s *Suite) Test_CompleteOrder() {
	err := s.orderSvc.CompleteOrder(context.Background(), s.testOrder.ID)
	s.NoError(err)

	ro, err := s.repo.GetById(context.Background(), s.testOrder.ID.String())
	s.NoError(err)
	s.Equal(domain.OrderCompleted, ro.Status)
}

func (s *Suite) Test_UpdateOrder() {
	s.testOrder.Description = "Updated Description"
	s.testOrder.DeliveryAddress = "Updated Address"
	s.testOrder.DeliveryDate = time.Now().UTC().Add(time.Hour)

	o, err := s.orderSvc.UpdateOrder(context.Background(), dto.UpdateOrderRequest{
		OrderID:         s.testOrder.ID,
		Description:     &s.testOrder.Description,
		Status:          ToPtr(s.testOrder.Status.String()),
		TotalPrice:      &s.testOrder.TotalPrice,
		PaymentMethod:   ToPtr(s.testOrder.PaymentMethod.String()),
		DeliveryMethod:  ToPtr(s.testOrder.DeliveryMethod.String()),
		DeliveryAddress: &s.testOrder.DeliveryAddress,
		DeliveryDate:    s.testOrder.DeliveryDate,
		Items:           s.testOrder.Items,
	})
	s.NoError(err)

	ro, err := s.repo.GetById(context.Background(), o.ID.String())
	s.NoError(err)

	s.Equal(s.testOrder.ID.String(), ro.ID.String())
	s.Equal(s.testOrder.Description, ro.Description)
	s.Equal(s.testOrder.DeliveryAddress, ro.DeliveryAddress)
	s.Equal(s.testOrder.DeliveryDate.Unix(), ro.DeliveryDate.Unix())
	s.Equal(s.testOrder.Status.String(), ro.Status.String())
	s.Equal(s.testOrder.Currency.String(), ro.Currency.String())
	s.Equal(s.testOrder.TotalPrice, ro.TotalPrice)
	s.Equal(s.testOrder.PaymentMethod.String(), ro.PaymentMethod.String())
	s.Equal(s.testOrder.DeliveryMethod.String(), ro.DeliveryMethod.String())
	s.Equal(s.testOrder.Items, ro.Items)
	s.Equal(s.testOrder.CreatedAt.Unix(), ro.CreatedAt.Unix())
	s.Equal(s.testOrder.UpdatedAt.Unix(), ro.UpdatedAt.Unix())
}

func (s *Suite) Test_DeleteOrder() {
	o := &domain.Order{
		ID:              uuid.New(),
		UserID:          uuid.UUID{},
		Description:     "",
		Status:          "",
		Currency:        "",
		TotalPrice:      0,
		PaymentMethod:   "",
		DeliveryMethod:  "",
		DeliveryAddress: "",
		DeliveryDate:    time.Time{},
		Items:           nil,
		CreatedAt:       time.Time{},
		UpdatedAt:       time.Time{},
	}

	err := s.repo.Save(context.TODO(), o)
	s.NoError(err)

	err = s.orderSvc.DeleteOrder(context.TODO(), o.ID)
	s.NoError(err)

	_, err = s.orderSvc.GetById(context.TODO(), o.ID)

	s.ErrorIs(err, domain.ErrOrderNotFound)
}

func (s *Suite) Test_GetById() {
	o, err := s.orderSvc.GetById(context.TODO(), s.testOrder.ID)

	s.NoError(err)
	s.Equal(s.testOrder.ID.String(), o.ID.String())
	s.Equal(s.testOrder.Description, o.Description)
	s.Equal(s.testOrder.DeliveryAddress, o.DeliveryAddress)
	s.Equal(s.testOrder.DeliveryDate.Unix(), o.DeliveryDate.Unix())
	s.Equal(s.testOrder.Status.String(), o.Status.String())
	s.Equal(s.testOrder.Currency.String(), o.Currency.String())
	s.Equal(s.testOrder.TotalPrice, o.TotalPrice)
	s.Equal(s.testOrder.PaymentMethod.String(), o.PaymentMethod.String())
	s.Equal(s.testOrder.DeliveryMethod.String(), o.DeliveryMethod.String())
	s.Equal(s.testOrder.Items, o.Items)
	s.Equal(s.testOrder.CreatedAt.Unix(), o.CreatedAt.Unix())
	s.Equal(s.testOrder.UpdatedAt.Unix(), o.UpdatedAt.Unix())
}

// TODO Когда добавлю логику получения айди пользователя из контекста
func (s *Suite) Test_ListByUser() {}

func (s *Suite) Test_SearchOrders() {

}

func ToPtr[T any](val T) *T {
	return &val
}
