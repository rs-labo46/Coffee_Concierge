package usecase

import (
	"sort"
	"strings"

	"coffee-spa/entity"
)

// beanToGeminiCandidatesは、DBのBeanをGeminiへ渡す最小限の公開JSONに変換する。
func beanToGeminiCandidates(beans []entity.Bean) []GeminiBeanCandidate {
	out := make([]GeminiBeanCandidate, 0, len(beans))
	for _, bean := range beans {
		if !bean.Active {
			continue
		}
		out = append(out, GeminiBeanCandidate{
			ID:         bean.ID,
			Name:       bean.Name,
			Roast:      bean.Roast,
			Origin:     bean.Origin,
			Flavor:     bean.Flavor,
			Acidity:    bean.Acidity,
			Bitterness: bean.Bitterness,
			Body:       bean.Body,
			Aroma:      bean.Aroma,
			Desc:       bean.Desc,
		})
	}
	return out
}

// applyBeanSelectionsは、Geminiの選定結果を登録済みBeanのみに絞り、rank順のRankItemへ変換する。
func applyBeanSelections(
	fallback []RankItem,
	allBeans []entity.Bean,
	selections []GeminiBeanSelection,
	limit int,
) []RankItem {
	if limit <= 0 {
		limit = 10
	}

	beanByID := make(map[uint]entity.Bean, len(allBeans))
	for _, bean := range allBeans {
		if bean.Active {
			beanByID[bean.ID] = bean
		}
	}

	seen := make(map[uint]struct{}, len(selections))
	valid := make([]GeminiBeanSelection, 0, len(selections))

	for _, selection := range selections {
		if selection.BeanID == 0 || selection.Rank <= 0 || selection.Score < 0 || selection.Score > 100 {
			continue
		}
		if strings.TrimSpace(selection.Reason) == "" {
			continue
		}
		if _, ok := beanByID[selection.BeanID]; !ok {
			continue
		}
		if _, ok := seen[selection.BeanID]; ok {
			continue
		}
		seen[selection.BeanID] = struct{}{}
		valid = append(valid, selection)
	}

	sort.SliceStable(valid, func(i int, j int) bool {
		if valid[i].Rank == valid[j].Rank {
			return valid[i].Score > valid[j].Score
		}
		return valid[i].Rank < valid[j].Rank
	})

	out := make([]RankItem, 0, limit)
	for _, selection := range valid {
		bean := beanByID[selection.BeanID]
		out = append(out, RankItem{
			Bean:   bean,
			Score:  selection.Score,
			Reason: strings.TrimSpace(selection.Reason),
		})
		if len(out) >= limit {
			return out
		}
	}

	for _, item := range fallback {
		if _, ok := seen[item.Bean.ID]; ok {
			continue
		}
		out = append(out, item)
		if len(out) >= limit {
			break
		}
	}

	return out
}
