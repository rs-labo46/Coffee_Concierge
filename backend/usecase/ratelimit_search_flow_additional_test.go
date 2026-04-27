package usecase_test

import (
	"errors"
	"testing"
	"time"

	"coffee-spa/entity"
	"coffee-spa/testutil/repomock"
	"coffee-spa/testutil/servicemock"
	"coffee-spa/testutil/valmock"
	"coffee-spa/usecase"
)

func TestRateLimitUsecase_AdditionalCoverage(t *testing.T) {
	// U-RL-ADD-01
	t.Run("AllowLoginIP denied returns retry after", func(t *testing.T) {
		store := &repomock.RateLimit{AllowFn: func(key string, rate float64, capacity float64, cost float64, now time.Time) (bool, int, error) {
			return false, 12, nil
		}}
		uc := usecase.NewRateLimitUC(store, servicemock.Clock{})
		allowed, retryAfter, err := uc.AllowLoginIP("127.0.0.1")
		if err != nil {
			t.Fatalf("AllowLoginIP() error = %v", err)
		}
		if allowed || retryAfter != 12 {
			t.Fatalf("allowed/retryAfter = %v/%d, want false/12", allowed, retryAfter)
		}
	})

	// U-RL-ADD-02
	t.Run("AllowRefreshToken returns store error", func(t *testing.T) {
		want := errors.New("redis down")
		store := &repomock.RateLimit{AllowFn: func(key string, rate float64, capacity float64, cost float64, now time.Time) (bool, int, error) {
			return false, 0, want
		}}
		uc := usecase.NewRateLimitUC(store, servicemock.Clock{})
		_, _, err := uc.AllowRefreshToken("hash")
		if !errors.Is(err, want) {
			t.Fatalf("AllowRefreshToken() error = %v, want %v", err, want)
		}
	})
}

func TestSearchFlowUsecase_AdditionalCoverage(t *testing.T) {
	// U-SF-ADD-01
	t.Run("SetPref rejects existing pref", func(t *testing.T) {
		owner := uint(1)
		sessions := &repomock.Session{GetSessionByIDFn: func(id uint) (*entity.Session, error) {
			return &entity.Session{ID: id, UserID: &owner, Status: entity.SessionActive}, nil
		}}
		uc := newSearchFlowUCTest(sessions, valmock.Search{})
		_, err := uc.SetPref(validSetPrefUCIn(&entity.Actor{UserID: 1}))
		if !errors.Is(err, usecase.ErrConflict) {
			t.Fatalf("SetPref() error = %v, want conflict", err)
		}
	})

	// U-SF-ADD-02
	t.Run("AddTurn rejects closed session", func(t *testing.T) {
		owner := uint(1)
		sessions := &repomock.Session{GetSessionByIDFn: func(id uint) (*entity.Session, error) {
			return &entity.Session{ID: id, UserID: &owner, Status: entity.SessionClosed}, nil
		}}
		uc := newSearchFlowUCTest(sessions, valmock.Search{})
		_, err := uc.AddTurn(usecase.AddTurnIn{Actor: &entity.Actor{UserID: 1}, SessionID: 1, Body: "軽め"})
		if !errors.Is(err, usecase.ErrConflict) {
			t.Fatalf("AddTurn() error = %v, want conflict", err)
		}
	})

	// U-SF-ADD-03
	t.Run("PatchPref returns validator error", func(t *testing.T) {
		want := errors.New("invalid patch")
		uc := newSearchFlowUCTest(&repomock.Session{}, valmock.Search{Err: want})
		flavor := 4
		_, err := uc.PatchPref(usecase.PatchPrefIn{Actor: &entity.Actor{UserID: 1}, SessionID: 1, Flavor: &flavor})
		if !errors.Is(err, want) {
			t.Fatalf("PatchPref() error = %v, want %v", err, want)
		}
	})
}

func newSearchFlowUCTest(sessions *repomock.Session, val usecase.SearchVal) usecase.SearchFlowUC {
	return usecase.NewSearchFlowUsecase(
		sessions,
		&repomock.Bean{},
		&repomock.Recipe{},
		&repomock.Item{},
		&repomock.Audit{},
		val,
		servicemock.Ranker{},
		servicemock.Gemini{},
		servicemock.Clock{},
		servicemock.IDGen{},
		24*time.Hour,
	)
}

func validSetPrefUCIn(actor *entity.Actor) usecase.SetPrefIn {
	return usecase.SetPrefIn{Actor: actor, SessionID: 1, Flavor: 3, Acidity: 3, Bitterness: 3, Body: 3, Aroma: 3, Mood: entity.MoodRelax, Method: entity.MethodDrip, Scene: entity.SceneRelax, TempPref: entity.TempHot}
}
