package validator_test

import (
	"strings"
	"testing"

	"coffee-spa/entity"
	"coffee-spa/usecase"
	"coffee-spa/validator"
)

func TestSearchValidator_Coverage(t *testing.T) {
	v := validator.NewSearchValidator()
	badScore := 6
	badMood := entity.Mood("bad")

	tests := []struct {
		name string
		fn   func() error
		want bool
	}{
		{name: "V-SEARCH-ADD-01 start ok", fn: func() error { return v.StartSession(usecase.StartSessionIn{Title: "morning"}) }, want: false},
		{name: "V-SEARCH-ADD-02 start title too long", fn: func() error { return v.StartSession(usecase.StartSessionIn{Title: strings.Repeat("あ", 201)}) }, want: true},
		{name: "V-SEARCH-ADD-03 set pref ok", fn: func() error { return v.SetPref(validSetPrefIn()) }, want: false},
		{name: "V-SEARCH-ADD-04 set pref bad score", fn: func() error { in := validSetPrefIn(); in.Flavor = 0; return v.SetPref(in) }, want: true},
		{name: "V-SEARCH-ADD-05 set pref bad enum", fn: func() error { in := validSetPrefIn(); in.Mood = entity.Mood("bad"); return v.SetPref(in) }, want: true},
		{name: "V-SEARCH-ADD-06 add turn ok", fn: func() error { return v.AddTurn(usecase.AddTurnIn{SessionID: 1, Body: "軽めがいい"}) }, want: false},
		{name: "V-SEARCH-ADD-07 add turn empty", fn: func() error { return v.AddTurn(usecase.AddTurnIn{SessionID: 1, Body: ""}) }, want: true},
		{name: "V-SEARCH-ADD-08 patch no fields", fn: func() error { return v.PatchPref(usecase.PatchPrefIn{SessionID: 1}) }, want: true},
		{name: "V-SEARCH-ADD-09 patch bad score", fn: func() error { return v.PatchPref(usecase.PatchPrefIn{SessionID: 1, Flavor: &badScore}) }, want: true},
		{name: "V-SEARCH-ADD-10 patch bad enum", fn: func() error { return v.PatchPref(usecase.PatchPrefIn{SessionID: 1, Mood: &badMood}) }, want: true},
		{name: "V-SEARCH-ADD-11 get session bad id", fn: func() error { return v.GetSession(usecase.GetSessionIn{SessionID: 0}) }, want: true},
		{name: "V-SEARCH-ADD-12 list history bad limit", fn: func() error { return v.ListHistory(usecase.ListHistoryIn{Limit: 0, Offset: 0}) }, want: true},
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

func validSetPrefIn() usecase.SetPrefIn {
	return usecase.SetPrefIn{
		SessionID:  1,
		Flavor:     3,
		Acidity:    3,
		Bitterness: 3,
		Body:       3,
		Aroma:      3,
		Mood:       entity.MoodRelax,
		Method:     entity.MethodDrip,
		Scene:      entity.SceneRelax,
		TempPref:   entity.TempHot,
	}
}
