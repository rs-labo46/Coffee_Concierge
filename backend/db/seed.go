package db

import (
	"fmt"
	"strings"
	"time"

	"coffee-spa/entity"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedDev(db *gorm.DB, adminEmail string, adminPassword string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := seedAdmin(tx, adminEmail, adminPassword); err != nil {
			return err
		}

		if err := seedSources(tx); err != nil {
			return err
		}

		if err := seedItems(tx); err != nil {
			return err
		}

		return nil
	})
}

type seedUserRow struct {
	Email         string    `gorm:"column:email"`
	PassHash      string    `gorm:"column:pass_hash"`
	Role          string    `gorm:"column:role"`
	TokenVer      int       `gorm:"column:token_ver"`
	EmailVerified bool      `gorm:"column:email_verified"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

// adminを作る。
// 既に同じメールがあれば、dev用としてadmin / verified / passwordを上書きする。
func seedAdmin(db *gorm.DB, adminEmail string, adminPassword string) error {
	email := normalizeEmail(adminEmail)
	pw := strings.TrimSpace(adminPassword)

	if email == "" {
		return fmt.Errorf("seed admin email is empty")
	}

	if pw == "" {
		return fmt.Errorf("seed admin password is empty")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now()

	var count int64
	if err := db.Table("users").
		Where("email = ?", email).
		Count(&count).Error; err != nil {
		return err
	}

	if count == 0 {
		row := seedUserRow{
			Email:         email,
			PassHash:      string(hash),
			Role:          "admin",
			TokenVer:      1,
			EmailVerified: true,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if err := db.Table("users").Create(&row).Error; err != nil {
			return err
		}

		return nil
	}

	upd := seedUserRow{
		PassHash:      string(hash),
		Role:          "admin",
		TokenVer:      1,
		EmailVerified: true,
		UpdatedAt:     now,
	}

	if err := db.Table("users").
		Where("email = ?", email).
		Updates(&upd).Error; err != nil {
		return err
	}

	return nil
}

func normalizeEmail(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func seedSources(db *gorm.DB) error {
	var count int64
	if err := db.Model(&entity.Source{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	sources := []entity.Source{
		{
			Name:    "Coffee Daily",
			SiteURL: "https://example.com/coffee-daily",
		},
		{
			Name:    "Roast Journal",
			SiteURL: "https://example.com/coffee-daily",
		},
		{
			Name:    "Home Brew Note",
			SiteURL: "https://example.com/coffee-daily",
		},
		{
			Name:    "Cafe Guide",
			SiteURL: "https://example.com/coffee-daily",
		},
	}

	for _, src := range sources {
		if err := db.Create(&src).Error; err != nil {
			return err
		}
	}

	return nil
}

func seedItems(db *gorm.DB) error {
	var sources []entity.Source
	if err := db.Order("id asc").Find(&sources).Error; err != nil {
		return err
	}

	if len(sources) == 0 {
		return fmt.Errorf("seed source not found")
	}

	now := time.Now()

	type seedGroup struct {
		kind     entity.ItemKind
		sourceID uint
		build    func(time.Time, uint) []entity.Item
	}

	groups := []seedGroup{
		{
			kind:     entity.ItemKindNews,
			sourceID: sources[0%len(sources)].ID,
			build:    buildNewsItems,
		},
		{
			kind:     entity.ItemKindRecipe,
			sourceID: sources[1%len(sources)].ID,
			build:    buildRecipeItems,
		},
		{
			kind:     entity.ItemKindDeal,
			sourceID: sources[2%len(sources)].ID,
			build:    buildDealItems,
		},
		{
			kind:     entity.ItemKindShop,
			sourceID: sources[3%len(sources)].ID,
			build:    buildShopItems,
		},
	}

	for _, g := range groups {
		var count int64
		if err := db.Model(&entity.Item{}).
			Where("kind = ?", g.kind).
			Count(&count).Error; err != nil {
			return err
		}

		if count > 0 {
			continue
		}

		items := g.build(now, g.sourceID)
		for _, item := range items {
			if err := db.Create(&item).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func min4(a int, b int, c int, d int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	if d < m {
		m = d
	}
	return m
}

func buildNewsItems(now time.Time, sourceID uint) []entity.Item {
	titles := []string{
		"スペシャルティコーヒー市場で浅煎り需要が再び拡大",
		"都市型ロースタリーがサブスク会員向け焙煎便を開始",
		"ペーパーフィルター価格の見直しで家庭抽出のコスト感に変化",
		"エチオピア新豆の入荷が始まりフローラル系の注目が上昇",
		"カフェ運営者の間で小型焙煎機の導入相談が増加",
		"コーヒーイベントで抽出器具の比較展示が話題に",
		"リユースカップ運用を進める店舗が都心部で増えている",
		"ミルの粒度安定性を重視した家庭用モデルが人気",
		"豆価格の変動を受けて、定番ブレンドの構成比を調整する店舗が増加",
		"抽出ログを活用した接客改善を進めるカフェが増えている",
	}

	summaries := []string{
		"業界トレンドの確認用ダミーデータです。短めの概要文を入れています。",
		"会員制モデルと定期配送の組み合わせが、小規模ロースターでも試され始めています。",
		"",
		"花のような香りや柑橘系の明るさを前面に出した構成が目立っています。",
		"導入コストを抑えながら自家焙煎へ移行したい事業者向けの話題です。",
		"",
		"実店舗での体験価値と環境配慮を両立する取り組みとして注目されています。",
		"刃の違い、回転数、清掃性など、比較軸が一般消費者にも広がっています。",
		"",
		"会員データや抽出記録をもとに、提案精度を上げる店舗運営が注目されています。",
	}

	images := []string{
		"https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?auto=format&fit=crop&w=1200&q=80",
		"https://images.unsplash.com/photo-1447933601403-0c6688de566e?auto=format&fit=crop&w=1200&q=80",
		"",
		"https://images.unsplash.com/photo-1517701604599-bb29b565090c?auto=format&fit=crop&w=1200&q=80",
		"",
		"https://images.unsplash.com/photo-1459755486867-b55449bb39ff?auto=format&fit=crop&w=1200&q=80",
		"",
		"https://images.unsplash.com/photo-1461988320302-91bde64fc8e4?auto=format&fit=crop&w=1200&q=80",
		"",
		"https://images.unsplash.com/photo-1494314671902-399b18174975?auto=format&fit=crop&w=1200&q=80",
	}

	urls := []string{
		"https://example.com/news/1",
		"https://example.com/news/2",
		"",
		"https://example.com/news/4",
		"",
		"https://example.com/news/6",
		"",
		"https://example.com/news/8",
		"",
		"https://example.com/news/10",
	}

	n := min4(len(titles), len(summaries), len(images), len(urls))
	items := make([]entity.Item, 0, n)

	for i := 0; i < n; i++ {
		items = append(items, entity.Item{
			Title:       titles[i],
			Summary:     seedSummary(titles[i], summaries[i]),
			URL:         strOrEmpty(urls[i]),
			ImageURL:    strOrEmpty(images[i]),
			Kind:        entity.ItemKindNews,
			SourceID:    sourceID,
			PublishedAt: now.Add(time.Duration(-(i + 1)) * 6 * time.Hour),
		})
	}

	return items
}

func buildRecipeItems(now time.Time, sourceID uint) []entity.Item {
	titles := []string{
		"ハンドドリップの基本比率を見直して甘さを出すレシピ",
		"アイスコーヒー向けに濃度を上げた抽出手順",
		"フレンチプレスで雑味を抑える湯温の考え方",
		"朝の一杯を早く淹れるための時短ドリップ構成",
		"中煎り豆でバランスを崩しにくい家庭向けレシピ",
		"エアロプレスで酸味を丸くする短時間抽出",
		"少量抽出でも味を薄くしにくい一人分レシピ",
		"来客時に安定して淹れやすい二杯取りの基準",
		"牛乳に合わせやすい深煎り向けの濃い抽出レシピ",
		"蒸らしを長めに取って香りを立たせる週末向けレシピ",
	}

	summaries := []string{
		"粉量、湯量、抽出時間の基本を見直したダミーレシピです。",
		"",
		"湯温を少し下げるだけでも口当たりが穏やかになります。",
		"",
		"失敗しにくいレシピを置いて、一覧の見え方を確認します。",
		"短時間でも薄くなりすぎないように攪拌を調整する想定です。",
		"",
		"抽出量が増えた時に味がぶれやすい人向けです。",
		"",
		"蒸らし時間を少し長めに取り、香りと甘さの立ち上がりを狙う想定です。",
	}

	images := []string{
		"https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?auto=format&fit=crop&w=1200&q=80",
		"",
		"https://images.unsplash.com/photo-1517701604599-bb29b565090c?auto=format&fit=crop&w=1200&q=80",
		"",
		"https://images.unsplash.com/photo-1442512595331-e89e73853f31?auto=format&fit=crop&w=1200&q=80",
		"",
		"https://images.unsplash.com/photo-1509042239860-f550ce710b93?auto=format&fit=crop&w=1200&q=80",
		"",
		"https://images.unsplash.com/photo-1461023058943-07fcbe16d735?auto=format&fit=crop&w=1200&q=80",
		"",
	}

	urls := []string{
		"",
		"https://example.com/recipe/2",
		"",
		"https://example.com/recipe/4",
		"",
		"https://example.com/recipe/6",
		"",
		"https://example.com/recipe/8",
		"",
		"https://example.com/recipe/10",
	}

	n := min4(len(titles), len(summaries), len(images), len(urls))
	items := make([]entity.Item, 0, n)

	for i := 0; i < n; i++ {
		items = append(items, entity.Item{
			Title:       titles[i],
			Summary:     seedSummary(titles[i], summaries[i]),
			URL:         strOrEmpty(urls[i]),
			ImageURL:    strOrEmpty(images[i]),
			Kind:        entity.ItemKindNews,
			SourceID:    sourceID,
			PublishedAt: now.Add(time.Duration(-(i + 1)) * 6 * time.Hour),
		})
	}

	return items
}

func buildDealItems(now time.Time, sourceID uint) []entity.Item {
	titles := []string{
		"週末限定でドリッパーが10%オフ",
		"初回購入向けの送料無料キャンペーン",
		"深煎りセットのまとめ買い値引き",
		"春の新生活向けコーヒー器具セール",
		"ミルとケトルの同時購入で割引適用",
		"定期便スタート記念のクーポン配布",
		"アイスコーヒー器具の季節セール",
		"店舗受け取り限定の豆セット特価",
		"レビュー投稿で次回使えるクーポン配布",
		"抽出スターターセットの期間限定セール",
	}

	summaries := []string{
		"価格表示の見え方確認用ダミーデータ。",
		"",
		"まとめ買い導線がある時の一覧密度を確かめます。",
		"",
		"複数商品を組み合わせた訴求の見え方確認用です。",
		"",
		"季節キャンペーンの短い説明文です。",
		"",
		"",
		"初級者向け器具をまとめた訴求の見え方確認用です。",
	}

	images := []string{
		"",
		"https://images.unsplash.com/photo-1512568400610-62da28bc8a13?auto=format&fit=crop&w=1200&q=80",
		"",
		"https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?auto=format&fit=crop&w=1200&q=80",
		"",
		"https://images.unsplash.com/photo-1447933601403-0c6688de566e?auto=format&fit=crop&w=1200&q=80",
		"",
		"https://images.unsplash.com/photo-1453614512568-c4024d13c247?auto=format&fit=crop&w=1200&q=80",
		"",
		"https://images.unsplash.com/photo-1461988320302-91bde64fc8e4?auto=format&fit=crop&w=1200&q=80",
	}

	urls := []string{
		"https://example.com/deal/1",
		"",
		"https://example.com/deal/3",
		"",
		"https://example.com/deal/5",
		"",
		"https://example.com/deal/7",
		"",
		"https://example.com/deal/9",
		"",
	}

	n := min4(len(titles), len(summaries), len(images), len(urls))
	items := make([]entity.Item, 0, n)

	for i := 0; i < n; i++ {
		items = append(items, entity.Item{
			Title:       titles[i],
			Summary:     seedSummary(titles[i], summaries[i]),
			URL:         strOrEmpty(urls[i]),
			ImageURL:    strOrEmpty(images[i]),
			Kind:        entity.ItemKindDeal,
			SourceID:    sourceID,
			PublishedAt: now.Add(time.Duration(-(i + 1)) * 6 * time.Hour),
		})
	}

	return items
}

func buildShopItems(now time.Time, sourceID uint) []entity.Item {
	titles := []string{
		"駅前に小型ロースタリー併設店がオープン",
		"朝営業に強いカフェの新店舗情報",
		"自家製スイーツと相性が良い人気店",
		"深夜まで営業する作業向けカフェ",
		"豆の量り売りに対応した地域密着店",
		"静かな空間でハンドドリップを味わえる店",
		"テイクアウト需要に強いスタンド型ショップ",
		"焙煎体験イベントを行う店舗の紹介",
		"地方ロースターの豆を週替わりで出す店",
		"駅近で朝の回転が速いエスプレッソバーの紹介",
	}

	summaries := []string{
		"新店カードの見え方確認用です。",
		"朝利用しやすい店舗情報を想定した短い説明です。",
		"",
		"作業利用、席数、電源有無などが気になる人向けの想定です。",
		"",
		"静かな店のニーズ確認用です。",
		"",
		"イベント性のある店舗情報が混ざった時の見え方確認です。",
		"",
		"朝の通勤導線に入りやすい立地と提供スピードを想定した説明です。",
	}

	images := []string{
		"https://images.unsplash.com/photo-1442512595331-e89e73853f31?auto=format&fit=crop&w=1200&q=80",
		"",
		"https://images.unsplash.com/photo-1453614512568-c4024d13c247?auto=format&fit=crop&w=1200&q=80",
		"",
		"https://images.unsplash.com/photo-1494314671902-399b18174975?auto=format&fit=crop&w=1200&q=80",
		"",
		"https://images.unsplash.com/photo-1509042239860-f550ce710b93?auto=format&fit=crop&w=1200&q=80",
		"",
		"https://images.unsplash.com/photo-1517701604599-bb29b565090c?auto=format&fit=crop&w=1200&q=80",
		"",
	}

	urls := []string{
		"",
		"https://example.com/shop/2",
		"",
		"https://example.com/shop/4",
		"",
		"https://example.com/shop/6",
		"",
		"https://example.com/shop/8",
		"",
		"https://example.com/shop/10",
	}

	n := min4(len(titles), len(summaries), len(images), len(urls))
	items := make([]entity.Item, 0, n)

	for i := 0; i < n; i++ {
		items = append(items, entity.Item{
			Title:       titles[i],
			Summary:     seedSummary(titles[i], summaries[i]),
			URL:         strOrEmpty(urls[i]),
			ImageURL:    strOrEmpty(images[i]),
			Kind:        entity.ItemKindShop,
			SourceID:    sourceID,
			PublishedAt: now.Add(time.Duration(-(i + 1)) * 6 * time.Hour),
		})
	}

	return items
}

func seedSummary(title string, raw string) string {
	m := map[string]string{
		"スペシャルティコーヒー市場で浅煎り需要が再び拡大":        "浅煎りの果実感や透明感を求める客層が増え、都市部のロースターで浅煎り豆の販売比率が上がっているという内容です。",
		"都市型ロースタリーがサブスク会員向け焙煎便を開始":        "焙煎日指定の定期便を始めたロースタリーが、会員向けに限定豆や抽出メモを同封する取り組みを始めた記事です。",
		"ペーパーフィルター価格の見直しで家庭抽出のコスト感に変化":    "フィルターや消耗品の価格改定が続き、自宅での一杯あたり原価を見直す人が増えているという話題です。",
		"エチオピア新豆の入荷が始まりフローラル系の注目が上昇":      "新豆の入荷により、ジャスミン系の香りや柑橘感を楽しめるエチオピアロットへの注目が高まっている内容です。",
		"カフェ運営者の間で小型焙煎機の導入相談が増加":          "客席数の少ない店でも自家焙煎へ移行しやすい小型焙煎機の相談が増えている背景をまとめた記事です。",
		"コーヒーイベントで抽出器具の比較展示が話題に":          "同じ豆を器具違いで飲み比べる展示が人気を集め、抽出器具の違いを体感できる企画が注目された話題です。",
		"リユースカップ運用を進める店舗が都心部で増えている":       "持ち帰り需要の増加に合わせて、返却型カップや割引施策を導入する店舗が増えている動きをまとめています。",
		"ミルの粒度安定性を重視した家庭用モデルが人気":          "家庭用ミル選びで価格よりも粒度の揃い方や清掃性を重視する人が増えているという内容です。",
		"豆価格の変動を受けて、定番ブレンドの構成比を調整する店舗が増加": "豆相場の変動を受けて、定番ブレンドの配合を変えながら味の軸を維持しようとする店が増えている話です。",
		"抽出ログを活用した接客改善を進めるカフェが増えている":      "抽出レシピや味の反応を記録し、接客や再提案の精度を上げる店舗運営が広がっている内容です。",

		"ハンドドリップの基本比率を見直して甘さを出すレシピ": "粉量と湯量の比率を見直し、甘さを感じやすい濃度帯へ寄せる基本レシピの記事です。",
		"アイスコーヒー向けに濃度を上げた抽出手順":      "氷で薄まる前提で、ホット抽出時点の濃度を高めに設計するアイス向けレシピです。",
		"フレンチプレスで雑味を抑える湯温の考え方":      "抽出温度を数度下げるだけで、重たさを残しつつ雑味を減らす考え方をまとめた内容です。",
		"朝の一杯を早く淹れるための時短ドリップ構成":     "忙しい朝でも味を崩しにくいように、手順を減らして再現性を優先したレシピです。",
		"中煎り豆でバランスを崩しにくい家庭向けレシピ":    "酸味と苦味が暴れにくく、家庭でも安定しやすい中煎り豆向けの抽出指針です。",
		"エアロプレスで酸味を丸くする短時間抽出":       "短時間抽出でも刺さらない酸味に寄せるための攪拌と押し切り時間の調整を扱っています。",
		"少量抽出でも味を薄くしにくい一人分レシピ":      "一杯分だけ淹れるときに起こりやすい薄さを避けるための粉量と注湯回数の工夫です。",
		"来客時に安定して淹れやすい二杯取りの基準":      "二杯分をまとめて淹れるときに味ぶれを減らすための比率と注ぎ方を整理した記事です。",
		"牛乳に合わせやすい深煎り向けの濃い抽出レシピ":    "ラテやオレにしても風味が負けないよう、苦味と厚みを意識した深煎り用レシピです。",
		"蒸らしを長めに取って香りを立たせる週末向けレシピ":  "時間に余裕がある時向けに、蒸らしを長めに取って香りの立ち上がりを楽しむ内容です。",

		"週末限定でドリッパーが10%オフ":   "週末限定で定番ドリッパーが割引になり、初めて器具を買う人でも手を出しやすいセール情報です。",
		"初回購入向けの送料無料キャンペーン":  "初回注文だけ送料を無料にすることで、豆や器具の試し買いを後押しする内容です。",
		"深煎りセットのまとめ買い値引き":    "深煎り中心のセットを複数買うと値引きが大きくなる、まとめ買い向けキャンペーンです。",
		"春の新生活向けコーヒー器具セール":   "新生活で器具をそろえる人向けに、ケトルやミルなどの入門機材をまとめて安くした内容です。",
		"ミルとケトルの同時購入で割引適用":   "単品よりも組み合わせ購入でお得になる、入門者向けの販促記事です。",
		"定期便スタート記念のクーポン配布":   "定期便開始に合わせて初回や数回分に使える割引コードを配布する施策の紹介です。",
		"アイスコーヒー器具の季節セール":    "気温上昇に合わせて、アイスコーヒー向け器具をまとめて訴求する季節セールです。",
		"店舗受け取り限定の豆セット特価":    "送料をかけずに店頭受け取りできる人向けに、豆セットを特価で出す企画です。",
		"レビュー投稿で次回使えるクーポン配布": "購入後レビューを書くと、次回の豆購入や器具購入に使える値引きを配る施策です。",
		"抽出スターターセットの期間限定セール": "これからハンドドリップを始める人向けに、必要最低限の器具をまとめたセットのセールです。",

		"駅前に小型ロースタリー併設店がオープン":   "駅前立地で立ち寄りやすく、焙煎機のある店内を見ながらコーヒーを楽しめる新店紹介です。",
		"朝営業に強いカフェの新店舗情報":       "出勤前でも使いやすい営業時間と提供スピードが強みの新店舗を紹介しています。",
		"自家製スイーツと相性が良い人気店":      "コーヒー単体だけでなく、自家製スイーツとの組み合わせに強い店の紹介記事です。",
		"深夜まで営業する作業向けカフェ":       "電源や席間隔、閉店時間の遅さを重視する人に向けた作業系カフェの紹介です。",
		"豆の量り売りに対応した地域密着店":      "常連客が豆を少量ずつ買いやすく、相談しながら選べる量り売り対応店の内容です。",
		"静かな空間でハンドドリップを味わえる店":   "会話よりも抽出の香りや静かな空間を楽しみたい人向けの店舗紹介です。",
		"テイクアウト需要に強いスタンド型ショップ":  "短時間で買えて持ち歩きやすい設計が強みのスタンド型店舗の内容です。",
		"焙煎体験イベントを行う店舗の紹介":      "焙煎見学や体験イベントを定期的に行い、学びのある来店動機を作っている店舗の紹介です。",
		"地方ロースターの豆を週替わりで出す店":    "毎週違うロースターの豆を試せる、セレクト型の魅力を持つ店舗を紹介しています。",
		"駅近で朝の回転が速いエスプレッソバーの紹介": "通勤前でも短時間で利用でき、エスプレッソ系ドリンクの回転が速い店の紹介です。",
	}

	if s, ok := m[title]; ok {
		return s
	}

	return strOrEmpty(raw)
}

func strOrEmpty(v string) string {
	return strings.TrimSpace(v)
}
