package usecase

import (
	"errors"
	"testing"
	"time"
)

func TestRateLimitUC_AllowSignup_UsesRuleAndClock(t *testing.T) {
	now := time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC)
	called := false
	uc := NewRateLimitUC(
		rateLimitStoreMock{allowFn: func(key string, rate float64, capacity float64, cost float64, gotNow time.Time) (bool, int, error) {
			called = true
			if key != "rl:signup:ip:127.0.0.1" || rate != 1 || capacity != 5 || cost != 1 || !gotNow.Equal(now) {
				t.Fatalf("unexpected args key=%s rate=%v cap=%v cost=%v now=%v", key, rate, capacity, cost, gotNow)
			}
			return true, 0, nil
		}},
		fixedClock{now: now},
		RateRule{Rate: 1, Capacity: 5, Cost: 1},
		RateRule{Rate: 1, Capacity: 5, Cost: 1},
		RateRule{Rate: 1, Capacity: 5, Cost: 1},
		RateRule{Rate: 1, Capacity: 5, Cost: 1},
		RateRule{Rate: 1, Capacity: 5, Cost: 1},
		RateRule{Rate: 1, Capacity: 5, Cost: 1},
		RateRule{Rate: 1, Capacity: 5, Cost: 1},
		RateRule{Rate: 1, Capacity: 5, Cost: 1},
		RateRule{Rate: 1, Capacity: 5, Cost: 1},
	)
	ok, retry, err := uc.AllowSignup("127.0.0.1")
	if err != nil || !ok || retry != 0 || !called {
		t.Fatalf("unexpected result ok=%v retry=%v err=%v called=%v", ok, retry, err, called)
	}
}

func TestRateLimitUC_AllowLogin_Denied(t *testing.T) {
	uc := NewRateLimitUC(rateLimitStoreMock{allowFn: func(key string, rate float64, capacity float64, cost float64, now time.Time) (bool, int, error) {
		return false, 9, nil
	}}, fixedClock{now: time.Now()}, RateRule{1,1,1}, RateRule{1,1,1}, RateRule{1,1,1}, RateRule{1,1,1}, RateRule{1,1,1}, RateRule{1,1,1}, RateRule{1,1,1}, RateRule{1,1,1}, RateRule{1,1,1})
	ok, retry, err := uc.AllowLogin("mailhash")
	if err != nil || ok || retry != 9 { t.Fatalf("unexpected result ok=%v retry=%v err=%v", ok, retry, err) }
}

func TestRateLimitUC_AllowWS_StoreError(t *testing.T) {
	expected := errors.New("redis down")
	uc := NewRateLimitUC(rateLimitStoreMock{allowFn: func(key string, rate float64, capacity float64, cost float64, now time.Time) (bool, int, error) {
		return false, 0, expected
	}}, fixedClock{now: time.Now()}, RateRule{1,1,1}, RateRule{1,1,1}, RateRule{1,1,1}, RateRule{1,1,1}, RateRule{1,1,1}, RateRule{1,1,1}, RateRule{1,1,1}, RateRule{1,1,1}, RateRule{1,1,1})
	_, _, err := uc.AllowWS("rl:ws:user:1")
	if !errors.Is(err, expected) { t.Fatalf("expected store error, got:%v", err) }
}
