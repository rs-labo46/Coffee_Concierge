package usecase

import (
	"coffee-spa/entity"
	"math"
	"sort"
)

type coffeeRanker struct{}

func NewCoffeeRanker() Ranker {
	return &coffeeRanker{}
}

// Pref と Bean 候補一覧を受け取り、類似度順(0〜100の整数に丸め、scoreとしてRankItemに入れる)に並べ替えて返す。
func (r *coffeeRanker) Rank(pref entity.Pref, beans []entity.Bean) ([]RankItem, error) {
	out := make([]RankItem, 0, len(beans))

	for _, bean := range beans {
		score := r.similarityScore(pref, bean)

		out = append(out, RankItem{
			Bean:   bean,
			Score:  score,
			Reason: "",
		})
	}

	//scoreの高い順に並び替えで、もし同点だった場合はIDの昇順で。
	sort.Slice(out, func(i int, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].Bean.ID < out[j].Bean.ID
		}
		return out[i].Score > out[j].Score
	})

	return out, nil

}

// 5軸のユークリッド距離を0〜100の類似度へ変換(距離が小さいほど高得点になり、完全一致なら100)
func (r *coffeeRanker) similarityScore(pref entity.Pref, bean entity.Bean) int {
	distance := r.enclideanDistance(pref.Flavor, bean.Flavor, pref.Acidity, bean.Acidity, pref.Bitterness, bean.Bitterness, pref.Body, bean.Body, pref.Aroma, bean.Aroma)
	maxDistance := math.Sqrt(80.0)
	if maxDistance == 0 {
		return 0
	}
	similarity := 100.0 - (distance/maxDistance)*100.0
	if similarity < 0 {
		similarity = 0
	}
	if similarity > 100 {
		similarity = 100
	}
	return int(math.Round(similarity))
}

// 5軸の差分からユークリッド距離を計算
func (r *coffeeRanker) enclideanDistance(prefFlavor int, beanFlavor int, prefAcidity int, beanAcidity int, prefBitterness int, beanBitterness int, prefBody int, beanBody int, prefAroma int, beanAroma int) float64 {
	df := float64(prefFlavor - beanFlavor)
	da := float64(prefAcidity - beanAcidity)
	db := float64(prefBitterness - beanBitterness)
	dbo := float64(prefBody - beanBody)
	dar := float64(prefAroma - beanAroma)

	return math.Sqrt(
		df*df + da*da + db*db + dbo*dbo + dar*dar,
	)
}
