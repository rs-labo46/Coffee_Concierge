package usecase

import (
	"coffee-spa/entity"
	"coffee-spa/repository"
	"encoding/json"
)

// セッション詳細取得、履歴一覧取得、セッション終了を担当する。
type SessionUC interface {
	GetSession(in GetSessionIn) (GetSessionOut, error)
	ListHistory(in ListHistoryIn) ([]entity.Session, error)
	CloseSession(in CloseSessionIn) error
}

// セッション詳細取得入力。
// 認証ユーザーはActorで本人確認し、guestはSessionKeyで確認をする。
type GetSessionIn struct {
	Actor      *entity.Actor
	SessionID  uint
	SessionKey string
}

// セッション詳細取得結果。
type GetSessionOut struct {
	Session     entity.Session
	Turns       []entity.Turn
	Pref        entity.Pref
	Suggestions []entity.Suggestion
}

// 履歴一覧取得入力。
// 履歴一覧は認証ユーザー専用。
type ListHistoryIn struct {
	Actor  entity.Actor
	Limit  int
	Offset int
}

// セッション終了入力。
// 認証ユーザーはActor、guestはSessionKeyで終了可否を判定。
type CloseSessionIn struct {
	Actor      *entity.Actor
	SessionID  uint
	SessionKey string
}

type sessionUsecase struct {
	sessions repository.SessionRepository
	audits   repository.AuditRepository
	val      SearchVal
	clock    Clock
}

func NewSessionUsecase(
	sessions repository.SessionRepository,
	audits repository.AuditRepository,
	val SearchVal,
	clock Clock,
) SessionUC {
	return &sessionUsecase{
		sessions: sessions,
		audits:   audits,
		val:      val,
		clock:    clock,
	}
}

// セッション詳細を返す。
// 認証ユーザーは本人のsessionのみで、guestはsession_id + sessionKey一致のみ許可する。
func (u *sessionUsecase) GetSession(in GetSessionIn) (GetSessionOut, error) {
	if err := u.val.GetSession(in); err != nil {
		return GetSessionOut{}, err
	}

	session, err := u.resolveReadableSession(in.Actor, in.SessionID, in.SessionKey)
	if err != nil {
		return GetSessionOut{}, err
	}

	turns, err := u.sessions.ListTurns(session.ID)
	if err != nil {
		return GetSessionOut{}, err
	}

	pref, err := u.sessions.GetPrefBySessionID(session.ID)
	if err != nil {
		return GetSessionOut{}, err
	}

	suggestions, err := u.sessions.ListSuggestions(session.ID)
	if err != nil {
		return GetSessionOut{}, err
	}

	return GetSessionOut{
		Session:     *session,
		Turns:       turns,
		Pref:        *pref,
		Suggestions: suggestions,
	}, nil
}

// 認証ユーザー本人の履歴一覧を返す。
// guestは履歴一覧を使えない。
func (u *sessionUsecase) ListHistory(in ListHistoryIn) ([]entity.Session, error) {
	if in.Actor.UserID == 0 {
		return nil, ErrUnauthorized
	}

	if err := u.val.ListHistory(in); err != nil {
		return nil, err
	}

	out, err := u.sessions.ListHistory(repository.HistoryQ{
		UserID: in.Actor.UserID,
		Limit:  in.Limit,
		Offset: in.Offset,
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

// セッションを終了する。
// 認証ユーザーは本人sessionのみで、guestはsession_id + sessionKey一致時のみ終了できる。
func (u *sessionUsecase) CloseSession(in CloseSessionIn) error {
	if err := u.val.CloseSession(in); err != nil {
		return err
	}

	session, err := u.resolveWritableSession(in.Actor, in.SessionID, in.SessionKey)
	if err != nil {
		return err
	}

	if session.Status != entity.SessionActive {
		return ErrConflict
	}

	if err := u.sessions.CloseSession(session.ID); err != nil {
		return err
	}

	u.writeAudit(
		"ai.session.close",
		userIDPtr(in.Actor),
		map[string]string{
			"session_id": uintToStr(session.ID),
			"mode":       actorMode(in.Actor),
		},
	)

	return nil
}

// 読み取り可能なsession。
// 認証ユーザーは本人所有、guestは、sessionKey一致かつ期限内のみ許可する。
func (u *sessionUsecase) resolveReadableSession(
	actor *entity.Actor,
	sessionID uint,
	sessionKey string,
) (*entity.Session, error) {
	if actor != nil && actor.UserID > 0 {
		session, err := u.sessions.GetSessionByID(sessionID)
		if err != nil {
			return nil, err
		}
		if session.UserID == nil {
			return nil, ErrForbidden
		}
		if *session.UserID != actor.UserID {
			return nil, ErrForbidden
		}
		return session, nil
	}

	if sessionKey == "" {
		return nil, ErrUnauthorized
	}

	return u.sessions.GetGuestSessionByID(
		sessionID,
		hashText(sessionKey),
		u.clock.Now(),
	)
}

// 書き込み可能なsession。
// 認証ユーザーは本人所有のみで、guestはsessionKey一致かつ期限内のみ許可する。
func (u *sessionUsecase) resolveWritableSession(
	actor *entity.Actor,
	sessionID uint,
	sessionKey string,
) (*entity.Session, error) {
	if actor != nil && actor.UserID > 0 {
		session, err := u.sessions.GetSessionByID(sessionID)
		if err != nil {
			return nil, err
		}
		if session.UserID == nil {
			return nil, ErrForbidden
		}
		if *session.UserID != actor.UserID {
			return nil, ErrForbidden
		}
		return session, nil
	}

	if sessionKey == "" {
		return nil, ErrUnauthorized
	}

	return u.sessions.GetGuestSessionByID(
		sessionID,
		hashText(sessionKey),
		u.clock.Now(),
	)
}

func (u *sessionUsecase) writeAudit(
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
