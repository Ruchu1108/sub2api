package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestUserHandlerBatchResetBalance(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	adminSvc := newStubAdminService()
	handler := NewUserHandler(adminSvc, nil, nil, nil, nil, nil, nil)
	router.POST("/api/v1/admin/users/batch-reset-balance", handler.BatchResetBalance)

	body, _ := json.Marshal(map[string]any{"user_ids": []int64{1, 2, 3}})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/batch-reset-balance", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, []int64{1, 2, 3}, adminSvc.batchResetUserIDs)

	var resp struct {
		Data struct {
			Affected int `json:"affected"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, 3, resp.Data.Affected)
}

func TestUserHandlerBatchResetBalance_RequiresUserIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	adminSvc := newStubAdminService()
	handler := NewUserHandler(adminSvc, nil, nil, nil, nil, nil, nil)
	router.POST("/api/v1/admin/users/batch-reset-balance", handler.BatchResetBalance)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/batch-reset-balance", bytes.NewBufferString(`{"user_ids":[]}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Nil(t, adminSvc.batchResetUserIDs, "空列表不应触发服务调用")
}

func TestUserHandlerBatchResetBalance_LimitsTo500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	adminSvc := newStubAdminService()
	handler := NewUserHandler(adminSvc, nil, nil, nil, nil, nil, nil)
	router.POST("/api/v1/admin/users/batch-reset-balance", handler.BatchResetBalance)

	ids := make([]int64, 501)
	for i := range ids {
		ids[i] = int64(i + 1)
	}
	body, _ := json.Marshal(map[string]any{"user_ids": ids})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/batch-reset-balance", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Nil(t, adminSvc.batchResetUserIDs, "超过 500 个用户不应触发服务调用")
}
