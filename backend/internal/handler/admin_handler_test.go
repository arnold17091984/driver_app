package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kento/driver/backend/internal/model"
)

func TestAdmin_ListUsers_Success(t *testing.T) {
	repo := &mockUserRepo{
		listFn: func(ctx context.Context) ([]model.User, error) {
			return []model.User{{ID: "u1", Name: "Test"}}, nil
		},
	}
	h := NewAdminHandler(repo, &mockAuditSvc{})
	req := httptest.NewRequest("GET", "/admin/users", nil)
	rec := httptest.NewRecorder()

	h.ListUsers(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestAdmin_UpdateRole_InvalidRole(t *testing.T) {
	h := NewAdminHandler(&mockUserRepo{}, &mockAuditSvc{})
	body := `{"role":"superuser"}`
	req := httptest.NewRequest("PUT", "/admin/users/u1/role", strings.NewReader(body))
	req = withChiParam(req, "id", "u1")
	req = withClaims(req, "admin1", "adm1", "admin")
	rec := httptest.NewRecorder()

	h.UpdateRole(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAdmin_UpdateRole_InvalidJSON(t *testing.T) {
	h := NewAdminHandler(&mockUserRepo{}, &mockAuditSvc{})
	req := httptest.NewRequest("PUT", "/admin/users/u1/role", strings.NewReader("bad"))
	req = withChiParam(req, "id", "u1")
	req = withClaims(req, "admin1", "adm1", "admin")
	rec := httptest.NewRecorder()

	h.UpdateRole(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAdmin_ListAuditLogs_InvalidLimit(t *testing.T) {
	h := NewAdminHandler(&mockUserRepo{}, &mockAuditSvc{})
	req := httptest.NewRequest("GET", "/admin/audit-logs?limit=abc", nil)
	rec := httptest.NewRecorder()

	h.ListAuditLogs(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAdmin_ListAuditLogs_InvalidFrom(t *testing.T) {
	h := NewAdminHandler(&mockUserRepo{}, &mockAuditSvc{})
	req := httptest.NewRequest("GET", "/admin/audit-logs?from=invalid", nil)
	rec := httptest.NewRecorder()

	h.ListAuditLogs(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAdmin_ListAuditLogs_Success(t *testing.T) {
	h := NewAdminHandler(&mockUserRepo{}, &mockAuditSvc{})
	req := httptest.NewRequest("GET", "/admin/audit-logs?limit=10", nil)
	rec := httptest.NewRecorder()

	h.ListAuditLogs(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
