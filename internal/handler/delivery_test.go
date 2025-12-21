package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Avito-courses/course-go-avito-domovonok/internal/handler"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/handler/mocks"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/logger"
	"github.com/Avito-courses/course-go-avito-domovonok/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -destination=mocks/delivery_service_mock.go -package=mocks github.com/Avito-courses/course-go-avito-domovonok/internal/handler deliveryService

func TestDeliveryHandler_Assign(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*mocks.MockdeliveryService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful assignment",
			requestBody: handler.AssignRequestDTO{
				OrderID: "order-123",
			},
			mockSetup: func(m *mocks.MockdeliveryService) {
				deadline := time.Now().Add(5 * time.Minute)
				courier := &model.Courier{
					ID:            1,
					Name:          "John Doe",
					Phone:         "+1234567890",
					Status:        "busy",
					TransportType: "car",
				}
				delivery := &model.Delivery{
					ID:         1,
					CourierID:  1,
					OrderID:    "order-123",
					AssignedAt: time.Now(),
					Deadline:   deadline,
				}
				m.EXPECT().
					AssignCourier(gomock.Any(), "order-123").
					Return(courier, nil, delivery)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.AssignResponseDTO
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, int64(1), response.CourierID)
				assert.Equal(t, "order-123", response.OrderID)
			},
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			mockSetup:      func(m *mocks.MockdeliveryService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "no available couriers",
			requestBody: handler.AssignRequestDTO{
				OrderID: "order-123",
			},
			mockSetup: func(m *mocks.MockdeliveryService) {
				m.EXPECT().
					AssignCourier(gomock.Any(), "order-123").
					Return(nil, model.ErrNoAvailableCouriers, nil)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockdeliveryService(ctrl)
			tt.mockSetup(mockService)

			log := logger.NewNopLogger()
			h := handler.NewDeliveryHandler(mockService, log)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/delivery/assign", bytes.NewReader(body))
			w := httptest.NewRecorder()

			h.Assign(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}

func TestDeliveryHandler_Unassign(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*mocks.MockdeliveryService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful unassignment",
			requestBody: handler.UnassignRequestDTO{
				OrderID: "order-123",
			},
			mockSetup: func(m *mocks.MockdeliveryService) {
				m.EXPECT().
					UnassignCourier(gomock.Any(), "order-123").
					Return(int64(5), nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response handler.UnassignResponseDTO
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, "order-123", response.OrderID)
				assert.Equal(t, int64(5), response.CourierID)
			},
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			mockSetup:      func(m *mocks.MockdeliveryService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "delivery not found",
			requestBody: handler.UnassignRequestDTO{
				OrderID: "order-999",
			},
			mockSetup: func(m *mocks.MockdeliveryService) {
				m.EXPECT().
					UnassignCourier(gomock.Any(), "order-999").
					Return(int64(0), model.ErrDeliveryNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "internal error",
			requestBody: handler.UnassignRequestDTO{
				OrderID: "order-123",
			},
			mockSetup: func(m *mocks.MockdeliveryService) {
				m.EXPECT().
					UnassignCourier(gomock.Any(), "order-123").
					Return(int64(0), errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockdeliveryService(ctrl)
			tt.mockSetup(mockService)

			log := logger.NewNopLogger()
			h := handler.NewDeliveryHandler(mockService, log)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/delivery/unassign", bytes.NewReader(body))
			w := httptest.NewRecorder()

			h.Unassign(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
		})
	}
}
