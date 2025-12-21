package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/handler"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/handler/mocks"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/logger"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -destination=mocks/courier_service_mock.go -package=mocks github.com/Avito-courses/course-go-avito-domovonok/internal/handler courierService

func TestCourierHandler_Ping(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockcourierService(ctrl)
	log := logger.NopLogger{}
	h := handler.NewCourierHandler(mockService, log)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	h.Ping(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "pong", response["message"])
}

func TestCourierHandler_Healthcheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		mockSetup      func(*mocks.MockcourierService)
		expectedStatus int
	}{
		{
			name: "healthy",
			mockSetup: func(m *mocks.MockcourierService) {
				m.EXPECT().HealthCheck(gomock.Any()).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "unhealthy",
			mockSetup: func(m *mocks.MockcourierService) {
				m.EXPECT().HealthCheck(gomock.Any()).Return(errors.New("db error"))
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockcourierService(ctrl)
			tt.mockSetup(mockService)

			log := logger.NopLogger{}
			h := handler.NewCourierHandler(mockService, log)

			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()

			h.Healthcheck(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCourierHandler_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*mocks.MockcourierService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful creation",
			requestBody: handler.CourierDTO{
				Name:          "John Doe",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "on_foot",
			},
			mockSetup: func(m *mocks.MockcourierService) {
				m.EXPECT().
					CreateCourier(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, c *model.Courier) (*model.Courier, error) {
						c.ID = 1
						return c, nil
					})
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.CourierDTO
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, int64(1), response.ID)
				assert.Equal(t, "John Doe", response.Name)
			},
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			mockSetup:      func(m *mocks.MockcourierService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.HTTPError
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, "invalid input", response.Message)
			},
		},
		{
			name: "service returns invalid input error",
			requestBody: handler.CourierDTO{
				Name:          "",
				Phone:         "+1234567890",
				Status:        "available",
				TransportType: "on_foot",
			},
			mockSetup: func(m *mocks.MockcourierService) {
				m.EXPECT().
					CreateCourier(gomock.Any(), gomock.Any()).
					Return(nil, model.ErrInvalidInput)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockcourierService(ctrl)
			tt.mockSetup(mockService)

			log := logger.NopLogger{}
			h := handler.NewCourierHandler(mockService, log)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/couriers", bytes.NewReader(body))
			w := httptest.NewRecorder()

			h.Create(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestCourierHandler_Get(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		courierID      string
		mockSetup      func(*mocks.MockcourierService)
		expectedStatus int
	}{
		{
			name:      "successful get",
			courierID: "1",
			mockSetup: func(m *mocks.MockcourierService) {
				m.EXPECT().
					GetCourier(gomock.Any(), int64(1)).
					Return(&model.Courier{
						ID:            1,
						Name:          "John Doe",
						Phone:         "+1234567890",
						Status:        "available",
						TransportType: "on_foot",
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid ID - not a number",
			courierID:      "abc",
			mockSetup:      func(m *mocks.MockcourierService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "courier not found",
			courierID: "999",
			mockSetup: func(m *mocks.MockcourierService) {
				m.EXPECT().
					GetCourier(gomock.Any(), int64(999)).
					Return(nil, model.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockcourierService(ctrl)
			tt.mockSetup(mockService)

			log := logger.NopLogger{}
			h := handler.NewCourierHandler(mockService, log)

			req := httptest.NewRequest(http.MethodGet, "/couriers/"+tt.courierID, nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.courierID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			h.Get(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCourierHandler_List(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		mockSetup      func(*mocks.MockcourierService)
		expectedStatus int
	}{
		{
			name: "successful list with couriers",
			mockSetup: func(m *mocks.MockcourierService) {
				m.EXPECT().
					ListCouriers(gomock.Any()).
					Return([]*model.Courier{
						{ID: 1, Name: "John Doe", Phone: "+1234567890", Status: "available", TransportType: "on_foot"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "service error",
			mockSetup: func(m *mocks.MockcourierService) {
				m.EXPECT().
					ListCouriers(gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockcourierService(ctrl)
			tt.mockSetup(mockService)

			log := logger.NopLogger{}
			h := handler.NewCourierHandler(mockService, log)

			req := httptest.NewRequest(http.MethodGet, "/couriers", nil)
			w := httptest.NewRecorder()

			h.List(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCourierHandler_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*mocks.MockcourierService)
		expectedStatus int
	}{
		{
			name: "successful update",
			requestBody: handler.CourierDTO{
				ID:            1,
				Name:          "John Doe Updated",
				Phone:         "+1234567890",
				Status:        "busy",
				TransportType: "car",
			},
			mockSetup: func(m *mocks.MockcourierService) {
				m.EXPECT().
					UpdateCourier(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, c *model.Courier) (*model.Courier, error) {
						return c, nil
					})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			mockSetup:      func(m *mocks.MockcourierService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockcourierService(ctrl)
			tt.mockSetup(mockService)

			log := logger.NopLogger{}
			h := handler.NewCourierHandler(mockService, log)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/couriers", bytes.NewReader(body))
			w := httptest.NewRecorder()

			h.Update(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
