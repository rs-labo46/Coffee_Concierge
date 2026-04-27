package usecase_test

import (
	"testing"
	"time"

	"coffee-spa/testutil/repomock"
	"coffee-spa/testutil/servicemock"
	"coffee-spa/usecase"
)

func TestRateLimitUC_AllowLoginIP_UsesLoginIPKey(t *testing.T) {
	store := &repomock.RateLimit{}
	uc := usecase.NewRateLimitUC(store, servicemock.Clock{NowFn: func() time.Time { return time.Unix(1, 0).UTC() }})

	allowed, retryAfter, err := uc.AllowLoginIP("127.0.0.1")
	if err != nil {
		t.Fatalf("AllowLoginIP() error = %v", err)
	}
	if !allowed || retryAfter != 0 {
		t.Fatalf("allowed/retryAfter = %v/%d", allowed, retryAfter)
	}
	if store.LastKey != "rl:login:ip:127.0.0.1" {
		t.Fatalf("key = %q", store.LastKey)
	}
	if store.LastRate <= 0 || store.LastCapacity <= 0 || store.LastCost <= 0 {
		t.Fatalf("invalid rule = rate:%v capacity:%v cost:%v", store.LastRate, store.LastCapacity, store.LastCost)
	}
}
