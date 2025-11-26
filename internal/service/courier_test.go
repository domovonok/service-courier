package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -destination=mocks/courier_repository_mock.go -package=mocks github.com/Avito-courses/course-go-avito-domovonok/internal/service courierRepository

func TestCourierService_CreateCourier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		courier     *model.Courier
		mockSetup   func(*mocks.MockcourierRepository)
		expectedID  int64
		expectedErr error
	}{
		{
			name: "successful creation with explicit transport type",
			courier: &model.Courier{
				Name:          "John Doe",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "car",
			},
			mockSetup: func(m *mocks.MockcourierRepository) {
				m.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(int64(1), nil)
			},
			expectedID:  1,
			expectedErr: nil,
		},
		{
			name: "successful creation with default transport type",
			courier: &model.Courier{
				Name:   "Jane Smith",
				Phone:  "+9876543210",
				Status: "available",
			},
			mockSetup: func(m *mocks.MockcourierRepository) {
				m.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, c *model.Courier) (int64, error) {
						assert.Equal(t, "on_foot", c.TransportType)
						return int64(2), nil
					})
			},
			expectedID:  2,
			expectedErr: nil,
		},
		{
			name: "invalid courier data - empty name",
			courier: &model.Courier{
				Name:          "",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "on_foot",
			},
			mockSetup:   func(m *mocks.MockcourierRepository) {},
			expectedID:  0,
			expectedErr: model.ErrInvalidInput,
		},
		{
			name: "invalid courier data - invalid phone",
			courier: &model.Courier{
				Name:          "John Doe",
				Phone:         "invalid",
				Status:        "available",
				TransportType: "on_foot",
			},
			mockSetup:   func(m *mocks.MockcourierRepository) {},
			expectedID:  0,
			expectedErr: model.ErrInvalidInput,
		},
		{
			name: "invalid courier data - invalid transport type",
			courier: &model.Courier{
				Name:          "John Doe",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "bicycle",
			},
			mockSetup:   func(m *mocks.MockcourierRepository) {},
			expectedID:  0,
			expectedErr: model.ErrInvalidInput,
		},
		{
			name: "repository error - conflict",
			courier: &model.Courier{
				Name:          "John Doe",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "on_foot",
			},
			mockSetup: func(m *mocks.MockcourierRepository) {
				m.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(int64(0), model.ErrConflict)
			},
			expectedID:  0,
			expectedErr: model.ErrConflict,
		},
		{
			name: "repository error - generic error",
			courier: &model.Courier{
				Name:          "John Doe",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "on_foot",
			},
			mockSetup: func(m *mocks.MockcourierRepository) {
				m.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(int64(0), errors.New("database error"))
			},
			expectedID:  0,
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockcourierRepository(ctrl)
			tt.mockSetup(mockRepo)

			service := NewCourierService(mockRepo)
			result, err := service.CreateCourier(context.Background(), tt.courier)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
				if errors.Is(tt.expectedErr, model.ErrInvalidInput) {
					assert.ErrorIs(t, err, model.ErrInvalidInput)
				} else if errors.Is(tt.expectedErr, model.ErrConflict) {
					assert.ErrorIs(t, err, model.ErrConflict)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedID, result.ID)
			}
		})
	}
}

func TestCourierService_GetCourier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		courierID    int64
		mockSetup    func(*mocks.MockcourierRepository)
		expectedData *model.Courier
		expectedErr  error
	}{
		{
			name:      "successful get",
			courierID: 1,
			mockSetup: func(m *mocks.MockcourierRepository) {
				m.EXPECT().
					GetByID(gomock.Any(), int64(1)).
					Return(&model.Courier{
						ID:            1,
						Name:          "John Doe",
						Phone:         "+1234567890",
						Status:        "available",
						TransportType: "on_foot",
					}, nil)
			},
			expectedData: &model.Courier{
				ID:            1,
				Name:          "John Doe",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "on_foot",
			},
			expectedErr: nil,
		},
		{
			name:         "invalid ID - zero",
			courierID:    0,
			mockSetup:    func(m *mocks.MockcourierRepository) {},
			expectedData: nil,
			expectedErr:  model.ErrInvalidID,
		},
		{
			name:         "invalid ID - negative",
			courierID:    -1,
			mockSetup:    func(m *mocks.MockcourierRepository) {},
			expectedData: nil,
			expectedErr:  model.ErrInvalidID,
		},
		{
			name:      "courier not found",
			courierID: 999,
			mockSetup: func(m *mocks.MockcourierRepository) {
				m.EXPECT().
					GetByID(gomock.Any(), int64(999)).
					Return(nil, model.ErrNotFound)
			},
			expectedData: nil,
			expectedErr:  model.ErrNotFound,
		},
		{
			name:      "repository error",
			courierID: 1,
			mockSetup: func(m *mocks.MockcourierRepository) {
				m.EXPECT().
					GetByID(gomock.Any(), int64(1)).
					Return(nil, errors.New("database error"))
			},
			expectedData: nil,
			expectedErr:  errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockcourierRepository(ctrl)
			tt.mockSetup(mockRepo)

			service := NewCourierService(mockRepo)
			result, err := service.GetCourier(context.Background(), tt.courierID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedData, result)
			}
		})
	}
}

func TestCourierService_ListCouriers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		mockSetup    func(*mocks.MockcourierRepository)
		expectedData []*model.Courier
		expectedErr  error
	}{
		{
			name: "successful list with multiple couriers",
			mockSetup: func(m *mocks.MockcourierRepository) {
				m.EXPECT().
					List(gomock.Any()).
					Return([]*model.Courier{
						{ID: 1, Name: "John Doe", Phone: "+1234567890", Status: "available", TransportType: "on_foot"},
						{ID: 2, Name: "Jane Smith", Phone: "+9876543210", Status: "busy", TransportType: "car"},
					}, nil)
			},
			expectedData: []*model.Courier{
				{ID: 1, Name: "John Doe", Phone: "+1234567890", Status: "available", TransportType: "on_foot"},
				{ID: 2, Name: "Jane Smith", Phone: "+9876543210", Status: "busy", TransportType: "car"},
			},
			expectedErr: nil,
		},
		{
			name: "successful list with empty result",
			mockSetup: func(m *mocks.MockcourierRepository) {
				m.EXPECT().
					List(gomock.Any()).
					Return([]*model.Courier{}, nil)
			},
			expectedData: []*model.Courier{},
			expectedErr:  nil,
		},
		{
			name: "repository error",
			mockSetup: func(m *mocks.MockcourierRepository) {
				m.EXPECT().
					List(gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			expectedData: nil,
			expectedErr:  errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockcourierRepository(ctrl)
			tt.mockSetup(mockRepo)

			service := NewCourierService(mockRepo)
			result, err := service.ListCouriers(context.Background())

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedData, result)
			}
		})
	}
}

func TestCourierService_UpdateCourier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		courier     *model.Courier
		mockSetup   func(*mocks.MockcourierRepository)
		expectedErr error
	}{
		{
			name: "successful update",
			courier: &model.Courier{
				ID:            1,
				Name:          "John Doe Updated",
				Phone:         "+1234567890",
				Status:        "busy",
				TransportType: "car",
			},
			mockSetup: func(m *mocks.MockcourierRepository) {
				m.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "invalid courier - no ID",
			courier: &model.Courier{
				ID:            0,
				Name:          "John Doe",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "on_foot",
			},
			mockSetup:   func(m *mocks.MockcourierRepository) {},
			expectedErr: model.ErrInvalidInput,
		},
		{
			name: "invalid courier - empty name",
			courier: &model.Courier{
				ID:            1,
				Name:          "",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "on_foot",
			},
			mockSetup:   func(m *mocks.MockcourierRepository) {},
			expectedErr: model.ErrInvalidInput,
		},
		{
			name: "courier not found",
			courier: &model.Courier{
				ID:            999,
				Name:          "John Doe",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "on_foot",
			},
			mockSetup: func(m *mocks.MockcourierRepository) {
				m.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(model.ErrNotFound)
			},
			expectedErr: model.ErrNotFound,
		},
		{
			name: "repository error",
			courier: &model.Courier{
				ID:            1,
				Name:          "John Doe",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "on_foot",
			},
			mockSetup: func(m *mocks.MockcourierRepository) {
				m.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))
			},
			expectedErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockcourierRepository(ctrl)
			tt.mockSetup(mockRepo)

			service := NewCourierService(mockRepo)
			result, err := service.UpdateCourier(context.Background(), tt.courier)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.courier, result)
			}
		})
	}
}

func TestCourierService_HealthCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mockSetup   func(*mocks.MockcourierRepository)
		expectedErr error
	}{
		{
			name: "successful health check",
			mockSetup: func(m *mocks.MockcourierRepository) {
				m.EXPECT().
					Ping(gomock.Any()).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "failed health check",
			mockSetup: func(m *mocks.MockcourierRepository) {
				m.EXPECT().
					Ping(gomock.Any()).
					Return(errors.New("connection failed"))
			},
			expectedErr: errors.New("connection failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockcourierRepository(ctrl)
			tt.mockSetup(mockRepo)

			service := NewCourierService(mockRepo)
			err := service.HealthCheck(context.Background())

			if tt.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
