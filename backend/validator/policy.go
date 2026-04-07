package validator

import (
	"coffee-spa/entity"
	"time"
)

// メールアドレス形式を検証
type EmailPol interface {
	Check(email string) error
}

// パスワードの形式を検証
type PwPol interface {
	Check(email string) error
}

// トークンの文字列を検証
type TokenPol interface {
	Check(token string) error
}

// URLの文字列を検証
type URLPol interface {
	Check(raw string) error
}

// source名やbean名などの名前の文字列を検証
type NamePol interface {
	Check(name string) error
}

// タイトルの文字列を検証
type TitlePol interface {
	Check(title string) error
}

// 要約の文字列を検証
type Summary interface {
	Check(summary string) error
}

// recipe系の短文を検証
type TextPol interface {
	Check(text string) error
}

// 発話の本文を検証
type BodyTextPol interface {
	Check(body string) error
}

// 産地の文字列を検証
type OriginPol interface {
	Check(origin string) error
}

// 1~5のスコアを検証
type ScorePol interface {
	Check(v int) error
}

// 焙煎度を検証
type RoastPol interface {
	Check(v entity.Roast) error
}

// 抽出方法を検証
type MethodPol interface {
	Check(v entity.Method) error
}

// 気分を検証
type ScenePol interface {
	Check(v entity.Scene) error
}

// 温度の検証
type TemplePol interface {
	Check(v entity.TempPref) error
}

// Itemの種別を検証
type ItemKindPol interface {
	Check(v entity.ItemKind) error
}

// Auditの種別を検証
type AuditTypePol interface {
	Check(v string) error
}

// 日時を検証
type TimePol interface {
	Check(t time.Time) error
}

// 温度の値を検証
type TempPol interface {
	Check(v int) error
}

// 秒数を検証(0は禁止)
type TimeSecPol interface {
	Check(v int) error
}

// 手順配列を検証(空禁止)
type StepsPol interface {
	Check(v []string) error
}

// 文字数や件数など除外条件の配列を検証
type ExcludesPol interface {
	Check(v []string) error
}

// IDの値を検証
type IDPol interface {
	Check(id uint) error
}

// ページングの値を検証
type PagePol interface {
	Check(limit int, offset int) error
}
