package repository

import "testing"

func TestNewRateLimitRepository_ReturnsRepository(t *testing.T) {
	repo := NewRateLimitRepository(nil)
	if repo == nil {
		t.Fatal("NewRateLimitRepository(nil) = nil")
	}
}
