package repository

import (
	"errors"

	"coffee-spa/entity"
)

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")
var ErrInternal = errors.New("internal")

type ItemQ struct {
	Q      string
	Kind   string
	Limit  int
	Offset int
}

type TopItems struct {
	News   []entity.Item
	Recipe []entity.Item
	Deal   []entity.Item
	Shop   []entity.Item
}

type UserRepository interface {
	Create(u entity.User) (entity.User, error)
	GetByEmail(email string) (entity.User, error)
	GetByID(id int64) (entity.User, error)
	SetEmailVerified(userID int64) error
	UpdatePassHash(userID int64, newHash string) error
	BumpTokenVer(userID int64) (int, error)
}

type EvRepository interface {
	Create(ev entity.EmailVerify) error
	GetByTokenHash(hash string) (entity.EmailVerify, error)
	Use(id int64) error
	RevokeUnusedByUser(userID int64) error
}

type PwRepository interface {
	Create(pw entity.PwReset) error
	GetByTokenHash(hash string) (entity.PwReset, error)
	Use(id int64) error
	RevokeUnusedByUser(userID int64) error
}

type RtRepository interface {
	Create(rt entity.RefreshToken) (entity.RefreshToken, error)
	GetByTokenHash(hash string) (entity.RefreshToken, error)
	Revoke(id int64) error
	MarkUsed(id int64) error
	SetReplacedBy(id int64, newID int64) error
	RevokeByFamilyID(familyID string) error
	RevokeAllByUser(userID int64) error
}

type SourceRepository interface {
	Create(s entity.Source) (entity.Source, error)
	List() ([]entity.Source, error)
}

type ItemRepository interface {
	Create(i entity.Item) (entity.Item, error)
	GetByID(id int64) (entity.Item, error)
	List(q ItemQ) ([]entity.Item, error)
	Top(cap int) (TopItems, error)
}

type AuditRepository interface {
	Create(a entity.AuditLog) error
}
