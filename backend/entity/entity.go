package entity

import (
	"time"

	"gorm.io/datatypes"
)

// user.roleの値
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// item.kindの値
type ItemKind string

const (
	KindNews   ItemKind = "news"   // ニュース
	KindRecipe ItemKind = "recipe" // レシピ
	KindDeal   ItemKind = "deal"   // セール
	KindShop   ItemKind = "shop"   // 店舗ショップ
)

// Userはusersテーブル
type User struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Email         string    `gorm:"column:email;type:varchar;not null;uniqueIndex" json:"email"`
	PassHash      string    `gorm:"column:pass_hash;type:varchar;not null" json:"-"`
	Role          string    `gorm:"column:role;type:varchar;not null" json:"role"`
	TokenVer      int       `gorm:"column:token_ver;not null;default:1" json:"token_ver"`
	EmailVerified bool      `gorm:"column:email_verified;not null;default:false" json:"email_verified"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`
}

// EmailVerifyはemail_verifiesテーブル
type EmailVerify struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    int64      `gorm:"column:user_id;not null;index" json:"user_id"`
	User      User       `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	TokenHash string     `gorm:"column:token_hash;type:varchar;not null;uniqueIndex" json:"-"`
	ExpiresAt time.Time  `gorm:"column:expires_at;not null" json:"expires_at"`
	UsedAt    *time.Time `gorm:"column:used_at" json:"used_at"`
}

// PwResetはpw_resetsテーブル
type PwReset struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    int64      `gorm:"column:user_id;not null;index" json:"user_id"`
	User      User       `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	TokenHash string     `gorm:"column:token_hash;type:varchar;not null;uniqueIndex" json:"-"`
	ExpiresAt time.Time  `gorm:"column:expires_at;not null" json:"expires_at"`
	UsedAt    *time.Time `gorm:"column:used_at" json:"used_at"`
	CreatedAt time.Time  `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
}

// RefreshTokenはrefresh_tokensテーブル
type RefreshToken struct {
	ID           int64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID       int64         `gorm:"column:user_id;not null;index" json:"user_id"`
	User         User          `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	FamilyID     string        `gorm:"column:family_id;type:varchar;not null;index" json:"family_id"`
	TokenHash    string        `gorm:"column:token_hash;type:varchar;not null;uniqueIndex" json:"-"`
	ExpiresAt    time.Time     `gorm:"column:expires_at;not null" json:"expires_at"`
	RevokedAt    *time.Time    `gorm:"column:revoked_at" json:"revoked_at"`
	UsedAt       *time.Time    `gorm:"column:used_at" json:"used_at"`
	ReplacedByID *int64        `gorm:"column:replaced_by_id" json:"replaced_by_id"`
	ReplacedBy   *RefreshToken `gorm:"foreignKey:ReplacedByID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	CreatedAt    time.Time     `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
}

// Sourceはsourcesテーブル
type Source struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name;type:varchar;not null;uniqueIndex" json:"name"`
	SiteURL   *string   `gorm:"column:site_url;type:varchar" json:"site_url"`
	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
}

// Itemはitemsテーブル
type Item struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Title       string    `gorm:"column:title;type:varchar;not null" json:"title"`
	Summary     *string   `gorm:"column:summary;type:varchar" json:"summary"`
	Body        *string   `gorm:"column:body;type:text" json:"body"`
	URL         *string   `gorm:"column:url;type:varchar" json:"url"`
	ImageURL    *string   `gorm:"column:image_url;type:varchar" json:"image_url"`
	Kind        string    `gorm:"column:kind;type:varchar;not null;index" json:"kind"`
	SourceID    int64     `gorm:"column:source_id;not null;index" json:"source_id"`
	Source      Source    `gorm:"foreignKey:SourceID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
	PublishedAt time.Time `gorm:"column:published_at;not null" json:"published_at"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
}

// AuditLogはaudit_logsテーブル
type AuditLog struct {
	ID        int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Type      string         `gorm:"column:type;type:varchar;not null;index" json:"type"`
	UserID    *int64         `gorm:"column:user_id;index" json:"user_id"`
	User      *User          `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	IP        string         `gorm:"column:ip;type:varchar;not null" json:"ip"`
	UA        string         `gorm:"column:ua;type:varchar;not null" json:"ua"`
	MetaJSON  datatypes.JSON `gorm:"column:meta_json;type:jsonb;not null;default:'{}'" json:"meta_json"`
	CreatedAt time.Time      `gorm:"column:created_at;not null;autoCreateTime;index" json:"created_at"`
}
