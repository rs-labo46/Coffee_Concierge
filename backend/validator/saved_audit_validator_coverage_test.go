package validator_test

import (
	"testing"

	"coffee-spa/usecase"
	"coffee-spa/validator"
)

func TestSavedAuditValidators_Coverage(t *testing.T) {
	savedVal := validator.NewSavedValidator()
	auditVal := validator.NewAuditValidator()
	uid := uint(1)
	zero := uint(0)

	tests := []struct {
		name string
		fn   func() error
		want bool
	}{
		{name: "V-SAVED-ADD-01 save ok", fn: func() error { return savedVal.Save(usecase.SaveSuggestionIn{SessionID: 1, SuggestionID: 2}) }, want: false},
		{name: "V-SAVED-ADD-02 save bad session", fn: func() error { return savedVal.Save(usecase.SaveSuggestionIn{SessionID: 0, SuggestionID: 2}) }, want: true},
		{name: "V-SAVED-ADD-03 delete bad id", fn: func() error { return savedVal.Delete(usecase.DeleteSavedIn{SuggestionID: 0}) }, want: true},
		{name: "V-SAVED-ADD-04 list bad offset", fn: func() error { return savedVal.List(usecase.ListSavedIn{Limit: 10, Offset: -1}) }, want: true},
		{name: "V-AUDIT-ADD-01 audit list ok", fn: func() error {
			return auditVal.List(usecase.AuditListIn{Type: "ai.turn.add", UserID: &uid, Limit: 20, Offset: 0})
		}, want: false},
		{name: "V-AUDIT-ADD-02 audit bad user id", fn: func() error { return auditVal.List(usecase.AuditListIn{UserID: &zero, Limit: 20, Offset: 0}) }, want: true},
		{name: "V-AUDIT-ADD-03 audit bad limit", fn: func() error { return auditVal.List(usecase.AuditListIn{Limit: 101, Offset: 0}) }, want: true},
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
