package entity

import (
	"time"

	"gorm.io/datatypes"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// items.kindの値(記事・販売導線・関連情報リンクを同じItemとして扱う)
type ItemKind string

const (
	ItemKindNews   ItemKind = "news"
	ItemKindRecipe ItemKind = "recipe"
	ItemKindDeal   ItemKind = "deal"
	ItemKindShop   ItemKind = "shop"
)

// 焙煎度:検索条件や表示ラベル
type Roast string

const (
	RoastLight  Roast = "light"
	RoastMedium Roast = "medium"
	RoastDark   Roast = "dark"
)

// 抽出方法・飲み方
type Method string

const (
	MethodDrip     Method = "drip"
	MethodEspresso Method = "espresso"
	MethodMilk     Method = "milk"
	MethodIced     Method = "iced"
)

// 今の気分
type Mood string

const (
	MoodMorning Mood = "morning"
	MoodWork    Mood = "work"
	MoodRelax   Mood = "relax"
	MoodNight   Mood = "night"
)

// 利用シーン
type Scene string

const (
	SceneWork      Scene = "work"
	SceneBreak     Scene = "break"
	SceneAfterMeal Scene = "after_meal"
	SceneRelax     Scene = "relax"
)

// 飲み方
type TempPref string

const (
	TempHot TempPref = "hot"
	TempIce TempPref = "ice"
)

// 対話検索セッションの状態
type SessionStatus string

const (
	SessionActive SessionStatus = "active"
	SessionClosed SessionStatus = "closed"
)

// 発話主体(ユーザー・アシスタント・システム)
type TurnRole string

const (
	TurnRoleUser      TurnRole = "user"
	TurnRoleAssistant TurnRole = "assistant"
	TurnRoleSystem    TurnRole = "system"
)

// 発話の種別(普通の会話、追質問、通知)
type TurnKind string

const (
	TurnKindMessage  TurnKind = "message"
	TurnKindFollowup TurnKind = "followup"
	TurnKindNotice   TurnKind = "notice"
)

// usersテーブルに対応する認証主体。
type User struct {
	ID            uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Email         string    `gorm:"column:email;type:text;not null;uniqueIndex" json:"email"`
	PassHash      string    `gorm:"column:pass_hash;type:text;not null" json:"-"`
	Role          Role      `gorm:"column:role;type:text;not null" json:"role"`
	TokenVer      int       `gorm:"column:token_ver;not null;default:1" json:"token_ver"`
	EmailVerified bool      `gorm:"column:email_verified;not null;default:false" json:"email_verified"`
	CreatedAt     time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`
}

// email_verifiesテーブルに対応するメール認証トークン(1回使ったら無効になるワンタイム)。
type EmailVerify struct {
	ID        uint       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    uint       `gorm:"column:user_id;not null;index" json:"user_id"`
	User      User       `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	TokenHash string     `gorm:"column:token_hash;type:text;not null;uniqueIndex" json:"-"`
	ExpiresAt time.Time  `gorm:"column:expires_at;not null" json:"expires_at"`
	UsedAt    *time.Time `gorm:"column:used_at" json:"used_at"`
}

// pw_resetsテーブルに対応するパスワード再設定トークン。
type PwReset struct {
	ID        uint       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    uint       `gorm:"column:user_id;not null;index" json:"user_id"`
	User      User       `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	TokenHash string     `gorm:"column:token_hash;type:text;not null;uniqueIndex" json:"-"`
	ExpiresAt time.Time  `gorm:"column:expires_at;not null" json:"expires_at"`
	UsedAt    *time.Time `gorm:"column:used_at" json:"used_at"`
	CreatedAt time.Time  `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
}

// refresh_tokensテーブルに対応するrefresh tokenの管理。
type Rt struct {
	ID           uint       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID       uint       `gorm:"column:user_id;not null;index" json:"user_id"`
	User         User       `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	FamilyID     string     `gorm:"column:family_id;type:text;not null;index" json:"family_id"`
	TokenHash    string     `gorm:"column:token_hash;type:text;not null;uniqueIndex" json:"-"`
	ExpiresAt    time.Time  `gorm:"column:expires_at;not null" json:"expires_at"`
	RevokedAt    *time.Time `gorm:"column:revoked_at" json:"revoked_at"`
	UsedAt       *time.Time `gorm:"column:used_at" json:"used_at"`
	ReplacedByID *uint      `gorm:"column:replaced_by_id" json:"replaced_by_id"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
}

func (Rt) TableName() string {
	return "refresh_tokens"
}

// Source は sourcesテーブルに対応する情報源(出典、ブランド、店舗などの元情報)。
type Source struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name;type:text;not null" json:"name"`
	SiteURL   string    `gorm:"column:site_url;type:text;not null" json:"site_url"`
	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
}

// itemsテーブルに対応する記事・販売導線カード。
type Item struct {
	ID          uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Title       string    `gorm:"column:title;type:text;not null" json:"title"`
	Summary     string    `gorm:"column:summary;type:text;not null" json:"summary"`
	URL         string    `gorm:"column:url;type:text;not null" json:"url"`
	ImageURL    string    `gorm:"column:image_url;type:text;not null" json:"image_url"`
	Kind        ItemKind  `gorm:"column:kind;type:text;not null;index" json:"kind"`
	SourceID    uint      `gorm:"column:source_id;not null;index" json:"source_id"`
	Source      Source    `gorm:"foreignKey:SourceID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"source"`
	PublishedAt time.Time `gorm:"column:published_at;not null" json:"published_at"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
}

// JWTなどから取り出した操作した者の情報。
type Actor struct {
	UserID   uint `json:"user_id"`
	Role     Role `json:"role"`
	TokenVer int  `json:"token_ver"`
}

// Item一覧検索の条件。
type ItemQ struct {
	Q      string   `json:"q"`
	Kind   ItemKind `json:"kind"`
	Limit  int      `json:"limit"`
	Offset int      `json:"offset"`
}

// top表示用のカテゴリ別。

type TopItems struct {
	News   []Item `json:"news"`
	Recipe []Item `json:"recipe"`
	Deal   []Item `json:"deal"`
	Shop   []Item `json:"shop"`
}

// audit_logsテーブルに対応する監査ログ。
type AuditLog struct {
	ID        uint           `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Type      string         `gorm:"column:type;type:text;not null;index" json:"type"`
	UserID    *uint          `gorm:"column:user_id;index" json:"user_id"`
	User      *User          `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	IP        string         `gorm:"column:ip;type:text;not null" json:"ip"`
	UA        string         `gorm:"column:ua;type:text;not null" json:"ua"`
	Meta      datatypes.JSON `gorm:"column:meta;type:jsonb;not null;default:'{}'" json:"meta"`
	CreatedAt time.Time      `gorm:"column:created_at;not null;autoCreateTime;index" json:"created_at"`
}

// beansテーブルに対応する推薦の中心データ(5軸スコアと基本情報を持ち、検索)。
type Bean struct {
	ID         uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name       string    `gorm:"column:name;type:text;not null" json:"name"`
	Roast      Roast     `gorm:"column:roast;type:text;not null;index" json:"roast"`
	Origin     string    `gorm:"column:origin;type:text;not null" json:"origin"`
	Flavor     int       `gorm:"column:flavor;not null" json:"flavor"`
	Acidity    int       `gorm:"column:acidity;not null" json:"acidity"`
	Bitterness int       `gorm:"column:bitterness;not null" json:"bitterness"`
	Body       int       `gorm:"column:body;not null" json:"body"`
	Aroma      int       `gorm:"column:aroma;not null" json:"aroma"`
	Desc       string    `gorm:"column:desc;type:text;not null" json:"desc"`
	BuyURL     string    `gorm:"column:buy_url;type:text;not null" json:"buy_url"`
	Active     bool      `gorm:"column:active;not null;default:true;index" json:"active"`
	CreatedAt  time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`
}

// recipesテーブルに対応する抽出提案。
// Beanに必ず紐づいて、method/temp_prefの一致に使う。
type Recipe struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	BeanID    uint      `gorm:"column:bean_id;not null;index" json:"bean_id"`
	Bean      Bean      `gorm:"foreignKey:BeanID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"bean"`
	Name      string    `gorm:"column:name;type:text;not null" json:"name"`
	Method    Method    `gorm:"column:method;type:text;not null;index" json:"method"`
	TempPref  TempPref  `gorm:"column:temp_pref;type:text;not null" json:"temp_pref"`
	Grind     string    `gorm:"column:grind;type:text;not null" json:"grind"`
	Ratio     string    `gorm:"column:ratio;type:text;not null" json:"ratio"`
	Temp      int       `gorm:"column:temp;not null" json:"temp"`
	TimeSec   int       `gorm:"column:time_sec;not null" json:"time_sec"`
	Steps     []string  `gorm:"column:steps;type:text[];not null" json:"steps"`
	Desc      string    `gorm:"column:desc;type:text;not null" json:"desc"`
	Active    bool      `gorm:"column:active;not null;default:true" json:"active"`
	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`
}

// sessionsテーブルに対応する対話検索。
type Session struct {
	ID             uint          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID         *uint         `gorm:"column:user_id;index" json:"user_id"`
	User           *User         `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	Title          string        `gorm:"column:title;type:text;not null" json:"title"`
	Status         SessionStatus `gorm:"column:status;type:text;not null;index" json:"status"`
	SessionKeyHash string        `gorm:"column:session_key_hash;type:text;uniqueIndex" json:"-"`
	GuestExpiresAt *time.Time    `gorm:"column:guest_expires_at" json:"guest_expires_at"`
	CreatedAt      time.Time     `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time     `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`
}

// 会話本文そのものを時系列で保持(会話の履歴)。
type Turn struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	SessionID uint      `gorm:"column:session_id;not null;index" json:"session_id"`
	Session   Session   `gorm:"foreignKey:SessionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Role      TurnRole  `gorm:"column:role;type:text;not null" json:"role"`
	Kind      TurnKind  `gorm:"column:kind;type:text;not null" json:"kind"`
	Body      string    `gorm:"column:body;type:text;not null" json:"body"`
	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
}

// テーブルに対応する検索条件。
type Pref struct {
	ID         uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	SessionID  uint      `gorm:"column:session_id;not null;uniqueIndex" json:"session_id"`
	Session    Session   `gorm:"foreignKey:SessionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Flavor     int       `gorm:"column:flavor;not null" json:"flavor"`
	Acidity    int       `gorm:"column:acidity;not null" json:"acidity"`
	Bitterness int       `gorm:"column:bitterness;not null" json:"bitterness"`
	Body       int       `gorm:"column:body;not null" json:"body"`
	Aroma      int       `gorm:"column:aroma;not null" json:"aroma"`
	Mood       Mood      `gorm:"column:mood;type:text;not null" json:"mood"`
	Method     Method    `gorm:"column:method;type:text;not null" json:"method"`
	Scene      Scene     `gorm:"column:scene;type:text;not null" json:"scene"`
	TempPref   TempPref  `gorm:"column:temp_pref;type:text;not null" json:"temp_pref"`
	Excludes   []string  `gorm:"column:excludes;type:text[];not null" json:"excludes"`
	Note       string    `gorm:"column:note;type:text;not null" json:"note"`
	CreatedAt  time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;not null;autoUpdateTime" json:"updated_at"`
}

// 検索結果のスナップショットとして順位と理由を保存。
type Suggestion struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	SessionID uint      `gorm:"column:session_id;not null;index" json:"session_id"`
	Session   Session   `gorm:"foreignKey:SessionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	BeanID    uint      `gorm:"column:bean_id;not null;index" json:"bean_id"`
	Bean      Bean      `gorm:"foreignKey:BeanID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"bean"`
	RecipeID  *uint     `gorm:"column:recipe_id" json:"recipe_id"`
	Recipe    *Recipe   `gorm:"foreignKey:RecipeID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"recipe"`
	ItemID    *uint     `gorm:"column:item_id" json:"item_id"`
	Item      *Item     `gorm:"foreignKey:ItemID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"item"`
	Score     int       `gorm:"column:score;not null" json:"score"`
	Reason    string    `gorm:"column:reason;type:text;not null" json:"reason"`
	Rank      int       `gorm:"column:rank;not null" json:"rank"`
	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
}

// 検索結果そのものと、ユーザーが保存したという行動。
type SavedSuggestion struct {
	ID           uint       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID       uint       `gorm:"column:user_id;not null;index" json:"user_id"`
	User         User       `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	SessionID    uint       `gorm:"column:session_id;not null;index" json:"session_id"`
	Session      Session    `gorm:"foreignKey:SessionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	SuggestionID uint       `gorm:"column:suggestion_id;not null;index" json:"suggestion_id"`
	Suggestion   Suggestion `gorm:"foreignKey:SuggestionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"suggestion"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null;autoCreateTime" json:"created_at"`
}
