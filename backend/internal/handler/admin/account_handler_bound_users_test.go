package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func setupBoundUsersRouter(adminSvc service.AdminService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := &AccountHandler{adminService: adminSvc}
	router.GET("/api/v1/admin/accounts/:id/bound-users", handler.GetBoundUsers)
	router.PUT("/api/v1/admin/accounts/:id/bound-users", handler.SetBoundUsers)
	return router
}

func TestAccountHandlerGetBoundUsers(t *testing.T) {
	adminSvc := newStubAdminService()
	adminSvc.boundUsers = []service.User{
		{ID: 7, Email: "a@example.com", Username: "alice", Balance: 1.5, DefaultAmount: 2},
		{ID: 8, Email: "b@example.com", Username: "bob", Balance: 0, DefaultAmount: 5},
	}
	router := setupBoundUsersRouter(adminSvc)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/accounts/9/bound-users", nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var envelope struct {
		Code int `json:"code"`
		Data struct {
			Users []struct {
				ID            int64   `json:"id"`
				Email         string  `json:"email"`
				DefaultAmount float64 `json:"default_amount"`
			} `json:"users"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Len(t, envelope.Data.Users, 2)
	require.Equal(t, int64(7), envelope.Data.Users[0].ID)
	require.Equal(t, "a@example.com", envelope.Data.Users[0].Email)
	require.Equal(t, 2.0, envelope.Data.Users[0].DefaultAmount)
}

func TestAccountHandlerSetBoundUsers(t *testing.T) {
	adminSvc := newStubAdminService()
	router := setupBoundUsersRouter(adminSvc)

	body, _ := json.Marshal(map[string]any{"user_ids": []int64{7, 8, 7}})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/accounts/9/bound-users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, int64(9), adminSvc.setBoundUsersAccountID)
	require.Equal(t, []int64{7, 8, 7}, adminSvc.setBoundUsersUserIDs)

	var envelope struct {
		Code int `json:"code"`
		Data struct {
			BoundCount int `json:"bound_count"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.Equal(t, 3, envelope.Data.BoundCount)
}

func TestAccountHandlerSetBoundUsers_RejectsTooManyIDs(t *testing.T) {
	adminSvc := newStubAdminService()
	router := setupBoundUsersRouter(adminSvc)

	ids := make([]int64, 501)
	for i := range ids {
		ids[i] = int64(i + 1)
	}
	body, _ := json.Marshal(map[string]any{"user_ids": ids})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/accounts/9/bound-users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Empty(t, adminSvc.setBoundUsersUserIDs, "超限请求不应触发服务调用")
}
