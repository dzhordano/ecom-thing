package integration

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/dzhordano/ecom-thing/services/product/internal/application/interfaces"
	"github.com/dzhordano/ecom-thing/services/product/internal/application/service"
	"github.com/dzhordano/ecom-thing/services/product/internal/domain"
	"github.com/dzhordano/ecom-thing/services/product/internal/domain/repository"
	"github.com/dzhordano/ecom-thing/services/product/internal/infrastructure/repository/pg"
	"github.com/dzhordano/ecom-thing/services/product/pkg/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

type IntegrationSuite struct {
	suite.Suite

	db             *pgxpool.Pool
	productRepo    repository.ProductRepository
	productService interfaces.ProductService

	testProduct1 domain.Product
	testProduct2 domain.Product
}

func (s *IntegrationSuite) SetupSuite() {
	_, currentFile, _, _ := runtime.Caller(0)

	currDir := filepath.Dir(currentFile)

	projectDir := filepath.Join(currDir, "..", "..")

	envPath := filepath.Join(projectDir, ".env")
	migrationsPath := filepath.Join(projectDir, "migrations")

	if err := godotenv.Load(envPath); err != nil {
		log.Fatalf("integration_suite: error loading .env file: %v", err)
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

	nilLogger := logger.MustInit("warn", "product-test", "json", false)

	s.db = pool
	s.productRepo = pg.NewProductRepository(s.db)
	s.productService = service.NewProductService(nilLogger, s.productRepo)

	s.SeedDatabase()
}

func (s *IntegrationSuite) SeedDatabase() {
	s.testProduct1 = domain.Product{
		ID:        uuid.New(),
		Name:      "Dummy1",
		Desc:      "Dummy1",
		Category:  "Dummy1",
		Price:     11.11,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.testProduct2 = domain.Product{
		ID:        uuid.New(),
		Name:      "Dummy2",
		Desc:      "Dummy2",
		Category:  "Dummy2",
		Price:     10.10,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.productRepo.Save(context.Background(), &s.testProduct1); err != nil {
		s.T().Fatalf("failed to seed database: %v", err)
	}

	if err := s.productRepo.Save(context.Background(), &s.testProduct2); err != nil {
		s.T().Fatalf("failed to seed database: %v", err)
	}
}

func (s *IntegrationSuite) TearDownSuite() {
	s.db.Close()
}

func (s *IntegrationSuite) SetupTest() {
	// TODO
}

func (s *IntegrationSuite) TearDownTest() {
	// TODO
}

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	suite.Run(t, new(IntegrationSuite))
}

func (s *IntegrationSuite) TestA_CreateProduct() {
	resp, err := s.productService.CreateProduct(
		context.Background(),
		"TestName",
		"TestDesc",
		"TestCat",
		10.10,
	)

	if s.Assert().NoError(err) {
		s.Assert().NotNil(resp)

		err = resp.Validate()
		s.Assert().NoError(err)
	}
}

func (s *IntegrationSuite) TestB_GetProduct() {
	resp, err := s.productService.GetById(
		context.Background(),
		s.testProduct1.ID,
	)

	if s.Assert().NoError(err) {
		s.Assert().NotNil(resp)

		err = resp.Validate()
		s.Assert().NoError(err)
	}
}

// Now THIS is NOT funny...
func ptrVal[T any](val T) *T {
	return &val
}

func (s *IntegrationSuite) TestC_SearchProducts() {
	respDummyQuery, err := s.productService.SearchProducts(
		context.Background(),
		map[string]any{
			"query": ptrVal("Dummy"),
		})

	if s.Assert().NoError(err) {
		s.Assert().Len(respDummyQuery, 2)
	}

	resp1Query, err := s.productService.SearchProducts(
		context.Background(),
		map[string]any{
			"query": ptrVal("1"),
		})

	if s.Assert().NoError(err) {
		s.Assert().Len(resp1Query, 1)
	}

	respCatQuery, err := s.productService.SearchProducts(
		context.Background(),
		map[string]any{
			"category": ptrVal("Dummy2"),
		})

	if s.Assert().NoError(err) {
		s.Assert().Len(respCatQuery, 1)
	}

	respMinPriceQuery, err := s.productService.SearchProducts(
		context.Background(),
		map[string]any{
			"minPrice": ptrVal(10.09),
		})

	if s.Assert().NoError(err) {
		s.Assert().Len(respMinPriceQuery, 3)
	}

	respMaxPriceQuery, err := s.productService.SearchProducts(
		context.Background(),
		map[string]any{
			"maxPrice": ptrVal(10.11),
		})

	if s.Assert().NoError(err) {
		s.Assert().Len(respMaxPriceQuery, 2)
	}

	respLimitQuery, err := s.productService.SearchProducts(
		context.Background(),
		map[string]any{
			"limit": ptrVal(uint64(1)),
		})

	if s.Assert().NoError(err) {
		s.Assert().Len(respLimitQuery, 1)
	}
}

func (s *IntegrationSuite) TestD_UpdateProduct() {
	resp, err := s.productService.UpdateProduct(
		context.Background(),
		s.testProduct1.ID,
		"NewDummy1",
		"NewDummy1",
		"NewDummy1",
		true,
		12.12,
	)

	if s.Assert().NoError(err) {
		s.Assert().NotNil(resp)

		err = resp.Validate()
		s.Assert().NoError(err)
	}

	repoProd, err := s.productRepo.GetById(context.Background(), s.testProduct1.ID)
	if s.Assert().NoError(err) {
		s.Assert().Equal("NewDummy1", repoProd.Name)
		s.Assert().Equal("NewDummy1", repoProd.Desc)
		s.Assert().Equal("NewDummy1", repoProd.Category)
		s.Assert().Equal(true, repoProd.IsActive)
		s.Assert().Equal(12.12, repoProd.Price)
	}
}

func (s *IntegrationSuite) TestE_DeactivateProduct() {
	resp, err := s.productService.DeactivateProduct(context.Background(), s.testProduct1.ID)

	if s.Assert().NoError(err) {
		s.Assert().NotNil(resp)

		err = resp.Validate()
		s.Assert().NoError(err)
	}
}
