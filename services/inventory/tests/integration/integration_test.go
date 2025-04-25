package integration

import (
	"context"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/application/service"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/domain"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/domain/repository"
	"github.com/dzhordano/ecom-thing/services/inventory/internal/infrastructure/repository/pg"
	"github.com/dzhordano/ecom-thing/services/inventory/pkg/logger"
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
	repo repository.ItemRepository
	svc  interfaces.ItemService

	testItem1 *domain.Item
}

func (s *Suite) SetupSuite() {
	// Get current file location
	_, currentFile, _, _ := runtime.Caller(0)

	// Get current dir from current file path
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

	testLogger := logger.MustInit(logger.LevelDebug, "inventory-test.log", "json", false)

	s.db = pool
	s.repo = pg.NewInventoryRepository(s.db)
	s.svc = service.NewItemService(testLogger, s.repo)

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
	s.testItem1 = &domain.Item{
		ProductID:         uuid.New(),
		AvailableQuantity: 10,
		ReservedQuantity:  10,
	}

	err := s.repo.SetItem(context.Background(), s.testItem1.ProductID.String(), s.testItem1.AvailableQuantity, s.testItem1.ReservedQuantity)

	s.NoError(err)
}

func (s *Suite) TearDownTest() {
	// TODO Удалить потом. Сервис не подразумевал удаления. Думал сделать просто крон, который переодично будет удалять записи.
}

func (s *Suite) Test_SetItemWithOp_ADD10() {
	err := s.svc.SetItemWithOp(context.Background(), s.testItem1.ProductID, 10, domain.OperationAdd)

	s.NoError(err)

	item, err := s.repo.GetItem(context.Background(), s.testItem1.ProductID.String())

	s.NoError(err)

	s.Equal(uint64(20), item.AvailableQuantity)
	s.Equal(uint64(10), item.ReservedQuantity)
}

func (s *Suite) Test_SetItemWithOp_SUB10() {
	err := s.svc.SetItemWithOp(context.Background(), s.testItem1.ProductID, 10, domain.OperationSub)

	s.NoError(err)

	item, err := s.repo.GetItem(context.Background(), s.testItem1.ProductID.String())

	s.NoError(err)

	s.Equal(uint64(0), item.AvailableQuantity)
	s.Equal(uint64(10), item.ReservedQuantity)
}

func (s *Suite) Test_SetItemWithOp_LOCK10() {
	err := s.svc.SetItemWithOp(context.Background(), s.testItem1.ProductID, 10, domain.OperationLock)

	s.NoError(err)

	item, err := s.repo.GetItem(context.Background(), s.testItem1.ProductID.String())

	s.NoError(err)

	s.Equal(uint64(0), item.AvailableQuantity)
	s.Equal(uint64(20), item.ReservedQuantity)
}

func (s *Suite) Test_SetItemsWithOp_ADD10() {
	testItem2 := domain.Item{
		ProductID:         uuid.New(),
		AvailableQuantity: 20,
		ReservedQuantity:  10,
	}

	err := s.repo.SetItem(context.Background(), testItem2.ProductID.String(), testItem2.AvailableQuantity, testItem2.ReservedQuantity)

	s.NoError(err)

	err = s.svc.SetItemsWithOp(context.Background(), map[string]uint64{
		s.testItem1.ProductID.String(): 10,
		testItem2.ProductID.String():   10,
	}, domain.OperationAdd)

	s.NoError(err)

	i1, err := s.repo.GetItem(context.Background(), s.testItem1.ProductID.String())

	s.NoError(err)

	i2, err := s.repo.GetItem(context.Background(), testItem2.ProductID.String())

	s.NoError(err)

	s.Equal(uint64(20), i1.AvailableQuantity)
	s.Equal(uint64(10), i1.ReservedQuantity)

	s.Equal(uint64(30), i2.AvailableQuantity)
	s.Equal(uint64(10), i2.ReservedQuantity)
}

func (s *Suite) Test_SetItemsWithOp_SUB10() {
	testItem2 := domain.Item{
		ProductID:         uuid.New(),
		AvailableQuantity: 20,
		ReservedQuantity:  10,
	}

	err := s.repo.SetItem(context.Background(), testItem2.ProductID.String(), testItem2.AvailableQuantity, testItem2.ReservedQuantity)

	s.NoError(err)

	err = s.svc.SetItemsWithOp(context.Background(), map[string]uint64{
		s.testItem1.ProductID.String(): 10,
		testItem2.ProductID.String():   10,
	}, domain.OperationSub)

	s.NoError(err)

	i1, err := s.repo.GetItem(context.Background(), s.testItem1.ProductID.String())

	s.NoError(err)

	i2, err := s.repo.GetItem(context.Background(), testItem2.ProductID.String())

	s.NoError(err)

	s.Equal(uint64(0), i1.AvailableQuantity)
	s.Equal(uint64(10), i1.ReservedQuantity)

	s.Equal(uint64(10), i2.AvailableQuantity)
	s.Equal(uint64(10), i2.ReservedQuantity)
}

func (s *Suite) Test_SetItemsWithOp_SUBLOCKED10() {
	testItem2 := domain.Item{
		ProductID:         uuid.New(),
		AvailableQuantity: 20,
		ReservedQuantity:  10,
	}

	err := s.repo.SetItem(context.Background(), testItem2.ProductID.String(), testItem2.AvailableQuantity, testItem2.ReservedQuantity)

	s.NoError(err)

	err = s.svc.SetItemsWithOp(context.Background(), map[string]uint64{
		s.testItem1.ProductID.String(): 10,
		testItem2.ProductID.String():   10,
	}, domain.OperationSubLocked)

	s.NoError(err)

	i1, err := s.repo.GetItem(context.Background(), s.testItem1.ProductID.String())

	s.NoError(err)

	i2, err := s.repo.GetItem(context.Background(), testItem2.ProductID.String())

	s.NoError(err)

	s.Equal(uint64(10), i1.AvailableQuantity)
	s.Equal(uint64(0), i1.ReservedQuantity)

	s.Equal(uint64(20), i2.AvailableQuantity)
	s.Equal(uint64(0), i2.ReservedQuantity)
}

func (s *Suite) Test_GetItem() {
	item, err := s.svc.GetItem(context.Background(), s.testItem1.ProductID)

	s.NoError(err)

	s.Equal(s.testItem1.ProductID, item.ProductID)
	s.Equal(s.testItem1.AvailableQuantity, item.AvailableQuantity)
	s.Equal(s.testItem1.ReservedQuantity, item.ReservedQuantity)
}

func (s *Suite) Test_IsReservable_TRUE() {
	isReservable, err := s.svc.IsReservable(context.Background(), map[string]uint64{
		s.testItem1.ProductID.String(): 10,
	})

	s.NoError(err)

	s.True(isReservable)
}

func (s *Suite) Test_IsReservable_FALSE() {
	isReservable, err := s.svc.IsReservable(context.Background(), map[string]uint64{
		s.testItem1.ProductID.String(): 11,
	})

	s.NoError(err)

	s.False(isReservable)
}

func (s *Suite) Test_IsReservable_NOTFOUND() {
	isReservable, err := s.svc.IsReservable(context.Background(), map[string]uint64{
		uuid.New().String(): 11,
	})

	s.ErrorIs(err, domain.ErrProductNotFound)

	s.False(isReservable)
}
