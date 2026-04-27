package validator_test

import (
	"testing"

	"coffee-spa/validator"
)

func TestAuthValidator_Coverage(t *testing.T) {
	v := validator.NewAuthValidator()

	tests := []struct {
		name string
		fn   func() error
		want bool
	}{
		{name: "V-AUTH-ADD-01 signup ok", fn: func() error { return v.Signup("user@example.com", "password123") }, want: false},
		{name: "V-AUTH-ADD-02 signup invalid email", fn: func() error { return v.Signup("bad", "password123") }, want: true},
		{name: "V-AUTH-ADD-03 signup short password", fn: func() error { return v.Signup("user@example.com", "short") }, want: true},
		{name: "V-AUTH-ADD-04 login allows short existing password input", fn: func() error { return v.Login("user@example.com", "x") }, want: false},
		{name: "V-AUTH-ADD-05 login empty password", fn: func() error { return v.Login("user@example.com", "") }, want: true},
		{name: "V-AUTH-ADD-06 new password too short", fn: func() error { return v.NewPw("short") }, want: true},
		{name: "V-AUTH-ADD-07 token ok", fn: func() error { return v.Token("1234567890123456") }, want: false},
		{name: "V-AUTH-ADD-08 token too short", fn: func() error { return v.Token("short") }, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if (err != nil) != tt.want {
				t.Fatalf("error presence = %v, want %v, err=%v", err != nil, tt.want, err)
			}
		})
	}
}
