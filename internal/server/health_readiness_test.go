package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Gentleman-Programming/engram/v3/internal/store"
)

func TestHealthUsesBoundedReadinessInsteadOfStats(t *testing.T) {
	st := newServerTestStore(t)
	srv := New(st, 0)
	oldStats, oldReadiness := loadServerStats, checkServerReadiness
	t.Cleanup(func() { loadServerStats, checkServerReadiness = oldStats, oldReadiness })
	loadServerStats = func(*store.Store) (*store.Stats, error) {
		t.Fatal("health must not execute database-wide statistics")
		return nil, errors.New("unexpected stats call")
	}
	called := false
	checkServerReadiness = func(s *store.Store) error {
		called = true
		return s.CheckReadiness()
	}
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK || !called {
		t.Fatalf("health status=%d readiness_called=%t body=%s", rec.Code, called, rec.Body.String())
	}
}

func TestHealthReportsReadinessFailure(t *testing.T) {
	st := newServerTestStore(t)
	srv := New(st, 0)
	if err := st.Close(); err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("health status=%d body=%s", rec.Code, rec.Body.String())
	}
}
