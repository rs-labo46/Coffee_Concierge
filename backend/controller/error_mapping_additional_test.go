package controller

import (
	"errors"
	"net/http"
	"testing"

	"coffee-spa/usecase"
)

func TestMapError_AdditionalCoverage(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "C-ERR-ADD-01 unauthorized", err: usecase.ErrUnauthorized, wantStatus: http.StatusUnauthorized, wantCode: "unauthorized"},
		{name: "C-ERR-ADD-02 forbidden", err: usecase.ErrForbidden, wantStatus: http.StatusForbidden, wantCode: "forbidden"},
		{name: "C-ERR-ADD-03 not found", err: usecase.ErrNotFound, wantStatus: http.StatusNotFound, wantCode: "not_found"},
		{name: "C-ERR-ADD-04 conflict", err: usecase.ErrConflict, wantStatus: http.StatusConflict, wantCode: "conflict"},
		{name: "C-ERR-ADD-05 rate limited", err: usecase.ErrRateLimited, wantStatus: http.StatusTooManyRequests, wantCode: "rate_limited"},
		{name: "C-ERR-ADD-06 unknown", err: errors.New("unknown"), wantStatus: http.StatusInternalServerError, wantCode: "internal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, code, _ := mapError(tt.err)
			if status != tt.wantStatus || code != tt.wantCode {
				t.Fatalf("mapError() = %d/%s, want %d/%s", status, code, tt.wantStatus, tt.wantCode)
			}
		})
	}
}
