package usecase

import (
	"encoding/json"
	"strings"

	"coffee-spa/entity"
	"coffee-spa/repository"

	"gorm.io/datatypes"
)

type SourceVal interface {
	NewSource(input AddSourceIn) error
}

type SourceUC struct {
	source repository.SourceRepository
	audit  repository.AuditRepository
	val    SourceVal
}

type sourceCreateMeta struct {
	UserID   int64 `json:"user_id"`
	SourceID int64 `json:"source_id"`
}

func NewSourceUC(
	source repository.SourceRepository,
	audit repository.AuditRepository,
	val SourceVal,
) SourceUsecase {
	return &SourceUC{
		source: source,
		audit:  audit,
		val:    val,
	}
}

func (u *SourceUC) Add(actor Actor, sourceInput AddSourceIn) (entity.Source, error) {
	if err := u.val.NewSource(sourceInput); err != nil {
		return entity.Source{}, ErrInvalidRequest
	}

	normalized := normalizeSourceInput(sourceInput)

	src, err := u.source.Create(entity.Source{
		Name:    normalized.Name,
		SiteURL: normalized.SiteURL,
	})
	if err != nil {
		return entity.Source{}, mapRepoErr(err)
	}

	metaBytes, err := json.Marshal(sourceCreateMeta{
		UserID:   actor.UserID,
		SourceID: src.ID,
	})
	if err != nil {
		return entity.Source{}, ErrInternal
	}

	auditErr := u.audit.Create(entity.AuditLog{
		Type:     "admin.sources.create",
		UserID:   int64Pointer(actor.UserID),
		IP:       actor.IP,
		UA:       actor.UA,
		MetaJSON: datatypes.JSON(metaBytes),
	})
	if auditErr != nil {
		return entity.Source{}, mapRepoErr(auditErr)
	}

	return src, nil
}

func (u *SourceUC) List() ([]entity.Source, error) {
	sources, err := u.source.List()
	if err != nil {
		return nil, mapRepoErr(err)
	}

	return sources, nil
}

func normalizeSourceInput(input AddSourceIn) AddSourceIn {
	input.Name = strings.TrimSpace(input.Name)
	input.SiteURL = trimNullableString(input.SiteURL)
	return input
}
