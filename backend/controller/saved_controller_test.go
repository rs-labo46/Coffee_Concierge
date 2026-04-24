package controller

import (
	"net/http"
	"testing"

	"coffee-spa/entity"
	"coffee-spa/usecase"
)

func TestSavedCtlSaveCreated(t *testing.T) {
	uc := &savedUCMock{saveFn: func(in usecase.SaveSuggestionIn) (entity.SavedSuggestion, error) {
		if in.Actor.UserID != 7 || in.SessionID != 3 || in.SuggestionID != 9 { t.Fatalf("unexpected input: %+v", in) }
		return entity.SavedSuggestion{ID: 1, UserID: in.Actor.UserID, SessionID: in.SessionID, SuggestionID: in.SuggestionID}, nil
	}}
	ctl := NewSavedCtl(uc)
	c, rec, err := newJSONContext(http.MethodPost, "/saved-suggestions", map[string]uint{"session_id": 3, "suggestion_id": 9})
	if err != nil { t.Fatal(err) }
	setActor(c, &entity.Actor{UserID: 7, Role: entity.RoleUser})
	if err := ctl.Save(c); err != nil { t.Fatal(err) }
	if rec.Code != http.StatusCreated { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
}

func TestSavedCtlDeleteBadID(t *testing.T) {
	uc := &savedUCMock{deleteFn: func(usecase.DeleteSavedIn) error { t.Fatal("should not be called"); return nil }}
	ctl := NewSavedCtl(uc)
	c, rec, err := newJSONContext(http.MethodDelete, "/saved-suggestions/x", nil)
	if err != nil { t.Fatal(err) }
	c.SetParamNames("suggestionId")
	c.SetParamValues("x")
	setActor(c, &entity.Actor{UserID: 7, Role: entity.RoleUser})
	if err := ctl.Delete(c); err != nil { t.Fatal(err) }
	if rec.Code != http.StatusBadRequest { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
}
