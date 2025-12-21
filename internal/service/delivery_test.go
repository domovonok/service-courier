package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/logger"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/service"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -destination=mocks/delivery_repository_mock.go -package=mocks github.com/Avito-courses/course-go-avito-domovonok/internal/service deliveryRepository
//go:generate mockgen -destination=mocks/courier_repository_mock.go -package=mocks github.com/Avito-courses/course-go-avito-domovonok/internal/service courierRepository
//go:generate mockgen -destination=mocks/delivery_time_calculator_factory_mock.go -package=mocks github.com/Avito-courses/course-go-avito-domovonok/internal/factory DeliveryTimeCalculatorFactory
//go:generate mockgen -destination=mocks/delivery_time_calculator_mock.go -package=mocks github.com/Avito-courses/course-go-avito-domovonok/internal/factory DeliveryTimeCalculator
//go:generate mockgen -destination=mocks/transaction_manager_mock.go -package=mocks github.com/Avito-courses/course-go-avito-domovonok/internal/transaction Manager

func TestDeliveryService_AssignCourier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		orderID     string
		mockSetup   func(*mocks.MockdeliveryRepository, *mocks.MockcourierRepository, *mocks.MockDeliveryTimeCalculatorFactory, *mocks.MockDeliveryTimeCalculator, *mocks.MockManager)
		expectedErr error
	}{
		{
			name:    "successful assignment",
			orderID: "order-123",
			mockSetup: func(repo *mocks.MockdeliveryRepository, courierRepo *mocks.MockcourierRepository, factory *mocks.MockDeliveryTimeCalculatorFactory, calc *mocks.MockDeliveryTimeCalculator, txMgr *mocks.MockManager) {
				courier := &model.Courier{
					ID:            1,
					Name:          "John Doe",
					Phone:         "+1234567890",
					Status:        "available",
					TransportType: "car",
				}

				txMgr.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						repo.EXPECT().GetAvailableCourier(gomock.Any()).Return(courier, nil)

						factory.EXPECT().CreateCalculator("car").Return(calc)
						calc.EXPECT().
							CalculateDeadline(gomock.Any()).
							DoAndReturn(func(fromTime time.Time) time.Time {
								return fromTime.Add(5 * time.Minute)
							})

						repo.EXPECT().
							Create(gomock.Any(), gomock.Any()).
							Return(int64(1), nil)

						courierRepo.EXPECT().
							Update(gomock.Any(), gomock.Any()).
							DoAndReturn(func(ctx context.Context, c *model.Courier) error {
								assert.Equal(t, "busy", c.Status)
								return nil
							})

						return fn(ctx)
					},
				)
			},
			expectedErr: nil,
		},
		{
			name:    "empty order ID",
			orderID: "",
			mockSetup: func(repo *mocks.MockdeliveryRepository, courierRepo *mocks.MockcourierRepository, factory *mocks.MockDeliveryTimeCalculatorFactory, calc *mocks.MockDeliveryTimeCalculator, txMgr *mocks.MockManager) {
			},
			expectedErr: model.ErrInvalidInput,
		},
		{
			name:    "no available couriers",
			orderID: "order-123",
			mockSetup: func(repo *mocks.MockdeliveryRepository, courierRepo *mocks.MockcourierRepository, factory *mocks.MockDeliveryTimeCalculatorFactory, calc *mocks.MockDeliveryTimeCalculator, txMgr *mocks.MockManager) {
				txMgr.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						repo.EXPECT().GetAvailableCourier(gomock.Any()).Return(nil, model.ErrNoAvailableCouriers)
						return fn(ctx)
					},
				)
			},
			expectedErr: model.ErrNoAvailableCouriers,
		},
		{
			name:    "create delivery fails",
			orderID: "order-123",
			mockSetup: func(repo *mocks.MockdeliveryRepository, courierRepo *mocks.MockcourierRepository, factory *mocks.MockDeliveryTimeCalculatorFactory, calc *mocks.MockDeliveryTimeCalculator, txMgr *mocks.MockManager) {
				courier := &model.Courier{
					ID:            1,
					Name:          "John Doe",
					Phone:         "+1234567890",
					Status:        "available",
					TransportType: "car",
				}

				txMgr.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						repo.EXPECT().GetAvailableCourier(gomock.Any()).Return(courier, nil)

						factory.EXPECT().CreateCalculator("car").Return(calc)
						calc.EXPECT().
							CalculateDeadline(gomock.Any()).
							Return(time.Now().Add(5 * time.Minute))

						repo.EXPECT().
							Create(gomock.Any(), gomock.Any()).
							Return(int64(0), errors.New("db error"))

						return fn(ctx)
					},
				)
			},
			expectedErr: errors.New("db error"),
		},
		{
			name:    "update courier status fails",
			orderID: "order-123",
			mockSetup: func(repo *mocks.MockdeliveryRepository, courierRepo *mocks.MockcourierRepository, factory *mocks.MockDeliveryTimeCalculatorFactory, calc *mocks.MockDeliveryTimeCalculator, txMgr *mocks.MockManager) {
				courier := &model.Courier{
					ID:            1,
					Name:          "John Doe",
					Phone:         "+1234567890",
					Status:        "available",
					TransportType: "car",
				}

				txMgr.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						repo.EXPECT().GetAvailableCourier(gomock.Any()).Return(courier, nil)

						factory.EXPECT().CreateCalculator("car").Return(calc)
						calc.EXPECT().
							CalculateDeadline(gomock.Any()).
							Return(time.Now().Add(5 * time.Minute))

						repo.EXPECT().
							Create(gomock.Any(), gomock.Any()).
							Return(int64(1), nil)

						courierRepo.EXPECT().
							Update(gomock.Any(), gomock.Any()).
							Return(errors.New("update error"))

						return fn(ctx)
					},
				)
			},
			expectedErr: errors.New("update error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockdeliveryRepository(ctrl)
			mockCourierRepo := mocks.NewMockcourierRepository(ctrl)
			mockFactory := mocks.NewMockDeliveryTimeCalculatorFactory(ctrl)
			mockCalc := mocks.NewMockDeliveryTimeCalculator(ctrl)
			mockTxMgr := mocks.NewMockManager(ctrl)

			tt.mockSetup(mockRepo, mockCourierRepo, mockFactory, mockCalc, mockTxMgr)

			log := logger.NopLogger{}
			svc := service.NewDeliveryService(mockRepo, mockCourierRepo, mockFactory, mockTxMgr, log)
			courier, err, delivery := svc.AssignCourier(context.Background(), tt.orderID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, courier)
				assert.Nil(t, delivery)
			} else {
				require.NoError(t, err)
				require.NotNil(t, courier)
				require.NotNil(t, delivery)
				assert.Equal(t, tt.orderID, delivery.OrderID)
				assert.Equal(t, courier.ID, delivery.CourierID)
			}
		})
	}
}

func TestDeliveryService_UnassignCourier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		orderID           string
		mockSetup         func(*mocks.MockdeliveryRepository, *mocks.MockcourierRepository, *mocks.MockManager)
		expectedCourierID int64
		expectedErr       error
	}{
		{
			name:    "successful unassignment",
			orderID: "order-123",
			mockSetup: func(repo *mocks.MockdeliveryRepository, courierRepo *mocks.MockcourierRepository, txMgr *mocks.MockManager) {
				delivery := &model.Delivery{
					ID:        1,
					CourierID: 5,
					OrderID:   "order-123",
				}
				courier := &model.Courier{
					ID:            5,
					Name:          "John Doe",
					Phone:         "+1234567890",
					Status:        "busy",
					TransportType: "car",
				}

				txMgr.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						repo.EXPECT().GetByOrderID(gomock.Any(), "order-123").Return(delivery, nil)
						repo.EXPECT().DeleteByOrderID(gomock.Any(), "order-123").Return(nil)
						courierRepo.EXPECT().GetByID(gomock.Any(), int64(5)).Return(courier, nil)
						courierRepo.EXPECT().
							Update(gomock.Any(), gomock.Any()).
							DoAndReturn(func(ctx context.Context, c *model.Courier) error {
								assert.Equal(t, "available", c.Status)
								return nil
							})
						return fn(ctx)
					},
				)
			},
			expectedCourierID: 5,
			expectedErr:       nil,
		},
		{
			name:    "empty order ID",
			orderID: "",
			mockSetup: func(repo *mocks.MockdeliveryRepository, courierRepo *mocks.MockcourierRepository, txMgr *mocks.MockManager) {
			},
			expectedCourierID: 0,
			expectedErr:       model.ErrInvalidInput,
		},
		{
			name:    "delivery not found",
			orderID: "order-999",
			mockSetup: func(repo *mocks.MockdeliveryRepository, courierRepo *mocks.MockcourierRepository, txMgr *mocks.MockManager) {
				txMgr.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						repo.EXPECT().GetByOrderID(gomock.Any(), "order-999").Return(nil, model.ErrDeliveryNotFound)
						return fn(ctx)
					},
				)
			},
			expectedCourierID: 0,
			expectedErr:       model.ErrDeliveryNotFound,
		},
		{
			name:    "delete delivery fails",
			orderID: "order-123",
			mockSetup: func(repo *mocks.MockdeliveryRepository, courierRepo *mocks.MockcourierRepository, txMgr *mocks.MockManager) {
				delivery := &model.Delivery{
					ID:        1,
					CourierID: 5,
					OrderID:   "order-123",
				}

				txMgr.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						repo.EXPECT().GetByOrderID(gomock.Any(), "order-123").Return(delivery, nil)
						repo.EXPECT().DeleteByOrderID(gomock.Any(), "order-123").Return(errors.New("delete error"))
						return fn(ctx)
					},
				)
			},
			expectedCourierID: 0,
			expectedErr:       errors.New("delete error"),
		},
		{
			name:    "update courier status fails",
			orderID: "order-123",
			mockSetup: func(repo *mocks.MockdeliveryRepository, courierRepo *mocks.MockcourierRepository, txMgr *mocks.MockManager) {
				delivery := &model.Delivery{
					ID:        1,
					CourierID: 5,
					OrderID:   "order-123",
				}
				courier := &model.Courier{
					ID:            5,
					Name:          "John Doe",
					Phone:         "+1234567890",
					Status:        "busy",
					TransportType: "car",
				}

				txMgr.EXPECT().RunInTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						repo.EXPECT().GetByOrderID(gomock.Any(), "order-123").Return(delivery, nil)
						repo.EXPECT().DeleteByOrderID(gomock.Any(), "order-123").Return(nil)
						courierRepo.EXPECT().GetByID(gomock.Any(), int64(5)).Return(courier, nil)
						courierRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("update error"))
						return fn(ctx)
					},
				)
			},
			expectedCourierID: 0,
			expectedErr:       errors.New("update error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockdeliveryRepository(ctrl)
			mockCourierRepo := mocks.NewMockcourierRepository(ctrl)
			mockFactory := mocks.NewMockDeliveryTimeCalculatorFactory(ctrl)
			mockTxMgr := mocks.NewMockManager(ctrl)

			tt.mockSetup(mockRepo, mockCourierRepo, mockTxMgr)

			log := logger.NopLogger{}
			svc := service.NewDeliveryService(mockRepo, mockCourierRepo, mockFactory, mockTxMgr, log)
			courierID, err := svc.UnassignCourier(context.Background(), tt.orderID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, int64(0), courierID)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedCourierID, courierID)
			}
		})
	}
}

func TestDeliveryService_CheckExpiredDeliveries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mockSetup   func(*mocks.MockdeliveryRepository, *mocks.MockcourierRepository, *mocks.MockManager)
		expectedErr error
	}{
		{
			name: "successful check with releases",
			mockSetup: func(repo *mocks.MockdeliveryRepository, courierRepo *mocks.MockcourierRepository, txMgr *mocks.MockManager) {
				txMgr.EXPECT().
					RunInTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				repo.EXPECT().
					ReleaseExpiredDeliveries(gomock.Any()).
					Return([]int64{1, 2, 3}, nil)

				courierRepo.EXPECT().
					UpdateStatusByIDs(gomock.Any(), []int64{1, 2, 3}, "available").
					Return(int64(3), nil)
			},
			expectedErr: nil,
		},
		{
			name: "successful check with no releases",
			mockSetup: func(repo *mocks.MockdeliveryRepository, courierRepo *mocks.MockcourierRepository, txMgr *mocks.MockManager) {
				txMgr.EXPECT().
					RunInTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				repo.EXPECT().
					ReleaseExpiredDeliveries(gomock.Any()).
					Return([]int64{}, nil)
			},
			expectedErr: nil,
		},
		{
			name: "repository error on release",
			mockSetup: func(repo *mocks.MockdeliveryRepository, courierRepo *mocks.MockcourierRepository, txMgr *mocks.MockManager) {
				txMgr.EXPECT().
					RunInTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				repo.EXPECT().
					ReleaseExpiredDeliveries(gomock.Any()).
					Return([]int64(nil), errors.New("db error"))
			},
			expectedErr: errors.New("db error"),
		},
		{
			name: "repository error on update status",
			mockSetup: func(repo *mocks.MockdeliveryRepository, courierRepo *mocks.MockcourierRepository, txMgr *mocks.MockManager) {
				txMgr.EXPECT().
					RunInTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				repo.EXPECT().
					ReleaseExpiredDeliveries(gomock.Any()).
					Return([]int64{1, 2}, nil)

				courierRepo.EXPECT().
					UpdateStatusByIDs(gomock.Any(), []int64{1, 2}, "available").
					Return(int64(0), errors.New("update error"))
			},
			expectedErr: errors.New("update error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockdeliveryRepository(ctrl)
			mockCourierRepo := mocks.NewMockcourierRepository(ctrl)
			mockFactory := mocks.NewMockDeliveryTimeCalculatorFactory(ctrl)
			mockTxMgr := mocks.NewMockManager(ctrl)

			tt.mockSetup(mockRepo, mockCourierRepo, mockTxMgr)

			log := logger.NopLogger{}
			svc := service.NewDeliveryService(mockRepo, mockCourierRepo, mockFactory, mockTxMgr, log)
			err := svc.CheckExpiredDeliveries(context.Background())

			if tt.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
