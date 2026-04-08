package policy

import (
	"net/mail"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"coffee-spa/entity"
	"coffee-spa/usecase"
)

// interface
type PwPol interface {
	usecase.PwPol
	Ok(pw string) error
}

type EmailPol interface {
	usecase.EmailPol
	Ok(email string) error
}

type TokenPol interface {
	usecase.TokenPol
	Ok(token string) error
}

type KindPol interface {
	usecase.KindPol
	Ok(kind string) error
}

type URLPol interface {
	usecase.URLPol
	Ok(raw string) error
}

type PagePol interface {
	usecase.PagePol
	Ok(limit int, offset int) error
}

type IDPol interface {
	usecase.IDPol
	Ok(id uint) error
}

type NamePol interface {
	usecase.NamePol
	Ok(name string) error
}

type TitlePol interface {
	usecase.TitlePol
	Ok(title string) error
}

type SummaryPol interface {
	usecase.SummaryPol
	Ok(summary string) error
}

type TextPol interface {
	usecase.TextPol
	Ok(text string) error
}

type BodyTextPol interface {
	usecase.BodyTextPol
	Ok(text string) error
}

type OriginPol interface {
	usecase.OriginPol
	Ok(origin string) error
}

type ScorePol interface {
	usecase.ScorePol
	Ok(score int) error
}

type RoastPol interface {
	usecase.RoastPol
	Ok(roast entity.Roast) error
}

type MethodPol interface {
	usecase.MethodPol
	Ok(method entity.Method) error
}

type MoodPol interface {
	usecase.MoodPol
	Ok(mood entity.Mood) error
}

type ScenePol interface {
	usecase.ScenePol
	Ok(scene entity.Scene) error
}

type TempPrefPol interface {
	usecase.TempPrefPol
	Ok(temp entity.TempPref) error
}

type AuditTypePol interface {
	usecase.AuditTypePol
	Ok(typ string) error
}

type TimePol interface {
	usecase.TimePol
	Ok(raw string) error
}

type TempPol interface {
	usecase.TempPol
	Ok(temp int) error
}

type TimeSecPol interface {
	usecase.TimeSecPol
	Ok(sec int) error
}

type StepsPol interface {
	usecase.StepsPol
	Ok(steps []string) error
}

type ExcludesPol interface {
	usecase.ExcludesPol
	Ok(excludes []string) error
}

// concrete type

type pwPol struct{}
type emailPol struct{}
type tokenPol struct{}
type kindPol struct{}
type urlPol struct{}
type pagePol struct{}
type idPol struct{}
type namePol struct{}
type titlePol struct{}
type summaryPol struct{}
type textPol struct{}
type bodyTextPol struct{}
type originPol struct{}
type scorePol struct{}
type roastPol struct{}
type methodPol struct{}
type moodPol struct{}
type scenePol struct{}
type tempPrefPol struct{}
type auditTypePol struct{}
type timePol struct{}
type tempPol struct{}
type timeSecPol struct{}
type stepsPol struct{}
type excludesPol struct{}

// constructor
func NewPwPol() PwPol {
	return &pwPol{}
}

func NewEmailPol() EmailPol {
	return &emailPol{}
}

func NewTokenPol() TokenPol {
	return &tokenPol{}
}

func NewKindPol() KindPol {
	return &kindPol{}
}

func NewURLPol() URLPol {
	return &urlPol{}
}

func NewPagePol() PagePol {
	return &pagePol{}
}

func NewIDPol() IDPol {
	return &idPol{}
}

func NewNamePol() NamePol {
	return &namePol{}
}

func NewTitlePol() TitlePol {
	return &titlePol{}
}

func NewSummaryPol() SummaryPol {
	return &summaryPol{}
}

func NewTextPol() TextPol {
	return &textPol{}
}

func NewBodyTextPol() BodyTextPol {
	return &bodyTextPol{}
}

func NewOriginPol() OriginPol {
	return &originPol{}
}

func NewScorePol() ScorePol {
	return &scorePol{}
}

func NewRoastPol() RoastPol {
	return &roastPol{}
}

func NewMethodPol() MethodPol {
	return &methodPol{}
}

func NewMoodPol() MoodPol {
	return &moodPol{}
}

func NewScenePol() ScenePol {
	return &scenePol{}
}

func NewTempPrefPol() TempPrefPol {
	return &tempPrefPol{}
}

func NewAuditTypePol() AuditTypePol {
	return &auditTypePol{}
}

func NewTimePol() TimePol {
	return &timePol{}
}

func NewTempPol() TempPol {
	return &tempPol{}
}

func NewTimeSecPol() TimeSecPol {
	return &timeSecPol{}
}

func NewStepsPol() StepsPol {
	return &stepsPol{}
}

func NewExcludesPol() ExcludesPol {
	return &excludesPol{}
}

// helper
func invalid() error {
	return usecase.ErrInvalidRequest
}

func runeLen(s string) int {
	return utf8.RuneCountInString(s)
}

func isBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

func nonEmptyShortText(s string, max int) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return invalid()
	}
	if runeLen(s) > max {
		return invalid()
	}
	return nil
}

// password
// usecase側から呼ばれるメソッド。
func (p *pwPol) Check(pw string) error {
	return p.Ok(pw)
}

func (p *pwPol) Ok(pw string) error {
	n := len(pw)
	if n < 12 || n > 72 {
		return invalid()
	}
	return nil
}

// email
func (p *emailPol) Check(email string) error {
	return p.Ok(email)
}

func (p *emailPol) Ok(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return invalid()
	}
	if len(email) > 254 {
		return invalid()
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return invalid()
	}
	return nil
}

// token
func (p *tokenPol) Check(token string) error {
	return p.Ok(token)
}

func (p *tokenPol) Ok(token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return invalid()
	}
	if len(token) < 16 || len(token) > 512 {
		return invalid()
	}
	return nil
}

// item kind
func (p *kindPol) Check(kind string) error {
	return p.Ok(kind)
}

func (p *kindPol) Ok(kind string) error {
	switch strings.TrimSpace(kind) {
	case string(entity.ItemKindNews),
		string(entity.ItemKindRecipe),
		string(entity.ItemKindDeal),
		string(entity.ItemKindShop):
		return nil
	default:
		return invalid()
	}
}

// url
func (p *urlPol) Check(raw string) error {
	return p.Ok(raw)
}

func (p *urlPol) Ok(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return invalid()
	}

	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return invalid()
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return invalid()
	}

	if u.Host == "" {
		return invalid()
	}

	return nil
}

// page
func (p *pagePol) Check(limit int, offset int) error {
	return p.Ok(limit, offset)
}

func (p *pagePol) Ok(limit int, offset int) error {
	if limit < 1 || limit > 50 {
		return invalid()
	}
	if offset < 0 {
		return invalid()
	}
	return nil
}

// id
func (p *idPol) Check(id uint) error {
	return p.Ok(id)
}

func (p *idPol) Ok(id uint) error {
	if id == 0 {
		return invalid()
	}
	return nil
}

// generic text
func (p *namePol) Check(name string) error {
	return p.Ok(name)
}

func (p *namePol) Ok(name string) error {
	return nonEmptyShortText(name, 100)
}

func (p *titlePol) Check(title string) error {
	return p.Ok(title)
}

func (p *titlePol) Ok(title string) error {
	return nonEmptyShortText(title, 200)
}

func (p *summaryPol) Check(summary string) error {
	return p.Ok(summary)
}

func (p *summaryPol) Ok(summary string) error {
	return nonEmptyShortText(summary, 1000)
}

func (p *textPol) Check(text string) error {
	return p.Ok(text)
}

func (p *textPol) Ok(text string) error {
	return nonEmptyShortText(text, 1000)
}

func (p *bodyTextPol) Check(text string) error {
	return p.Ok(text)
}

func (p *bodyTextPol) Ok(text string) error {
	return nonEmptyShortText(text, 2000)
}

func (p *originPol) Check(origin string) error {
	return p.Ok(origin)
}

func (p *originPol) Ok(origin string) error {
	return nonEmptyShortText(origin, 100)
}

// score / enum
func (p *scorePol) Check(score int) error {
	return p.Ok(score)
}

func (p *scorePol) Ok(score int) error {
	if score < 1 || score > 5 {
		return invalid()
	}
	return nil
}

func (p *roastPol) Check(roast entity.Roast) error {
	return p.Ok(roast)
}

func (p *roastPol) Ok(roast entity.Roast) error {
	switch roast {
	case entity.RoastLight, entity.RoastMedium, entity.RoastDark:
		return nil
	default:
		return invalid()
	}
}

func (p *methodPol) Check(method entity.Method) error {
	return p.Ok(method)
}

func (p *methodPol) Ok(method entity.Method) error {
	switch method {
	case entity.MethodDrip, entity.MethodEspresso, entity.MethodMilk, entity.MethodIced:
		return nil
	default:
		return invalid()
	}
}

func (p *moodPol) Check(mood entity.Mood) error {
	return p.Ok(mood)
}

func (p *moodPol) Ok(mood entity.Mood) error {
	switch mood {
	case entity.MoodMorning, entity.MoodWork, entity.MoodRelax, entity.MoodNight:
		return nil
	default:
		return invalid()
	}
}

func (p *scenePol) Check(scene entity.Scene) error {
	return p.Ok(scene)
}

func (p *scenePol) Ok(scene entity.Scene) error {
	switch scene {
	case entity.SceneWork, entity.SceneBreak, entity.SceneAfterMeal, entity.SceneRelax:
		return nil
	default:
		return invalid()
	}
}

func (p *tempPrefPol) Check(temp entity.TempPref) error {
	return p.Ok(temp)
}

func (p *tempPrefPol) Ok(temp entity.TempPref) error {
	switch temp {
	case entity.TempHot, entity.TempIce:
		return nil
	default:
		return invalid()
	}
}

// audit type
func (p *auditTypePol) Check(typ string) error {
	return p.Ok(typ)
}

func (p *auditTypePol) Ok(typ string) error {
	typ = strings.TrimSpace(typ)

	// 空は一覧検索の「全件」を許容する。
	if typ == "" {
		return nil
	}

	if runeLen(typ) > 100 {
		return invalid()
	}

	if !strings.Contains(typ, ".") {
		return invalid()
	}

	return nil
}

// time
func (p *timePol) Check(raw string) error {
	return p.Ok(raw)
}

func (p *timePol) Ok(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return invalid()
	}

	// RFC3339
	if _, err := time.Parse(time.RFC3339, raw); err != nil {
		return invalid()
	}

	return nil
}

// temp / time_sec
func (p *tempPol) Check(temp int) error {
	return p.Ok(temp)
}

func (p *tempPol) Ok(temp int) error {
	if temp < 60 || temp > 100 {
		return invalid()
	}
	return nil
}

func (p *timeSecPol) Check(sec int) error {
	return p.Ok(sec)
}

func (p *timeSecPol) Ok(sec int) error {
	if sec < 1 || sec > 600 {
		return invalid()
	}
	return nil
}

// steps
func (p *stepsPol) Check(steps []string) error {
	return p.Ok(steps)
}

func (p *stepsPol) Ok(steps []string) error {
	if len(steps) == 0 {
		return invalid()
	}
	if len(steps) > 20 {
		return invalid()
	}

	for _, step := range steps {
		if isBlank(step) {
			return invalid()
		}
		if runeLen(step) > 500 {
			return invalid()
		}
	}

	return nil
}

// excludes
func (p *excludesPol) Check(excludes []string) error {
	return p.Ok(excludes)
}

func (p *excludesPol) Ok(excludes []string) error {
	if len(excludes) > 10 {
		return invalid()
	}

	allowed := map[string]struct{}{
		"acidic":      {},
		"bitter":      {},
		"dark_roast":  {},
		"milk_recipe": {},
	}

	for _, v := range excludes {
		v = strings.TrimSpace(v)
		if v == "" {
			return invalid()
		}
		if _, ok := allowed[v]; !ok {
			return invalid()
		}
	}

	return nil
}
