package usecase

import (
	"coffee-spa/entity"
	"coffee-spa/repository"
)

// 監査一覧取得。
type AuditUC interface {
	List(actor entity.Actor, in AuditListIn) ([]entity.AuditLog, error)
}

// 監査一覧。
type AuditListIn struct {
	Type   string
	UserID *uint
	Limit  int
	Offset int
}

type auditUsecase struct {
	audits repository.AuditRepository
	val    AuditVal
}

func NewAuditUsecase(
	audits repository.AuditRepository,
	val AuditVal,
) AuditUC {
	return &auditUsecase{
		audits: audits,
		val:    val,
	}
}

// 監査一覧はadminのみ許可する。
func (u *auditUsecase) List(actor entity.Actor, in AuditListIn) ([]entity.AuditLog, error) {
	if actor.Role != entity.RoleAdmin {
		return nil, repository.ErrForbidden
	}

	if err := u.val.List(in); err != nil {
		return nil, err
	}

	out, err := u.audits.List(repository.AuditListQ{
		Type:   in.Type,
		UserID: in.UserID,
		Limit:  in.Limit,
		Offset: in.Offset,
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}
