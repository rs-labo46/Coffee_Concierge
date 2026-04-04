package usecase

import (
	"coffee-spa/entity"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	"gorm.io/datatypes"
)

func (u *AuthUC) writeAudit(
	typ string,
	userID *int64,
	ip string,
	ua string,
	metaSrc interface{},
) error {
	meta := datatypes.JSON([]byte("{}"))

	if metaSrc != nil {
		b, err := json.Marshal(metaSrc)
		if err != nil {
			return ErrInternal
		}
		meta = datatypes.JSON(b)
	}

	err := u.audit.Create(entity.AuditLog{
		Type:     typ,
		UserID:   userID,
		IP:       ip,
		UA:       ua,
		MetaJSON: meta,
	})
	if err != nil {
		return mapRepoErr(err)
	}

	return nil
}

func normEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func maskEmail(email string) string {
	at := strings.Index(email, "@")
	if at <= 1 {
		return "***"
	}

	name := email[:at]
	domain := email[at:]

	if len(name) <= 2 {
		return name[:1] + "***" + domain
	}

	return name[:1] + "***" + name[len(name)-1:] + domain
}
