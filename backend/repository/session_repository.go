package repository

import (
	"errors"
	"time"

	"coffee-spa/apperr"
	"coffee-spa/entity"
	"coffee-spa/usecase/port"

	"gorm.io/gorm"
)

// session / turn / pref / suggestionの永続化。
type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) port.SessionRepository {
	return &sessionRepository{
		db: db,
	}
}

// sessionを新規作成。
func (r *sessionRepository) CreateSession(session *entity.Session) error {
	if session == nil {
		return apperr.ErrInvalidState
	}

	// レコードをINSERT。
	err := r.db.Create(session).Error
	if err != nil {
		// unique / FK 制約違反はconflict。
		if isDup(err) || isFK(err) {
			return apperr.ErrConflict
		}
		return apperr.ErrInternal
	}

	return nil
}

// IDでsessionを取得。
func (r *sessionRepository) GetSessionByID(id uint) (*entity.Session, error) {
	// 0 は有効な ID ではありません。
	if id == 0 {
		return nil, apperr.ErrNotFound
	}

	var session entity.Session

	// 主キー検索で取得。
	err := r.db.First(&session, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.ErrInternal
	}

	return &session, nil
}

// guest sessionを取得。
// user_idがnullであること、session_key_hashが一致すること、期限内であることを条件。
func (r *sessionRepository) GetGuestSessionByID(
	id uint,
	sessionKeyHash string,
	now time.Time,
) (*entity.Session, error) {
	// IDまたはsessionKeyHashが不正なら対象なし。
	if id == 0 || sessionKeyHash == "" {
		return nil, apperr.ErrNotFound
	}

	var session entity.Session

	// guestの再開条件を満たす、sessionだけを取得。
	err := r.db.
		Where("id = ?", id).
		Where("user_id IS NULL").
		Where("session_key_hash = ?", sessionKeyHash).
		Where("guest_expires_at IS NOT NULL").
		Where("guest_expires_at > ?", now).
		First(&session).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.ErrInternal
	}

	return &session, nil
}

// 認証ユーザーの session履歴一覧を返す。
func (r *sessionRepository) ListHistory(q port.HistoryQ) ([]entity.Session, error) {
	// userID がなければ認証ユーザーの履歴として扱えません。
	if q.UserID == 0 {
		return nil, apperr.ErrUnauthorized
	}

	// limitは未指定なら20、上限は100に。
	limit := q.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// offsetはマイナスを許可しない。
	offset := q.Offset
	if offset < 0 {
		offset = 0
	}

	var sessions []entity.Session

	// 新しい履歴から見やすいよう created_at DESCで返す。
	err := r.db.
		Model(&entity.Session{}).
		Where("user_id = ?", q.UserID).
		Order("created_at DESC").
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&sessions).
		Error
	if err != nil {
		return nil, apperr.ErrInternal
	}

	return sessions, nil
}

// sessionをclosedに更新。
func (r *sessionRepository) CloseSession(id uint) error {
	if id == 0 {
		return apperr.ErrNotFound
	}

	now := time.Now()

	// 更新対象の値だけを持つ。
	session := entity.Session{
		Status:    entity.SessionClosed,
		UpdatedAt: now,
	}

	// statusとupdated_atだけを更新。
	res := r.db.
		Model(&entity.Session{}).
		Where("id = ?", id).
		Select("status", "updated_at").
		Updates(&session)
	if res.Error != nil {
		return apperr.ErrInternal
	}
	if res.RowsAffected == 0 {
		return apperr.ErrNotFound
	}

	return nil
}

// turnを新規作成。
func (r *sessionRepository) CreateTurn(turn *entity.Turn) error {
	if turn == nil {
		return apperr.ErrInvalidState
	}

	// レコードをINSERT 。
	err := r.db.Create(turn).Error
	if err != nil {
		// FK制約違反はsession不整合のため、conflict。
		if isFK(err) {
			return apperr.ErrConflict
		}
		return apperr.ErrInternal
	}

	return nil
}

// session配下のturn一覧を時系列順で返す。
func (r *sessionRepository) ListTurns(sessionID uint) ([]entity.Turn, error) {
	if sessionID == 0 {
		return nil, apperr.ErrNotFound
	}

	var turns []entity.Turn

	// 作成順に並ぶようcreated_at ASC → id ASCで返す。
	err := r.db.
		Model(&entity.Turn{}).
		Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Order("id ASC").
		Find(&turns).
		Error
	if err != nil {
		return nil, apperr.ErrInternal
	}

	return turns, nil
}

// prefを新規作成。
func (r *sessionRepository) CreatePref(pref *entity.Pref) error {
	if pref == nil {
		return apperr.ErrInvalidState
	}

	// レコードをINSERT 。
	err := r.db.Create(pref).Error
	if err != nil {
		// unique / FK 制約違反はconflict。
		if isDup(err) || isFK(err) {
			return apperr.ErrConflict
		}
		return apperr.ErrInternal
	}

	return nil
}

// 既存のprefを更新。
func (r *sessionRepository) UpdatePref(pref *entity.Pref) error {

	if pref == nil || pref.ID == 0 {
		return apperr.ErrInvalidState
	}

	// 更新可能カラムだけを明示して更新。
	err := r.db.
		Model(&entity.Pref{}).
		Where("id = ?", pref.ID).
		Select(
			"session_id",
			"flavor",
			"acidity",
			"bitterness",
			"body",
			"aroma",
			"mood",
			"method",
			"scene",
			"temp_pref",
			"excludes",
			"note",
			"updated_at",
		).
		Updates(pref).
		Error
	if err != nil {
		if isDup(err) || isFK(err) {
			return apperr.ErrConflict
		}
		return apperr.ErrInternal
	}

	return nil
}

// sessionに紐づくprefを1件取得。
func (r *sessionRepository) GetPrefBySessionID(sessionID uint) (*entity.Pref, error) {
	if sessionID == 0 {
		return nil, apperr.ErrNotFound
	}

	var pref entity.Pref

	// 1件取得。
	err := r.db.
		Where("session_id = ?", sessionID).
		First(&pref).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.ErrInternal
	}

	return &pref, nil
}

// 検索結果のスナップショット更新(session配下のsuggestionsを全置換)。
func (r *sessionRepository) ReplaceSuggestions(
	sessionID uint,
	suggestions []entity.Suggestion,
) error {
	if sessionID == 0 {
		return apperr.ErrInvalidState
	}

	// delete → insertは途中失敗すると壊れるためtransactionで。
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 既存のsuggestionsをsession単位で削除。
		if err := tx.
			Where("session_id = ?", sessionID).
			Delete(&entity.Suggestion{}).
			Error; err != nil {
			return apperr.ErrInternal
		}

		// 新しいsuggestionsが空なら削除だけで終了。
		if len(suggestions) == 0 {
			return nil
		}

		// 再作成のため、IDを０に戻し、sessionIDを揃える。
		for i := range suggestions {
			suggestions[i].ID = 0
			suggestions[i].SessionID = sessionID
		}

		// 新しいsuggestionsを一括でINSERT 。
		if err := tx.Create(&suggestions).Error; err != nil {
			if isDup(err) || isFK(err) {
				return apperr.ErrConflict
			}
			return apperr.ErrInternal
		}

		return nil
	})
}

// sessionに紐づくsuggestion一覧を順位順で返す。
func (r *sessionRepository) ListSuggestions(sessionID uint) ([]entity.Suggestion, error) {
	if sessionID == 0 {
		return nil, apperr.ErrNotFound
	}

	var suggestions []entity.Suggestion

	// 表示や詳細でそのまま使いやすいように関連をpreload 。
	err := r.db.
		Preload("Bean").
		Preload("Recipe").
		Preload("Item").
		Preload("Item.Source").
		Where("session_id = ?", sessionID).
		Order("rank ASC").
		Order("id ASC").
		Find(&suggestions).
		Error
	if err != nil {
		return nil, apperr.ErrInternal
	}

	return suggestions, nil
}

// suggestionを1件取得。
func (r *sessionRepository) GetSuggestionByID(id uint) (*entity.Suggestion, error) {
	if id == 0 {
		return nil, apperr.ErrNotFound
	}

	var suggestion entity.Suggestion

	// 関連先読み取りした状態で取得。
	err := r.db.
		Preload("Bean").
		Preload("Recipe").
		Preload("Item").
		Preload("Item.Source").
		First(&suggestion, id).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, apperr.ErrInternal
	}

	return &suggestion, nil
}
