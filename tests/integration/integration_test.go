package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/factory"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/repository/postgres"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	ctx := context.Background()

	pgContainer, err := pgcontainer.Run(ctx,
		"postgres:15",
		pgcontainer.WithDatabase("testdb"),
		pgcontainer.WithUsername("testuser"),
		pgcontainer.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	require.NoError(t, err)

	connString, err := pgContainer.ConnectionString(ctx)
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connString)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS couriers (
			id             BIGSERIAL PRIMARY KEY,
			name           TEXT NOT NULL,
			phone          TEXT NOT NULL UNIQUE,
			status         TEXT NOT NULL DEFAULT 'available',
			transport_type TEXT NOT NULL DEFAULT 'on_foot',
			created_at     TIMESTAMP DEFAULT NOW(),
			updated_at     TIMESTAMP DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS delivery (
			id          BIGSERIAL PRIMARY KEY,
			courier_id  BIGINT NOT NULL,
			order_id    VARCHAR(255) NOT NULL,
			assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deadline    TIMESTAMP NOT NULL
		);
	`)
	require.NoError(t, err)

	cleanup := func() {
		pool.Close()
		_ = pgContainer.Terminate(ctx)
	}

	return pool, cleanup
}

func TestCourierIntegration_CreateAndGet(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

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
	pool, cleanup := setupTestDB(t)
	defer cleanup()

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
	pool, cleanup := setupTestDB(t)
	defer cleanup()

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
	pool, cleanup := setupTestDB(t)
	defer cleanup()

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
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := postgres.NewCourierRepository(pool)
	svc := service.NewCourierService(repo)

	_, err := svc.GetCourier(ctx, 999999)
	assert.Error(t, err)
	assert.ErrorIs(t, err, model.ErrNotFound)
}

func TestDeliveryIntegration_AssignAndUnassign(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

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
	timeFactory := factory.NewDeliveryTimeFactory()
	deliverySvc := service.NewDeliveryService(deliveryRepo, timeFactory)

	assignedCourier, delivery, err := deliverySvc.AssignCourier(ctx, "order-123")
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
	pool, cleanup := setupTestDB(t)
	defer cleanup()

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
	timeFactory := factory.NewDeliveryTimeFactory()
	deliverySvc := service.NewDeliveryService(deliveryRepo, timeFactory)

	courier1, delivery1, err := deliverySvc.AssignCourier(ctx, "order-1")
	require.NoError(t, err)
	assert.Equal(t, "order-1", delivery1.OrderID)

	courier2, delivery2, err := deliverySvc.AssignCourier(ctx, "order-2")
	require.NoError(t, err)
	assert.Equal(t, "order-2", delivery2.OrderID)
	assert.NotEqual(t, courier1.ID, courier2.ID)

	courier3, delivery3, err := deliverySvc.AssignCourier(ctx, "order-3")
	require.NoError(t, err)
	assert.Equal(t, "order-3", delivery3.OrderID)
	assert.NotEqual(t, courier1.ID, courier3.ID)
	assert.NotEqual(t, courier2.ID, courier3.ID)
}

func TestDeliveryIntegration_NoAvailableCouriers(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

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
	timeFactory := factory.NewDeliveryTimeFactory()
	deliverySvc := service.NewDeliveryService(deliveryRepo, timeFactory)

	_, _, err = deliverySvc.AssignCourier(ctx, "order-1")
	require.NoError(t, err)

	_, _, err = deliverySvc.AssignCourier(ctx, "order-2")
	assert.Error(t, err)
	assert.ErrorIs(t, err, model.ErrNoAvailableCouriers)
}

func TestDeliveryIntegration_UnassignNonExistent(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	deliveryRepo := postgres.NewDeliveryRepository(pool)
	timeFactory := factory.NewDeliveryTimeFactory()
	deliverySvc := service.NewDeliveryService(deliveryRepo, timeFactory)

	_, err := deliverySvc.UnassignCourier(ctx, "non-existent-order")
	assert.Error(t, err)
	assert.ErrorIs(t, err, model.ErrDeliveryNotFound)
}

func TestDeliveryIntegration_ExpiredDeliveries(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

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
	timeFactory := factory.NewDeliveryTimeFactory()
	deliverySvc := service.NewDeliveryService(deliveryRepo, timeFactory)

	assignedCourier, delivery, err := deliverySvc.AssignCourier(ctx, "order-test")
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
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	courierRepo := postgres.NewCourierRepository(pool)
	courierSvc := service.NewCourierService(courierRepo)

	deliveryRepo := postgres.NewDeliveryRepository(pool)
	timeFactory := factory.NewDeliveryTimeFactory()
	deliverySvc := service.NewDeliveryService(deliveryRepo, timeFactory)

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
		_, delivery, err := deliverySvc.AssignCourier(ctx, fmt.Sprintf("order-%d", i+1))
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
