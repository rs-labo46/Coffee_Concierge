package usecase

import (
	"coffee-spa/entity"
	"coffee-spa/usecase/port"
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
	audits port.AuditRepository
	val    AuditVal
}

func NewAuditUsecase(
	audits port.AuditRepository,
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
		return nil, ErrForbidden
	}

	if err := u.val.List(in); err != nil {
		return nil, err
	}

	out, err := u.audits.List(port.AuditListQ{
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
