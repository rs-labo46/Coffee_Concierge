package usecase

import "coffee-spa/entity"

// 認証系の入力検証を行うinterface。
type AuthVal interface {
	// サインアップ時のemail / passwordを検証。
	Signup(email string, pw string) error

	// ログイン時の email / passwordを検証。
	Login(email string, pw string) error

	// 新しいパスワードの形式を検証。
	NewPw(pw string) error

	// verify/resetなどで使うtoken文字列を検証。
	Token(token string) error
}

// Source系APIの入力検証を行うinterface。
type SourceVal interface {
	// Source作成入力を検証。
	Create(in CreateSourceIn) error

	// Source詳細取得時のidを検証。
	Get(id uint) error

	// Source 一覧取得時のページング値を検証。
	List(limit int, offset int) error
}

// Item系APIの入力検証を行うinterface。
type ItemVal interface {
	// Item作成入力を検証。
	Create(in CreateItemIn) error

	// Item詳細取得時のidを検証。
	Get(id uint) error

	// Item一覧取得時の検索条件を検証。
	List(q entity.ItemQ) error

	// Top表示用Item取得時の件数を検証。
	Top(limit int) error
}

// Bean系APIの入力検証を行うinterface。
type BeanVal interface {
	// Bean作成入力を検証。
	Create(in CreateBeanIn) error

	//Bean更新入力を検証。
	Update(in UpdateBeanIn) error

	// Bean詳細取得時のidを検証。
	Get(id uint) error

	// Bean一覧取得時の検索条件を検証。
	List(in BeanListIn) error
}

// Recipe系APIの入力検証を行うinterface。
type RecipeVal interface {
	// Recipe作成入力を検証。
	Create(in CreateRecipeIn) error

	//Recipe更新入力を検証。
	Update(in UpdateRecipeIn) error

	// Recipe詳細取得時のidを検証。
	Get(id uint) error

	// Recipe一覧取得時の検索条件を検証。
	List(in RecipeListIn) error
}

// 検索フローやセッション系の入力検証を行うinterface。
type SearchVal interface {
	// 検索セッション開始入力を検証。
	StartSession(in StartSessionIn) error

	// 初回条件設定入力を検証。
	SetPref(in SetPrefIn) error

	// 会話の発話追加入力を検証。
	AddTurn(in AddTurnIn) error

	// 条件の部分更新入力を検証。
	PatchPref(in PatchPrefIn) error

	// セッション詳細取得入力を検証。
	GetSession(in GetSessionIn) error

	// 履歴一覧取得入力を検証。
	ListHistory(in ListHistoryIn) error

	// セッション終了入力を検証。
	CloseSession(in CloseSessionIn) error
}

// 保存済み提案系の入力検証を行うinterface。
type SavedVal interface {
	// 提案保存入力を検証。
	Save(in SaveSuggestionIn) error

	// 保存済み提案一覧取得入力を検証。
	List(in ListSavedIn) error

	// 保存済み提案削除入力を検証。
	Delete(in DeleteSavedIn) error
}

// 監査ログ一覧取得の入力検証を行う interface。
type AuditVal interface {
	// 監査ログ一覧取得入力を検証。
	List(in AuditListIn) error
}
