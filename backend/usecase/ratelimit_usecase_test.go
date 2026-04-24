package usecase

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRateLimitUC_AllowSignup_UsesRuleAndClock(t *testing.T) {
	now := time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC)
	called := false
	uc := NewRateLimitUC(
		rateLimitStoreMock{allowFn: func(key string, rate float64, capacity float64, cost float64, gotNow time.Time) (bool, int, error) {
			called = true
			if key != "rl:signup:ip:127.0.0.1" || rate != 0.1 || capacity != 3 || cost != 1 || !gotNow.Equal(now) {
				t.Fatalf("unexpected args key=%s rate=%v cap=%v cost=%v now=%v", key, rate, capacity, cost, gotNow)
			}
			return true, 0, nil
		}},
		fixedClock{now: now},
	)
	ok, retry, err := uc.AllowSignup("127.0.0.1")
	if err != nil || !ok || retry != 0 || !called {
		t.Fatalf("unexpected result ok=%v retry=%v err=%v called=%v", ok, retry, err, called)
	}
}

func TestRateLimitUC_AllowLogin_Denied(t *testing.T) {
	uc := NewRateLimitUC(rateLimitStoreMock{allowFn: func(key string, rate float64, capacity float64, cost float64, now time.Time) (bool, int, error) {
		if key != "rl:login:mail:mailhash" {
			t.Fatalf("unexpected key:%s", key)
		}
		return false, 9, nil
	}}, fixedClock{now: time.Now()})
	ok, retry, err := uc.AllowLogin("mailhash")
	if err != nil || ok || retry != 9 {
		t.Fatalf("unexpected result ok=%v retry=%v err=%v", ok, retry, err)
	}
}

func TestRateLimitUC_AllowWS_StoreError(t *testing.T) {
	expected := errors.New("redis down")
	uc := NewRateLimitUC(rateLimitStoreMock{allowFn: func(key string, rate float64, capacity float64, cost float64, now time.Time) (bool, int, error) {
		return false, 0, expected
	}}, fixedClock{now: time.Now()})
	_, _, err := uc.AllowWS("rl:ws:user:1")
	if !errors.Is(err, expected) {
		t.Fatalf("expected store error, got:%v", err)
	}
}

func TestRateLimitUC_Allow_RejectsBrokenDependenciesAndRules(t *testing.T) {
	cases := []struct {
		name string
		uc   *RateLimitUC
		key  string
		want string
	}{
		{name: "storeなし", uc: &RateLimitUC{clock: fixedClock{now: time.Now()}}, key: "k", want: "store is nil"},
		{name: "clockなし", uc: &RateLimitUC{store: rateLimitStoreMock{}}, key: "k", want: "clock is nil"},
		{name: "key空", uc: &RateLimitUC{store: rateLimitStoreMock{}, clock: fixedClock{now: time.Now()}, signupIP: RateRule{Rate: 1, Capacity: 1, Cost: 1}}, key: "", want: "key is empty"},
		{name: "rate不正", uc: &RateLimitUC{store: rateLimitStoreMock{}, clock: fixedClock{now: time.Now()}, signupIP: RateRule{Rate: 0, Capacity: 1, Cost: 1}}, key: "k", want: "invalid rule rate"},
		{name: "capacity不正", uc: &RateLimitUC{store: rateLimitStoreMock{}, clock: fixedClock{now: time.Now()}, signupIP: RateRule{Rate: 1, Capacity: 0, Cost: 1}}, key: "k", want: "invalid rule capacity"},
		{name: "cost不正", uc: &RateLimitUC{store: rateLimitStoreMock{}, clock: fixedClock{now: time.Now()}, signupIP: RateRule{Rate: 1, Capacity: 1, Cost: 0}}, key: "k", want: "invalid rule cost"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := tt.uc.allow(tt.key, tt.uc.signupIP)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected %q, got:%v", tt.want, err)
			}
		})
	}
}
