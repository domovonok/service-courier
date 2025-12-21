package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/factory"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/logger"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/repository/postgres"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/service"
	"github.com/Avito-courses/course-go-avito-domovonok/tests/testdb"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDB *testdb.TestDatabase

func TestMain(m *testing.M) {
	ctx := context.Background()

	db, err := testdb.SetupTestDatabase(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to setup test database: %v", err))
	}
	testDB = db

	code := m.Run()

	_ = testDB.Teardown(ctx)

	os.Exit(code)
}

func cleanupTestData(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	err := testdb.CleanupTables(ctx, pool)
	require.NoError(t, err)
}

func TestCourierIntegration_CreateAndGet(t *testing.T) {
	pool := testDB.Pool
	defer cleanupTestData(t, pool)

	ctx := context.Background()
	repo := postgres.NewCourierRepository(pool)
	svc := service.NewCourierService(repo)

	courier := &model.Courier{
		Name:          "John Doe",
		Phone:         "+1234567890",
		Status:        "available",
		TransportType: "on_foot",
	}

	created, err := svc.CreateCourier(ctx, courier)
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Greater(t, created.ID, int64(0))
	assert.Equal(t, "John Doe", created.Name)

	retrieved, err := svc.GetCourier(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.Name, retrieved.Name)
	assert.Equal(t, created.Phone, retrieved.Phone)
	assert.Equal(t, created.Status, retrieved.Status)
	assert.Equal(t, created.TransportType, retrieved.TransportType)
}

func TestCourierIntegration_List(t *testing.T) {
	pool := testDB.Pool
	defer cleanupTestData(t, pool)

	ctx := context.Background()
	repo := postgres.NewCourierRepository(pool)
	svc := service.NewCourierService(repo)

	couriers := []*model.Courier{
		{Name: "John Doe", Phone: "+1234567890", Status: "available", TransportType: "on_foot"},
		{Name: "Jane Smith", Phone: "+9876543210", Status: "busy", TransportType: "car"},
		{Name: "Bob Brown", Phone: "+5555555555", Status: "available", TransportType: "scooter"},
	}

	for _, c := range couriers {
		_, err := svc.CreateCourier(ctx, c)
		require.NoError(t, err)
	}

	list, err := svc.ListCouriers(ctx)
	require.NoError(t, err)
	assert.Len(t, list, 3)
}

func TestCourierIntegration_Update(t *testing.T) {
	pool := testDB.Pool
	defer cleanupTestData(t, pool)

	ctx := context.Background()
	repo := postgres.NewCourierRepository(pool)
	svc := service.NewCourierService(repo)

	courier := &model.Courier{
		Name:          "John Doe",
		Phone:         "+1234567890",
		Status:        "available",
		TransportType: "on_foot",
	}

	created, err := svc.CreateCourier(ctx, courier)
	require.NoError(t, err)

	created.Name = "John Doe Updated"
	created.Status = "busy"
	created.TransportType = "car"

	updated, err := svc.UpdateCourier(ctx, created)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "John Doe Updated", updated.Name)
	assert.Equal(t, "busy", updated.Status)
	assert.Equal(t, "car", updated.TransportType)

	retrieved, err := svc.GetCourier(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "John Doe Updated", retrieved.Name)
	assert.Equal(t, "busy", retrieved.Status)
}

func TestCourierIntegration_CreateDuplicate(t *testing.T) {
	pool := testDB.Pool
	defer cleanupTestData(t, pool)

	ctx := context.Background()
	repo := postgres.NewCourierRepository(pool)
	svc := service.NewCourierService(repo)

	courier := &model.Courier{
		Name:          "John Doe",
		Phone:         "+1234567890",
		Status:        "available",
		TransportType: "on_foot",
	}

	_, err := svc.CreateCourier(ctx, courier)
	require.NoError(t, err)

	duplicate := &model.Courier{
		Name:          "Jane Doe",
		Phone:         "+1234567890",
		Status:        "available",
		TransportType: "car",
	}

	_, err = svc.CreateCourier(ctx, duplicate)
	assert.Error(t, err)
	assert.ErrorIs(t, err, model.ErrConflict)
}

func TestCourierIntegration_GetNonExistent(t *testing.T) {
	pool := testDB.Pool
	defer cleanupTestData(t, pool)

	ctx := context.Background()
	repo := postgres.NewCourierRepository(pool)
	svc := service.NewCourierService(repo)

	_, err := svc.GetCourier(ctx, 999999)
	assert.Error(t, err)
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestDeliveryIntegration_AssignAndUnassign(t *testing.T) {
	pool := testDB.Pool
	defer cleanupTestData(t, pool)

	ctx := context.Background()
	courierRepo := postgres.NewCourierRepository(pool)
	courierSvc := service.NewCourierService(courierRepo)

	courier := &model.Courier{
		Name:          "John Doe",
		Phone:         "+1234567890",
		Status:        "available",
		TransportType: "car",
	}

	created, err := courierSvc.CreateCourier(ctx, courier)
	require.NoError(t, err)

	deliveryRepo := postgres.NewDeliveryRepository(pool)
	calculatorFactory := factory.NewDeliveryTimeCalculatorFactory()
	txManager := postgres.NewTransactionManager(pool)
	log := logger.NopLogger{}
	deliverySvc := service.NewDeliveryService(deliveryRepo, courierRepo, calculatorFactory, txManager, log)

	assignedCourier, err, delivery := deliverySvc.AssignCourier(ctx, "order-123")
	require.NoError(t, err)
	require.NotNil(t, assignedCourier)
	require.NotNil(t, delivery)
	assert.Equal(t, created.ID, assignedCourier.ID)
	assert.Equal(t, "order-123", delivery.OrderID)
	assert.Greater(t, delivery.ID, int64(0))

	retrievedCourier, err := courierSvc.GetCourier(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "busy", retrievedCourier.Status)

	courierID, err := deliverySvc.UnassignCourier(ctx, "order-123")
	require.NoError(t, err)
	assert.Equal(t, created.ID, courierID)

	retrievedCourier, err = courierSvc.GetCourier(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "available", retrievedCourier.Status)
}

func TestDeliveryIntegration_AssignMultipleCouriers(t *testing.T) {
	pool := testDB.Pool
	defer cleanupTestData(t, pool)

	ctx := context.Background()
	courierRepo := postgres.NewCourierRepository(pool)
	courierSvc := service.NewCourierService(courierRepo)

	for i := 1; i <= 3; i++ {
		courier := &model.Courier{
			Name:          fmt.Sprintf("Courier %d", i),
			Phone:         fmt.Sprintf("+123456789%d", i),
			Status:        "available",
			TransportType: "on_foot",
		}
		_, err := courierSvc.CreateCourier(ctx, courier)
		require.NoError(t, err)
	}

	deliveryRepo := postgres.NewDeliveryRepository(pool)
	calculatorFactory := factory.NewDeliveryTimeCalculatorFactory()
	txManager := postgres.NewTransactionManager(pool)
	log := logger.NopLogger{}
	deliverySvc := service.NewDeliveryService(deliveryRepo, courierRepo, calculatorFactory, txManager, log)

	courier1, err, delivery1 := deliverySvc.AssignCourier(ctx, "order-1")
	require.NoError(t, err)
	assert.Equal(t, "order-1", delivery1.OrderID)

	courier2, err, delivery2 := deliverySvc.AssignCourier(ctx, "order-2")
	require.NoError(t, err)
	assert.Equal(t, "order-2", delivery2.OrderID)
	assert.NotEqual(t, courier1.ID, courier2.ID)

	courier3, err, delivery3 := deliverySvc.AssignCourier(ctx, "order-3")
	require.NoError(t, err)
	assert.Equal(t, "order-3", delivery3.OrderID)
	assert.NotEqual(t, courier1.ID, courier3.ID)
	assert.NotEqual(t, courier2.ID, courier3.ID)
}

func TestDeliveryIntegration_NoAvailableCouriers(t *testing.T) {
	pool := testDB.Pool
	defer cleanupTestData(t, pool)

	ctx := context.Background()
	courierRepo := postgres.NewCourierRepository(pool)
	courierSvc := service.NewCourierService(courierRepo)

	courier := &model.Courier{
		Name:          "John Doe",
		Phone:         "+1234567890",
		Status:        "available",
		TransportType: "car",
	}
	_, err := courierSvc.CreateCourier(ctx, courier)
	require.NoError(t, err)

	deliveryRepo := postgres.NewDeliveryRepository(pool)
	calculatorFactory := factory.NewDeliveryTimeCalculatorFactory()
	txManager := postgres.NewTransactionManager(pool)
	log := logger.NopLogger{}
	deliverySvc := service.NewDeliveryService(deliveryRepo, courierRepo, calculatorFactory, txManager, log)

	_, err, _ = deliverySvc.AssignCourier(ctx, "order-1")
	require.NoError(t, err)

	_, err, _ = deliverySvc.AssignCourier(ctx, "order-2")
	assert.Error(t, err)
	assert.ErrorIs(t, err, model.ErrNoAvailableCouriers)
}

func TestDeliveryIntegration_UnassignNonExistent(t *testing.T) {
	pool := testDB.Pool
	defer cleanupTestData(t, pool)

	ctx := context.Background()
	courierRepo := postgres.NewCourierRepository(pool)
	deliveryRepo := postgres.NewDeliveryRepository(pool)
	calculatorFactory := factory.NewDeliveryTimeCalculatorFactory()
	txManager := postgres.NewTransactionManager(pool)
	log := logger.NopLogger{}
	deliverySvc := service.NewDeliveryService(deliveryRepo, courierRepo, calculatorFactory, txManager, log)

	_, err := deliverySvc.UnassignCourier(ctx, "non-existent-order")
	assert.Error(t, err)
	assert.ErrorIs(t, err, model.ErrDeliveryNotFound)
}

func TestDeliveryIntegration_ExpiredDeliveries(t *testing.T) {
	pool := testDB.Pool
	defer cleanupTestData(t, pool)

	ctx := context.Background()
	courierRepo := postgres.NewCourierRepository(pool)
	courierSvc := service.NewCourierService(courierRepo)

	courier := &model.Courier{
		Name:          "John Doe",
		Phone:         "+1234567890",
		Status:        "available",
		TransportType: "car",
	}
	created, err := courierSvc.CreateCourier(ctx, courier)
	require.NoError(t, err)

	deliveryRepo := postgres.NewDeliveryRepository(pool)
	calculatorFactory := factory.NewDeliveryTimeCalculatorFactory()
	txManager := postgres.NewTransactionManager(pool)
	log := logger.NopLogger{}
	deliverySvc := service.NewDeliveryService(deliveryRepo, courierRepo, calculatorFactory, txManager, log)

	assignedCourier, err, delivery := deliverySvc.AssignCourier(ctx, "order-test")
	require.NoError(t, err)
	require.NotNil(t, assignedCourier)
	require.NotNil(t, delivery)

	retrievedCourier, err := courierSvc.GetCourier(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "busy", retrievedCourier.Status)

	_, err = pool.Exec(ctx, "UPDATE delivery SET deadline = CURRENT_TIMESTAMP - INTERVAL '1 hour' WHERE order_id = 'order-test'")
	require.NoError(t, err)

	err = deliverySvc.CheckExpiredDeliveries(ctx)
	require.NoError(t, err)

	retrievedCourier, err = courierSvc.GetCourier(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "available", retrievedCourier.Status)
}

func TestDeliveryIntegration_DeadlineCalculation(t *testing.T) {
	pool := testDB.Pool
	defer cleanupTestData(t, pool)

	ctx := context.Background()
	courierRepo := postgres.NewCourierRepository(pool)
	courierSvc := service.NewCourierService(courierRepo)

	deliveryRepo := postgres.NewDeliveryRepository(pool)
	calculatorFactory := factory.NewDeliveryTimeCalculatorFactory()
	txManager := postgres.NewTransactionManager(pool)
	log := logger.NopLogger{}
	deliverySvc := service.NewDeliveryService(deliveryRepo, courierRepo, calculatorFactory, txManager, log)

	transportTypes := []struct {
		transportType    string
		expectedDuration time.Duration
	}{
		{"car", 5 * time.Minute},
		{"scooter", 15 * time.Minute},
		{"on_foot", 30 * time.Minute},
	}

	for i, tt := range transportTypes {
		courier := &model.Courier{
			Name:          fmt.Sprintf("Courier %d", i+1),
			Phone:         fmt.Sprintf("+123456789%d", i+1),
			Status:        "available",
			TransportType: tt.transportType,
		}
		_, err := courierSvc.CreateCourier(ctx, courier)
		require.NoError(t, err)

		beforeAssign := time.Now()
		_, err, delivery := deliverySvc.AssignCourier(ctx, fmt.Sprintf("order-%d", i+1))
		require.NoError(t, err)

		expectedDeadline := beforeAssign.Add(tt.expectedDuration)
		actualDuration := delivery.Deadline.Sub(delivery.AssignedAt)

		assert.InDelta(t, tt.expectedDuration.Seconds(), actualDuration.Seconds(), 1.0,
			"Deadline calculation incorrect for %s", tt.transportType)
		assert.True(t, delivery.Deadline.After(expectedDeadline.Add(-2*time.Second)),
			"Deadline too early for %s", tt.transportType)
		assert.True(t, delivery.Deadline.Before(expectedDeadline.Add(2*time.Second)),
			"Deadline too late for %s", tt.transportType)
	}
}
