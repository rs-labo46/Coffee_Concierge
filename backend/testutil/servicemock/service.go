package servicemock

import (
	"testing"
	"time"

	"coffee-spa/entity"
	"coffee-spa/usecase"
)

type Clock struct {
	T     *testing.T
	NowFn func() time.Time
}

func (m Clock) Now() time.Time {
	if m.NowFn == nil {
		return time.Unix(1700000000, 0).UTC()
	}
	return m.NowFn()
}

type IDGen struct {
	T     *testing.T
	NewFn func() string
}

func (m IDGen) New() string {
	if m.NewFn == nil {
		return "test-id"
	}
	return m.NewFn()
}

type Hasher struct {
	T         *testing.T
	HashFn    func(string) (string, error)
	CompareFn func(string, string) error
}

func (m Hasher) Hash(raw string) (string, error) {
	if m.HashFn == nil {
		return "hashed:" + raw, nil
	}
	return m.HashFn(raw)
}

func (m Hasher) Compare(hash string, raw string) error {
	if m.CompareFn == nil {
		return nil
	}
	return m.CompareFn(hash, raw)
}

type Token struct {
	T            *testing.T
	SignAccessFn func(entity.User) (string, error)
}

func (m Token) SignAccess(user entity.User) (string, error) {
	if m.SignAccessFn == nil {
		return "access-token", nil
	}
	return m.SignAccessFn(user)
}

type Refresh struct {
	T      *testing.T
	NewFn  func() (string, string, error)
	HashFn func(string) string
}

func (m Refresh) New() (string, string, error) {
	if m.NewFn == nil {
		return "refresh-plain", "refresh-hash", nil
	}
	return m.NewFn()
}

func (m Refresh) Hash(token string) string {
	if m.HashFn == nil {
		return "hash:" + token
	}
	return m.HashFn(token)
}

type Mailer struct {
	T                  *testing.T
	SendVerifyEmailFn  func(string, string) error
	SendResetPwEmailFn func(string, string) error
}

func (m Mailer) SendVerifyEmail(to string, token string) error {
	if m.SendVerifyEmailFn == nil {
		return nil
	}
	return m.SendVerifyEmailFn(to, token)
}

func (m Mailer) SendResetPwEmail(to string, token string) error {
	if m.SendResetPwEmailFn == nil {
		return nil
	}
	return m.SendResetPwEmailFn(to, token)
}

type Ranker struct {
	T      *testing.T
	RankFn func(entity.Pref, []entity.Bean) ([]usecase.RankItem, error)
}

func (m Ranker) Rank(pref entity.Pref, beans []entity.Bean) ([]usecase.RankItem, error) {
	if m.RankFn == nil {
		items := make([]usecase.RankItem, 0, len(beans))
		for i, bean := range beans {
			items = append(items, usecase.RankItem{Bean: bean, Score: 100 - i, Reason: "matched"})
		}
		return items, nil
	}
	return m.RankFn(pref, beans)
}

type Gemini struct {
	T                    *testing.T
	BuildConditionDiffFn func(usecase.GeminiConditionDiffIn) (usecase.GeminiConditionDiffOut, usecase.GeminiAuditMeta, error)
	SelectBeansFn        func(usecase.GeminiBeanSelectionIn) ([]usecase.GeminiBeanSelection, usecase.GeminiAuditMeta, error)
	BuildSearchBundleFn  func(usecase.GeminiSearchBundleIn) (usecase.GeminiSearchBundleOut, usecase.GeminiAuditMeta, error)
	BuildReasonsFn       func(usecase.GeminiReasonIn) ([]usecase.GeminiReason, usecase.GeminiAuditMeta, error)
	BuildFollowupsFn     func(usecase.GeminiFollowupIn) ([]string, usecase.GeminiAuditMeta, error)
}

func (m Gemini) Info() (string, string) {
	return "mock", "mock-model"
}

func (m Gemini) BuildConditionDiff(in usecase.GeminiConditionDiffIn) (usecase.GeminiConditionDiffOut, usecase.GeminiAuditMeta, error) {
	meta := usecase.GeminiAuditMeta{Provider: "mock", Model: "mock-model", Status: "success"}
	if m.BuildConditionDiffFn == nil {
		return usecase.GeminiConditionDiffOut{}, meta, nil
	}
	return m.BuildConditionDiffFn(in)
}

func (m Gemini) SelectBeans(in usecase.GeminiBeanSelectionIn) ([]usecase.GeminiBeanSelection, usecase.GeminiAuditMeta, error) {
	meta := usecase.GeminiAuditMeta{Provider: "mock", Model: "mock-model", Status: "success"}
	if m.SelectBeansFn == nil {
		out := make([]usecase.GeminiBeanSelection, 0, len(in.Candidates))
		limit := in.Limit
		if limit <= 0 {
			limit = 10
		}
		for _, bean := range in.Candidates {
			out = append(out, usecase.GeminiBeanSelection{BeanID: bean.ID, Rank: len(out) + 1, Score: 90, Reason: "登録済み豆から選定しました。"})
			if len(out) >= limit {
				break
			}
		}
		return out, meta, nil
	}
	return m.SelectBeansFn(in)
}

func (m Gemini) BuildSearchBundle(in usecase.GeminiSearchBundleIn) (usecase.GeminiSearchBundleOut, usecase.GeminiAuditMeta, error) {
	meta := usecase.GeminiAuditMeta{Provider: "mock", Model: "mock-model", Status: "success"}
	if m.BuildSearchBundleFn == nil {
		limit := in.Limit
		if limit <= 0 {
			limit = 10
		}
		selections := make([]usecase.GeminiBeanSelection, 0, limit)
		for _, bean := range in.Candidates {
			selections = append(selections, usecase.GeminiBeanSelection{BeanID: bean.ID, Rank: len(selections) + 1, Score: 90, Reason: "登録済み豆から選定しました。"})
			if len(selections) >= limit {
				break
			}
		}
		return usecase.GeminiSearchBundleOut{Selections: selections, Followups: []string{"もう少し軽めにしますか？"}}, meta, nil
	}
	return m.BuildSearchBundleFn(in)
}

func (m Gemini) BuildReasons(in usecase.GeminiReasonIn) ([]usecase.GeminiReason, usecase.GeminiAuditMeta, error) {
	meta := usecase.GeminiAuditMeta{Provider: "mock", Model: "mock-model", Status: "success"}
	if m.BuildReasonsFn == nil {
		return nil, meta, nil
	}
	return m.BuildReasonsFn(in)
}

func (m Gemini) BuildFollowups(in usecase.GeminiFollowupIn) ([]string, usecase.GeminiAuditMeta, error) {
	meta := usecase.GeminiAuditMeta{Provider: "mock", Model: "mock-model", Status: "success"}
	if m.BuildFollowupsFn == nil {
		return []string{"もう少し軽めにしますか？"}, meta, nil
	}
	return m.BuildFollowupsFn(in)
}
