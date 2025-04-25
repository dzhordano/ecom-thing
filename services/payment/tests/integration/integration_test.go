package integration

import (
	"context"
	"github.com/dzhordano/ecom-thing/services/payment/internal/application/dto"
	"github.com/dzhordano/ecom-thing/services/payment/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/payment/internal/application/service"
	"github.com/dzhordano/ecom-thing/services/payment/internal/domain"
	"github.com/dzhordano/ecom-thing/services/payment/internal/domain/repository"
	"github.com/dzhordano/ecom-thing/services/payment/internal/infrastructure/repository/pg"
	"github.com/dzhordano/ecom-thing/services/payment/pkg/logger"
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
)

type Suite struct {
	suite.Suite

	db   *pgxpool.Pool
	repo repository.PaymentRepository
	svc  interfaces.PaymentService

	testPayment1 *domain.Payment
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

	testLogger := logger.MustInit(logger.LevelDebug, "payment-test.log", "json", false)

	s.db = pool
	s.repo = pg.NewPaymentRepository(s.db)
	s.svc = service.NewPaymerService(testLogger, s.repo)
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
	s.testPayment1 = &domain.Payment{
		ID:            uuid.New(),
		UserID:        uuid.New(),
		OrderID:       uuid.New(),
		Currency:      domain.USD,
		TotalPrice:    99.99,
		PaymentMethod: domain.PaymentMethodCard,
		Description:   "Test Description",
		RedirectURL:   "https://test.some-url.com/bruh",
		Status:        domain.PaymentPending,
		CreatedAt:     time.Date(2025, time.April, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2025, time.April, 1, 12, 0, 0, 0, time.UTC),
	}

	err := s.repo.Save(context.Background(), s.testPayment1)

	s.NoError(err)
}

func (s *Suite) TearDownTest() {
	err := s.repo.Delete(context.Background(), s.testPayment1.ID.String())

	s.NoError(err)
}

func (s *Suite) Test_CreatePayment() {
	newPayment, err := domain.NewPayment(
		uuid.New(),
		uuid.New(),
		domain.USD.String(),
		99.99,
		domain.PaymentMethodCard.String(),
		"Test Description",
		"https://test.some-url.com/bruh",
		domain.PaymentPending.String(),
	)

	s.NoError(err)

	p, err := s.svc.CreatePayment(context.Background(), dto.CreatePaymentRequest{
		OrderId:       newPayment.OrderID,
		UserId:        newPayment.UserID,
		Currency:      newPayment.Currency.String(),
		TotalPrice:    newPayment.TotalPrice,
		PaymentMethod: newPayment.PaymentMethod.String(),
		Description:   newPayment.Description,
		RedirectURL:   newPayment.RedirectURL,
	})

	s.NoError(err)

	sp, err := s.repo.GetById(context.Background(), p.ID.String(), p.UserID.String())

	s.NoError(err)

	s.Equal(p.ID, sp.ID)
	s.Equal(p.OrderID, sp.OrderID)
	s.Equal(p.UserID, sp.UserID)
	s.Equal(p.Currency, sp.Currency)
	s.Equal(p.TotalPrice, sp.TotalPrice)
	s.Equal(p.PaymentMethod, sp.PaymentMethod)
	s.Equal(p.Description, sp.Description)
	s.Equal(p.Status, sp.Status)
	s.Equal(p.CreatedAt.Unix(), sp.CreatedAt.Unix())
	s.Equal(p.UpdatedAt.Unix(), sp.UpdatedAt.Unix())
}

func (s *Suite) Test_CancelPayment() {
	err := s.svc.CancelPayment(context.Background(), s.testPayment1.ID, s.testPayment1.UserID)

	s.NoError(err)

	p, err := s.repo.GetById(context.Background(), s.testPayment1.ID.String(), s.testPayment1.UserID.String())

	s.NoError(err)

	s.Equal(p.Status, domain.PaymentCancelled)
}

func (s *Suite) Test_ConfirmPayment() {
	err := s.svc.ConfirmPayment(context.Background(), s.testPayment1.ID, s.testPayment1.UserID)

	s.NoError(err)

	p, err := s.repo.GetById(context.Background(), s.testPayment1.ID.String(), s.testPayment1.UserID.String())

	s.NoError(err)

	s.Equal(domain.PaymentCompleted, p.Status)
}

func (s *Suite) Test_RetryPayment() {
	s.testPayment1.SetStatus(domain.PaymentCancelled)

	err := s.repo.Update(context.Background(), s.testPayment1)

	s.NoError(err)

	err = s.svc.RetryPayment(context.Background(), s.testPayment1.ID, s.testPayment1.UserID)

	s.NoError(err)

	p, err := s.repo.GetById(context.Background(), s.testPayment1.ID.String(), s.testPayment1.UserID.String())

	s.NoError(err)

	s.Equal(domain.PaymentPending, p.Status)
}

func (s *Suite) Test_GetPaymentStatus() {
	ps, err := s.svc.GetPaymentStatus(context.Background(), s.testPayment1.ID, s.testPayment1.UserID)

	s.NoError(err)

	s.Equal(s.testPayment1.Status.String(), ps)
}
