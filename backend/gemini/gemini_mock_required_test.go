package gemini_test

import (
	"errors"
	"testing"

	"coffee-spa/entity"
	"coffee-spa/gemini"
	"coffee-spa/usecase"
)

func basePref() entity.Pref {
	return entity.Pref{Flavor: 3, Acidity: 3, Bitterness: 3, Body: 3, Aroma: 3, Mood: entity.MoodRelax, Method: entity.MethodDrip, Scene: entity.SceneRelax, TempPref: entity.TempHot}
}

func TestGeminiMockClient_RequiredCoverage(t *testing.T) {
	client := gemini.NewMockClient()
	t.Run("empty input fails with audit meta", func(t *testing.T) {
		_, meta, err := client.BuildConditionDiff(usecase.GeminiConditionDiffIn{InputText: "   ", Pref: basePref()})
		if !errors.Is(err, gemini.ErrEmptyInput) {
			t.Fatalf("err=%v", err)
		}
		if meta.Status == "" || meta.ErrorType == "" {
			t.Fatalf("meta=%#v", meta)
		}
	})

	t.Run("light body request lowers body within range", func(t *testing.T) {
		out, _, err := client.BuildConditionDiff(usecase.GeminiConditionDiffIn{InputText: "もう少し軽め", Pref: basePref()})
		if err != nil {
			t.Fatalf("err=%v", err)
		}
		if out.Body == nil || *out.Body != 2 {
			t.Fatalf("body=%v", out.Body)
		}
	})

	t.Run("milk request sets method", func(t *testing.T) {
		out, _, err := client.BuildConditionDiff(usecase.GeminiConditionDiffIn{InputText: "ミルクに合うやつ", Pref: basePref()})
		if err != nil {
			t.Fatalf("err=%v", err)
		}
		if out.Method == nil || *out.Method != entity.MethodMilk {
			t.Fatalf("method=%v", out.Method)
		}
	})

	t.Run("reasons are generated for suggestions", func(t *testing.T) {
		reasons, meta, err := client.BuildReasons(usecase.GeminiReasonIn{Pref: basePref(), Suggestions: []entity.Suggestion{{ID: 1, BeanID: 1, Rank: 1}}, Beans: []entity.Bean{{ID: 1, Name: "Bean"}}})
		if err != nil {
			t.Fatalf("err=%v", err)
		}
		if meta.Status == "" {
			t.Fatalf("meta=%#v", meta)
		}
		if len(reasons) == 0 || reasons[0].Reason == "" {
			t.Fatalf("reasons=%#v", reasons)
		}
	})

	t.Run("followups are capped and non-empty", func(t *testing.T) {
		qs, _, err := client.BuildFollowups(usecase.GeminiFollowupIn{Pref: basePref(), Beans: []entity.Bean{{ID: 1, Name: "Bean"}}})
		if err != nil {
			t.Fatalf("err=%v", err)
		}
		if len(qs) == 0 || len(qs) > 3 {
			t.Fatalf("questions=%#v", qs)
		}
	})
}

func TestGeminiServiceConstructor_RequiredCoverage(t *testing.T) {
	if _, err := gemini.NewService("", "gemini-2.5-flash"); err == nil {
		t.Fatal("empty api key should fail")
	}
	if _, err := gemini.NewService("key", ""); err == nil {
		t.Fatal("empty model should fail")
	}
	if svc, err := gemini.NewService("key", "gemini-2.5-flash"); err != nil || svc == nil {
		t.Fatalf("svc=%v err=%v", svc, err)
	}
}
