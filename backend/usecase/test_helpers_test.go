package usecase

import (
	"errors"
	"sync"
	"time"

	"coffee-spa/entity"
	"coffee-spa/repository"
)

// 固定時刻を返すClockモック。
type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

// 固定値を返すIDGenモック。
type fixedIDGen struct {
	id string
}

func (g fixedIDGen) New() string {
	return g.id
}

// AuthValモック。
type authValMock struct {
	signupErr error
	loginErr  error
	emailErr  error
	newPwErr  error
	tokenErr  error
}

func (m authValMock) Signup(email string, pw string) error { return m.signupErr }
func (m authValMock) Login(email string, pw string) error  { return m.loginErr }
func (m authValMock) Email(email string) error             { return m.emailErr }
func (m authValMock) NewPw(pw string) error                { return m.newPwErr }
func (m authValMock) Token(token string) error             { return m.tokenErr }

// SourceValモック。
type sourceValMock struct{ createErr, getErr, listErr error }

func (m sourceValMock) Create(in CreateSourceIn) error   { return m.createErr }
func (m sourceValMock) Get(id uint) error                { return m.getErr }
func (m sourceValMock) List(limit int, offset int) error { return m.listErr }

// ItemValモック。
type itemValMock struct{ createErr, getErr, listErr, topErr error }

func (m itemValMock) Create(in CreateItemIn) error { return m.createErr }
func (m itemValMock) Get(id uint) error            { return m.getErr }
func (m itemValMock) List(q entity.ItemQ) error    { return m.listErr }
func (m itemValMock) Top(limit int) error          { return m.topErr }

// BeanValモック。
type beanValMock struct{ createErr, updateErr, getErr, listErr error }

func (m beanValMock) Create(in CreateBeanIn) error { return m.createErr }
func (m beanValMock) Update(in UpdateBeanIn) error { return m.updateErr }
func (m beanValMock) Get(id uint) error            { return m.getErr }
func (m beanValMock) List(in BeanListIn) error     { return m.listErr }

// RecipeValモック。
type recipeValMock struct{ createErr, updateErr, getErr, listErr error }

func (m recipeValMock) Create(in CreateRecipeIn) error { return m.createErr }
func (m recipeValMock) Update(in UpdateRecipeIn) error { return m.updateErr }
func (m recipeValMock) Get(id uint) error              { return m.getErr }
func (m recipeValMock) List(in RecipeListIn) error     { return m.listErr }

// SearchValモック。
type searchValMock struct {
	startErr error
	setErr   error
	addErr   error
	patchErr error
	getErr   error
	listErr  error
	closeErr error
}

func (m searchValMock) StartSession(in StartSessionIn) error { return m.startErr }
func (m searchValMock) SetPref(in SetPrefIn) error           { return m.setErr }
func (m searchValMock) AddTurn(in AddTurnIn) error           { return m.addErr }
func (m searchValMock) PatchPref(in PatchPrefIn) error       { return m.patchErr }
func (m searchValMock) GetSession(in GetSessionIn) error     { return m.getErr }
func (m searchValMock) ListHistory(in ListHistoryIn) error   { return m.listErr }
func (m searchValMock) CloseSession(in CloseSessionIn) error { return m.closeErr }

// SavedValモック。
type savedValMock struct{ saveErr, listErr, deleteErr error }

func (m savedValMock) Save(in SaveSuggestionIn) error { return m.saveErr }
func (m savedValMock) List(in ListSavedIn) error      { return m.listErr }
func (m savedValMock) Delete(in DeleteSavedIn) error  { return m.deleteErr }

// AuditValモック。
type auditValMock struct{ listErr error }

func (m auditValMock) List(in AuditListIn) error { return m.listErr }

// Hasherモック。
type hasherMock struct {
	hash       string
	hashErr    error
	compareErr error
}

func (m hasherMock) Hash(raw string) (string, error)       { return m.hash, m.hashErr }
func (m hasherMock) Compare(hash string, raw string) error { return m.compareErr }

// TokenSvcモック。
type tokenSvcMock struct {
	token string
	err   error
}

func (m tokenSvcMock) SignAccess(user entity.User) (string, error) { return m.token, m.err }

// RefreshSvcモック。
type refreshSvcMock struct {
	plain  string
	hash   string
	newErr error
	hashFn func(string) string
}

func (m refreshSvcMock) New() (string, string, error) { return m.plain, m.hash, m.newErr }
func (m refreshSvcMock) Hash(token string) string {
	if m.hashFn != nil {
		return m.hashFn(token)
	}
	return m.hash
}

// Mailerモック。
type mailerMock struct {
	verifyTo    string
	verifyToken string
	verifyErr   error
	resetTo     string
	resetToken  string
	resetErr    error
}

func (m *mailerMock) SendVerifyEmail(to string, token string) error {
	m.verifyTo = to
	m.verifyToken = token
	return m.verifyErr
}
func (m *mailerMock) SendResetPwEmail(to string, token string) error {
	m.resetTo = to
	m.resetToken = token
	return m.resetErr
}

// UserRepositoryモック。
type userRepoMock struct {
	createFn         func(*entity.User) error
	getByIDFn        func(uint) (*entity.User, error)
	getByEmailFn     func(string) (*entity.User, error)
	updateFn         func(*entity.User) error
	updateTokenVerFn func(uint, int) error
}

func (m userRepoMock) Create(user *entity.User) error {
	if m.createFn != nil {
		return m.createFn(user)
	}
	return nil
}
func (m userRepoMock) GetByID(id uint) (*entity.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, ErrNotFound
}
func (m userRepoMock) GetByEmail(email string) (*entity.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(email)
	}
	return nil, ErrNotFound
}
func (m userRepoMock) Update(user *entity.User) error {
	if m.updateFn != nil {
		return m.updateFn(user)
	}
	return nil
}
func (m userRepoMock) UpdateTokenVer(userID uint, tokenVer int) error {
	if m.updateTokenVerFn != nil {
		return m.updateTokenVerFn(userID, tokenVer)
	}
	return nil
}

// EmailVerifyRepositoryモック。
type emailVerifyRepoMock struct {
	createFn         func(*entity.EmailVerify) error
	getByTokenHashFn func(string) (*entity.EmailVerify, error)
	markUsedFn       func(uint, time.Time) error
}

func (m emailVerifyRepoMock) Create(v *entity.EmailVerify) error {
	if m.createFn != nil {
		return m.createFn(v)
	}
	return nil
}
func (m emailVerifyRepoMock) GetByTokenHash(tokenHash string) (*entity.EmailVerify, error) {
	if m.getByTokenHashFn != nil {
		return m.getByTokenHashFn(tokenHash)
	}
	return nil, ErrNotFound
}
func (m emailVerifyRepoMock) MarkUsed(id uint, usedAt time.Time) error {
	if m.markUsedFn != nil {
		return m.markUsedFn(id, usedAt)
	}
	return nil
}
func (m emailVerifyRepoMock) DeleteExpired(now time.Time) error { return nil }

// PwResetRepositoryモック。
type pwResetRepoMock struct {
	createFn         func(*entity.PwReset) error
	getByTokenHashFn func(string) (*entity.PwReset, error)
	markUsedFn       func(uint, time.Time) error
}

func (m pwResetRepoMock) Create(r *entity.PwReset) error {
	if m.createFn != nil {
		return m.createFn(r)
	}
	return nil
}
func (m pwResetRepoMock) GetByTokenHash(tokenHash string) (*entity.PwReset, error) {
	if m.getByTokenHashFn != nil {
		return m.getByTokenHashFn(tokenHash)
	}
	return nil, ErrNotFound
}
func (m pwResetRepoMock) MarkUsed(id uint, usedAt time.Time) error {
	if m.markUsedFn != nil {
		return m.markUsedFn(id, usedAt)
	}
	return nil
}
func (m pwResetRepoMock) DeleteExpired(now time.Time) error { return nil }

// RtRepositoryモック。
type rtRepoMock struct {
	createFn         func(*entity.Rt) error
	getByTokenHashFn func(string) (*entity.Rt, error)
	updateFn         func(*entity.Rt) error
	revokeFamilyFn   func(string, time.Time) error
}

func (m rtRepoMock) Create(rt *entity.Rt) error {
	if m.createFn != nil {
		return m.createFn(rt)
	}
	return nil
}
func (m rtRepoMock) GetByTokenHash(tokenHash string) (*entity.Rt, error) {
	if m.getByTokenHashFn != nil {
		return m.getByTokenHashFn(tokenHash)
	}
	return nil, ErrNotFound
}
func (m rtRepoMock) Update(rt *entity.Rt) error {
	if m.updateFn != nil {
		return m.updateFn(rt)
	}
	return nil
}
func (m rtRepoMock) RevokeFamily(familyID string, revokedAt time.Time) error {
	if m.revokeFamilyFn != nil {
		return m.revokeFamilyFn(familyID, revokedAt)
	}
	return nil
}
func (m rtRepoMock) DeleteExpired(now time.Time) error { return nil }

// AuditRepositoryモック。
type auditRepoMock struct {
	mu        sync.Mutex
	logs      []*entity.AuditLog
	listFn    func(repository.AuditListQ) ([]entity.AuditLog, error)
	createErr error
}

func (m *auditRepoMock) Create(log *entity.AuditLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.createErr != nil {
		return m.createErr
	}
	cp := *log
	m.logs = append(m.logs, &cp)
	return nil
}
func (m *auditRepoMock) List(q repository.AuditListQ) ([]entity.AuditLog, error) {
	if m.listFn != nil {
		return m.listFn(q)
	}
	return nil, nil
}

// SourceRepositoryモック。
type sourceRepoMock struct {
	createFn  func(*entity.Source) error
	getByIDFn func(uint) (*entity.Source, error)
	listFn    func(repository.SourceListQ) ([]entity.Source, error)
}

func (m sourceRepoMock) Create(src *entity.Source) error {
	if m.createFn != nil {
		return m.createFn(src)
	}
	return nil
}
func (m sourceRepoMock) GetByID(id uint) (*entity.Source, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, ErrNotFound
}
func (m sourceRepoMock) List(q repository.SourceListQ) ([]entity.Source, error) {
	if m.listFn != nil {
		return m.listFn(q)
	}
	return nil, nil
}

// ItemRepositoryモック。
type itemRepoMock struct {
	createFn  func(*entity.Item) error
	getByIDFn func(uint) (*entity.Item, error)
	listFn    func(repository.ItemListQ) ([]entity.Item, error)
	topFn     func(int) (*entity.TopItems, error)
}

func (m itemRepoMock) Create(item *entity.Item) error {
	if m.createFn != nil {
		return m.createFn(item)
	}
	return nil
}
func (m itemRepoMock) GetByID(id uint) (*entity.Item, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, ErrNotFound
}
func (m itemRepoMock) List(q repository.ItemListQ) ([]entity.Item, error) {
	if m.listFn != nil {
		return m.listFn(q)
	}
	return nil, nil
}
func (m itemRepoMock) Top(limit int) (*entity.TopItems, error) {
	if m.topFn != nil {
		return m.topFn(limit)
	}
	return &entity.TopItems{}, nil
}
func (m itemRepoMock) SearchRelated(beanName string, roast entity.Roast, origin string, mood entity.Mood, method entity.Method, limit int, now time.Time) ([]entity.Item, error) {
	return nil, nil
}

// BeanRepositoryモック。
type beanRepoMock struct {
	createFn       func(*entity.Bean) error
	updateFn       func(*entity.Bean) error
	getByIDFn      func(uint) (*entity.Bean, error)
	listFn         func(repository.BeanListQ) ([]entity.Bean, error)
	searchByPrefFn func(entity.Pref, int) ([]entity.Bean, error)
}

func (m beanRepoMock) Create(bean *entity.Bean) error {
	if m.createFn != nil {
		return m.createFn(bean)
	}
	return nil
}
func (m beanRepoMock) Update(bean *entity.Bean) error {
	if m.updateFn != nil {
		return m.updateFn(bean)
	}
	return nil
}
func (m beanRepoMock) GetByID(id uint) (*entity.Bean, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, ErrNotFound
}
func (m beanRepoMock) List(q repository.BeanListQ) ([]entity.Bean, error) {
	if m.listFn != nil {
		return m.listFn(q)
	}
	return nil, nil
}
func (m beanRepoMock) SearchByPref(pref entity.Pref, limit int) ([]entity.Bean, error) {
	if m.searchByPrefFn != nil {
		return m.searchByPrefFn(pref, limit)
	}
	return nil, nil
}

// RecipeRepositoryモック。
type recipeRepoMock struct {
	createFn      func(*entity.Recipe) error
	updateFn      func(*entity.Recipe) error
	getByIDFn     func(uint) (*entity.Recipe, error)
	listFn        func(repository.RecipeListQ) ([]entity.Recipe, error)
	findPrimaryFn func(uint, entity.Method, entity.TempPref) (*entity.Recipe, error)
}

func (m recipeRepoMock) Create(recipe *entity.Recipe) error {
	if m.createFn != nil {
		return m.createFn(recipe)
	}
	return nil
}
func (m recipeRepoMock) Update(recipe *entity.Recipe) error {
	if m.updateFn != nil {
		return m.updateFn(recipe)
	}
	return nil
}
func (m recipeRepoMock) GetByID(id uint) (*entity.Recipe, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, ErrNotFound
}
func (m recipeRepoMock) List(q repository.RecipeListQ) ([]entity.Recipe, error) {
	if m.listFn != nil {
		return m.listFn(q)
	}
	return nil, nil
}
func (m recipeRepoMock) FindPrimaryByBean(beanID uint, method entity.Method, tempPref entity.TempPref) (*entity.Recipe, error) {
	if m.findPrimaryFn != nil {
		return m.findPrimaryFn(beanID, method, tempPref)
	}
	return nil, ErrNotFound
}

// SavedRepositoryモック。
type savedRepoMock struct {
	createFn func(*entity.SavedSuggestion) error
	listFn   func(repository.SavedListQ) ([]entity.SavedSuggestion, error)
	deleteFn func(uint, uint) error
	getFn    func(uint, uint) (*entity.SavedSuggestion, error)
}

func (m savedRepoMock) Create(saved *entity.SavedSuggestion) error {
	if m.createFn != nil {
		return m.createFn(saved)
	}
	return nil
}
func (m savedRepoMock) List(q repository.SavedListQ) ([]entity.SavedSuggestion, error) {
	if m.listFn != nil {
		return m.listFn(q)
	}
	return nil, nil
}
func (m savedRepoMock) DeleteByUserAndSuggestionID(userID uint, suggestionID uint) error {
	if m.deleteFn != nil {
		return m.deleteFn(userID, suggestionID)
	}
	return nil
}
func (m savedRepoMock) GetByUserAndSuggestionID(userID uint, suggestionID uint) (*entity.SavedSuggestion, error) {
	if m.getFn != nil {
		return m.getFn(userID, suggestionID)
	}
	return nil, ErrNotFound
}

// SessionRepositoryモック。
type sessionRepoMock struct {
	createSessionFn      func(*entity.Session) error
	getSessionByIDFn     func(uint) (*entity.Session, error)
	getGuestSessionFn    func(uint, string, time.Time) (*entity.Session, error)
	listHistoryFn        func(repository.HistoryQ) ([]entity.Session, error)
	closeSessionFn       func(uint) error
	createTurnFn         func(*entity.Turn) error
	listTurnsFn          func(uint) ([]entity.Turn, error)
	createPrefFn         func(*entity.Pref) error
	updatePrefFn         func(*entity.Pref) error
	getPrefFn            func(uint) (*entity.Pref, error)
	replaceSuggestionsFn func(uint, []entity.Suggestion) error
	listSuggestionsFn    func(uint) ([]entity.Suggestion, error)
	getSuggestionByIDFn  func(uint) (*entity.Suggestion, error)
}

func (m sessionRepoMock) CreateSession(session *entity.Session) error {
	if m.createSessionFn != nil {
		return m.createSessionFn(session)
	}
	return nil
}
func (m sessionRepoMock) GetSessionByID(id uint) (*entity.Session, error) {
	if m.getSessionByIDFn != nil {
		return m.getSessionByIDFn(id)
	}
	return nil, ErrNotFound
}
func (m sessionRepoMock) GetGuestSessionByID(id uint, sessionKeyHash string, now time.Time) (*entity.Session, error) {
	if m.getGuestSessionFn != nil {
		return m.getGuestSessionFn(id, sessionKeyHash, now)
	}
	return nil, ErrNotFound
}
func (m sessionRepoMock) ListHistory(q repository.HistoryQ) ([]entity.Session, error) {
	if m.listHistoryFn != nil {
		return m.listHistoryFn(q)
	}
	return nil, nil
}
func (m sessionRepoMock) CloseSession(id uint) error {
	if m.closeSessionFn != nil {
		return m.closeSessionFn(id)
	}
	return nil
}
func (m sessionRepoMock) CreateTurn(turn *entity.Turn) error {
	if m.createTurnFn != nil {
		return m.createTurnFn(turn)
	}
	return nil
}
func (m sessionRepoMock) ListTurns(sessionID uint) ([]entity.Turn, error) {
	if m.listTurnsFn != nil {
		return m.listTurnsFn(sessionID)
	}
	return nil, nil
}
func (m sessionRepoMock) CreatePref(pref *entity.Pref) error {
	if m.createPrefFn != nil {
		return m.createPrefFn(pref)
	}
	return nil
}
func (m sessionRepoMock) UpdatePref(pref *entity.Pref) error {
	if m.updatePrefFn != nil {
		return m.updatePrefFn(pref)
	}
	return nil
}
func (m sessionRepoMock) GetPrefBySessionID(sessionID uint) (*entity.Pref, error) {
	if m.getPrefFn != nil {
		return m.getPrefFn(sessionID)
	}
	return nil, ErrNotFound
}
func (m sessionRepoMock) ReplaceSuggestions(sessionID uint, suggestions []entity.Suggestion) error {
	if m.replaceSuggestionsFn != nil {
		return m.replaceSuggestionsFn(sessionID, suggestions)
	}
	return nil
}
func (m sessionRepoMock) ListSuggestions(sessionID uint) ([]entity.Suggestion, error) {
	if m.listSuggestionsFn != nil {
		return m.listSuggestionsFn(sessionID)
	}
	return nil, nil
}
func (m sessionRepoMock) GetSuggestionByID(id uint) (*entity.Suggestion, error) {
	if m.getSuggestionByIDFn != nil {
		return m.getSuggestionByIDFn(id)
	}
	return nil, ErrNotFound
}

// RateLimitStoreモック。
type rateLimitStoreMock struct {
	allowFn func(string, float64, float64, float64, time.Time) (bool, int, error)
}

func (m rateLimitStoreMock) Allow(key string, rate float64, capacity float64, cost float64, now time.Time) (bool, int, error) {
	if m.allowFn != nil {
		return m.allowFn(key, rate, capacity, cost, now)
	}
	return true, 0, nil
}

// PrefParserモック。
type prefParserMock struct{}

func (prefParserMock) ParseConditionDiff(in ParseConditionDiffIn) (ParseConditionDiffOut, error) {
	return ParseConditionDiffOut{}, nil
}

// ExplainSvcモック。
type explainSvcMock struct {
	buildFn func(BuildReasonsIn) ([]ReasonResult, bool, error)
}

func (m explainSvcMock) BuildReasons(in BuildReasonsIn) ([]ReasonResult, bool, error) {
	if m.buildFn != nil {
		return m.buildFn(in)
	}
	return nil, false, nil
}

// FollowupSvcモック。
type followupSvcMock struct {
	buildFn func(BuildQuestionsIn) ([]string, bool, error)
}

func (m followupSvcMock) BuildQuestions(in BuildQuestionsIn) ([]string, bool, error) {
	if m.buildFn != nil {
		return m.buildFn(in)
	}
	return nil, false, nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func errDummy(msg string) error { return errors.New(msg) }
