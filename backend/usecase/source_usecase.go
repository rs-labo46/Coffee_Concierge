package usecase

import (
	"encoding/json"

	"coffee-spa/entity"
	"coffee-spa/repository"
)

// Sourceの作成・取得・一覧。
type SourceUC interface {
	Create(actor entity.Actor, in CreateSourceIn) (entity.Source, error)
	Get(id uint) (entity.Source, error)
	List(limit int, offset int) ([]entity.Source, error)
}

// Source作成入力。
type CreateSourceIn struct {
	Name    string
	SiteURL string
}

// Source系の、validator。
type SourceVal interface {
	Create(in CreateSourceIn) error
	Get(id uint) error
	List(limit int, offset int) error
}

type sourceUsecase struct {
	sources repository.SourceRepository
	audits  repository.AuditRepository
	val     SourceVal
}

func NewSourceUsecase(
	sources repository.SourceRepository,
	audits repository.AuditRepository,
	val SourceVal,
) SourceUC {
	return &sourceUsecase{
		sources: sources,
		audits:  audits,
		val:     val,
	}
}

// Sourceを新規作成。
// 管理操作。adminのみ許可する。
func (u *sourceUsecase) Create(actor entity.Actor, in CreateSourceIn) (entity.Source, error) {
	if actor.Role != entity.RoleAdmin {
		return entity.Source{}, repository.ErrForbidden
	}

	if err := u.val.Create(in); err != nil {
		return entity.Source{}, err
	}

	src := &entity.Source{
		Name:    in.Name,
		SiteURL: in.SiteURL,
	}

	if err := u.sources.Create(src); err != nil {
		return entity.Source{}, err
	}

	u.writeAudit(
		"admin.sources.create",
		&actor.UserID,
		map[string]string{
			"source_id": uintToStr(src.ID),
			"name":      src.Name,
		},
	)

	return *src, nil
}

// Sourceを1件取得する。
func (u *sourceUsecase) Get(id uint) (entity.Source, error) {
	if err := u.val.Get(id); err != nil {
		return entity.Source{}, err
	}

	src, err := u.sources.GetByID(id)
	if err != nil {
		return entity.Source{}, err
	}

	return *src, nil
}

// Source一覧を返す。
func (u *sourceUsecase) List(limit int, offset int) ([]entity.Source, error) {
	if err := u.val.List(limit, offset); err != nil {
		return nil, err
	}

	out, err := u.sources.List(repository.SourceListQ{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

// 監査ログ保存。
func (u *sourceUsecase) writeAudit(
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
