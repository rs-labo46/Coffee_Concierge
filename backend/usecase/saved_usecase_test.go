package usecase

import (
	"testing"

	"coffee-spa/entity"
	"coffee-spa/usecase/port"
)

func TestSavedUsecase_Save(t *testing.T) {
	uc := NewSavedUsecase(
		savedRepoMock{createFn: func(saved *entity.SavedSuggestion) error { saved.ID = 41; return nil }},
		sessionRepoMock{
			getSessionByIDFn: func(id uint) (*entity.Session, error) { uid := uint(1); return &entity.Session{ID: id, UserID: &uid}, nil },
			getSuggestionByIDFn: func(id uint) (*entity.Suggestion, error) { return &entity.Suggestion{ID: id, SessionID: 9}, nil },
		},
		&auditRepoMock{},
		savedValMock{},
	)
	out, err := uc.Save(SaveSuggestionIn{Actor: entity.Actor{UserID: 1, Role: entity.RoleUser}, SessionID: 9, SuggestionID: 8})
	if err != nil { t.Fatalf("unexpected error:%v", err) }
	if out.ID != 41 || out.UserID != 1 { t.Fatalf("unexpected output:%+v", out) }
}

func TestSavedUsecase_ListAndDelete(t *testing.T) {
	uc := NewSavedUsecase(
		savedRepoMock{
			listFn: func(q port.SavedListQ) ([]entity.SavedSuggestion, error) { return []entity.SavedSuggestion{{UserID: q.UserID}}, nil },
			getFn: func(userID uint, suggestionID uint) (*entity.SavedSuggestion, error) { return &entity.SavedSuggestion{UserID: userID, SuggestionID: suggestionID}, nil },
		},
		sessionRepoMock{},
		&auditRepoMock{},
		savedValMock{},
	)
	list, err := uc.List(ListSavedIn{Actor: entity.Actor{UserID: 7}, Limit: 5})
	if err != nil || len(list) != 1 || list[0].UserID != 7 { t.Fatalf("unexpected list:%+v err=%v", list, err) }
	if err := uc.Delete(DeleteSavedIn{Actor: entity.Actor{UserID: 7}, SuggestionID: 3}); err != nil { t.Fatalf("unexpected delete error:%v", err) }
}
