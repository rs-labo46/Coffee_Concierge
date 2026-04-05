package usecase

import (
	"coffee-spa/entity"
	"coffee-spa/repository"
	"encoding/json"
)

// 保存ずみの提案の保存・一覧・削除
type SavedUC interface {
	Save(in SaveSuggestionIn) (entity.SavedSuggestion, error)
	List(in ListSavedIn) ([]entity.SavedSuggestion, error)
	Delete(in DeleteSavedIn) error
}

// 提案保存
type SaveSuggestionIn struct {
	Actor        entity.Actor
	SessionID    uint
	SuggestionID uint
}

// 保存一覧
type ListSavedIn struct {
	Actor  entity.Actor
	Limit  int
	Offset int
}

// 保存削除
type DeleteSavedIn struct {
	Actor        entity.Actor
	SuggestionID uint
}

// 保存バリデーター
type SavedVal interface {
	Save(in SaveSuggestionIn) error
	List(in ListSavedIn) error
	Delete(in DeleteSavedIn) error
}

type savedUsecase struct {
	saveds   repository.SavedRepository
	sessions repository.SessionRepository
	audits   repository.AuditRepository
	val      SavedVal
}

func NewSavedUsecase(
	saveds repository.SavedRepository,
	sessions repository.SessionRepository,
	audits repository.AuditRepository,
	val SavedVal,
) SavedUC {
	return &savedUsecase{
		saveds:   saveds,
		sessions: sessions,
		audits:   audits,
		val:      val,
	}
}

// suggestionを保存・認証ユーザーのみ。
func (u *savedUsecase) Save(in SaveSuggestionIn) (entity.SavedSuggestion, error) {
	if in.Actor.UserID == 0 {
		return entity.SavedSuggestion{}, repository.ErrUnauthorized
	}

	if err := u.val.Save(in); err != nil {
		return entity.SavedSuggestion{}, err
	}

	session, err := u.sessions.GetSessionByID(in.SessionID)
	if err != nil {
		return entity.SavedSuggestion{}, err
	}
	// 保存対象 session は認証ユーザー本人のものだけ許可する。
	if session.UserID == nil {
		return entity.SavedSuggestion{}, repository.ErrForbidden
	}
	if *session.UserID != in.Actor.UserID {
		return entity.SavedSuggestion{}, repository.ErrForbidden
	}

	suggestion, err := u.sessions.GetSuggestionByID(in.SuggestionID)
	if err != nil {
		return entity.SavedSuggestion{}, err
	}

	// suggestionは指定sessionに属している必要がある。
	if suggestion.SessionID != in.SessionID {
		return entity.SavedSuggestion{}, repository.ErrConflict
	}

	saved := &entity.SavedSuggestion{
		UserID:       in.Actor.UserID,
		SessionID:    in.SessionID,
		SuggestionID: in.SuggestionID,
	}

	if err := u.saveds.Create(saved); err != nil {
		return entity.SavedSuggestion{}, err
	}

	u.writeAudit(
		"ai.suggestion.save",
		&in.Actor.UserID,
		map[string]string{
			"user_id":       uintToStr(in.Actor.UserID),
			"session_id":    uintToStr(in.SessionID),
			"suggestion_id": uintToStr(in.SuggestionID),
		},
	)

	return *saved, nil
}

// 認証ユーザーの保存済み提案一覧を返す。
func (u *savedUsecase) List(in ListSavedIn) ([]entity.SavedSuggestion, error) {
	if in.Actor.UserID == 0 {
		return nil, repository.ErrUnauthorized
	}

	if err := u.val.List(in); err != nil {
		return nil, err
	}

	out, err := u.saveds.List(repository.SavedListQ{
		UserID: in.Actor.UserID,
		Limit:  in.Limit,
		Offset: in.Offset,
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

// 認証ユーザーの保存済み提案を削除する。
// suggestion_idをキーにしつつ、user_id条件で本人データだけ削除する。
func (u *savedUsecase) Delete(in DeleteSavedIn) error {
	if in.Actor.UserID == 0 {
		return repository.ErrUnauthorized
	}

	if err := u.val.Delete(in); err != nil {
		return err
	}

	// 先に存在確認して、削除対象なしをnot foundとして返す。
	_, err := u.saveds.GetByUserAndSuggestionID(in.Actor.UserID, in.SuggestionID)
	if err != nil {
		return err
	}

	if err := u.saveds.DeleteByUserAndSuggestionID(in.Actor.UserID, in.SuggestionID); err != nil {
		return err
	}

	return nil
}

func (u *savedUsecase) writeAudit(
	typ string,
	userID *uint,
	meta map[string]string,
) {
	if u.audits == nil {
		return
	}

	raw, err := json.Marshal(meta)
	if err != nil {
		raw = []byte(`{}`)
	}

	_ = u.audits.Create(&entity.AuditLog{
		Type:   typ,
		UserID: userID,
		IP:     "",
		UA:     "",
		Meta:   raw,
	})
}
