package usecase

import (
	"coffee-spa/entity"
)

// auth usecase テスト用の user repositoryモック。
type authUserRepoMock struct {
	createFn           func(u entity.User) (entity.User, error)
	getByEmailFn       func(email string) (entity.User, error)
	getByIDFn          func(id int64) (entity.User, error)
	setEmailVerifiedFn func(userID int64) error
	updatePassHashFn   func(userID int64, newHash string) error
	bumpTokenVerFn     func(userID int64) (int, error)
}

func (m *authUserRepoMock) Create(u entity.User) (entity.User, error) {
	return m.createFn(u)
}

func (m *authUserRepoMock) GetByEmail(email string) (entity.User, error) {
	return m.getByEmailFn(email)
}

func (m *authUserRepoMock) GetByID(id int64) (entity.User, error) {
	return m.getByIDFn(id)
}

func (m *authUserRepoMock) SetEmailVerified(userID int64) error {
	return m.setEmailVerifiedFn(userID)
}

func (m *authUserRepoMock) UpdatePassHash(userID int64, newHash string) error {
	return m.updatePassHashFn(userID, newHash)
}

func (m *authUserRepoMock) BumpTokenVer(userID int64) (int, error) {
	return m.bumpTokenVerFn(userID)
}

// email verify repositoryモック。
type authEvRepoMock struct {
	createFn           func(ev entity.EmailVerify) error
	getByTokenHashFn   func(hash string) (entity.EmailVerify, error)
	useFn              func(id int64) error
	revokeUnusedByUser func(userID int64) error
}

func (m *authEvRepoMock) Create(ev entity.EmailVerify) error {
	return m.createFn(ev)
}

func (m *authEvRepoMock) GetByTokenHash(hash string) (entity.EmailVerify, error) {
	return m.getByTokenHashFn(hash)
}

func (m *authEvRepoMock) Use(id int64) error {
	return m.useFn(id)
}

func (m *authEvRepoMock) RevokeUnusedByUser(userID int64) error {
	return m.revokeUnusedByUser(userID)
}

// password reset repositoryモック。
type authPwRepoMock struct {
	createFn           func(pw entity.PwReset) error
	getByTokenHashFn   func(hash string) (entity.PwReset, error)
	useFn              func(id int64) error
	revokeUnusedByUser func(userID int64) error
}

func (m *authPwRepoMock) Create(pw entity.PwReset) error {
	return m.createFn(pw)
}

func (m *authPwRepoMock) GetByTokenHash(hash string) (entity.PwReset, error) {
	return m.getByTokenHashFn(hash)
}

func (m *authPwRepoMock) Use(id int64) error {
	return m.useFn(id)
}

func (m *authPwRepoMock) RevokeUnusedByUser(userID int64) error {
	return m.revokeUnusedByUser(userID)
}

// refresh token repositoryモック。
type authRtRepoMock struct {
	createFn          func(rt entity.Rt) (entity.Rt, error)
	getByTokenHashFn  func(hash string) (entity.Rt, error)
	revokeFn          func(id int64) error
	markUsedFn        func(id int64) error
	setReplacedByFn   func(id int64, newID int64) error
	revokeByFamilyFn  func(familyID string) error
	revokeAllByUserFn func(userID int64) error
}

func (m *authRtRepoMock) Create(rt entity.Rt) (entity.Rt, error) {
	return m.createFn(rt)
}

func (m *authRtRepoMock) GetByTokenHash(hash string) (entity.Rt, error) {
	return m.getByTokenHashFn(hash)
}

func (m *authRtRepoMock) Revoke(id int64) error {
	return m.revokeFn(id)
}

func (m *authRtRepoMock) MarkUsed(id int64) error {
	return m.markUsedFn(id)
}

func (m *authRtRepoMock) SetReplacedBy(id int64, newID int64) error {
	return m.setReplacedByFn(id, newID)
}

func (m *authRtRepoMock) RevokeByFamilyID(familyID string) error {
	return m.revokeByFamilyFn(familyID)
}

func (m *authRtRepoMock) RevokeAllByUser(userID int64) error {
	return m.revokeAllByUserFn(userID)
}

// audit repositoryモック。
type authAuditRepoMock struct {
	createFn func(a entity.AuditLog) error
}

func (m *authAuditRepoMock) Create(a entity.AuditLog) error {
	return m.createFn(a)
}

// validatorモック。
type authValMock struct {
	signupFn    func(email string, pw string) error
	loginFn     func(email string, pw string) error
	emailOnlyFn func(email string) error
	newPwFn     func(pw string) error
}

func (m *authValMock) Signup(email string, pw string) error {
	return m.signupFn(email, pw)
}

func (m *authValMock) Login(email string, pw string) error {
	return m.loginFn(email, pw)
}

func (m *authValMock) EmailOnly(email string) error {
	return m.emailOnlyFn(email)
}

func (m *authValMock) NewPw(pw string) error {
	return m.newPwFn(pw)
}

// パスワードハッシュモック。
type pwHashMock struct {
	hashFn    func(pw string) (string, error)
	compareFn func(hash string, pw string) error
}

func (m *pwHashMock) Hash(pw string) (string, error) {
	return m.hashFn(pw)
}

func (m *pwHashMock) Compare(hash string, pw string) error {
	return m.compareFn(hash, pw)
}

// token makerモック。
type tokMock struct {
	newAccessFn   func(userID int64, role string, tokenVer int) (string, error)
	newCSRFFn     func() (string, error)
	newOpaqueFn   func() (string, error)
	newFamilyIDFn func() (string, error)
}

func (m *tokMock) NewAccess(userID int64, role string, tokenVer int) (string, error) {
	return m.newAccessFn(userID, role, tokenVer)
}

func (m *tokMock) NewCSRF() (string, error) {
	return m.newCSRFFn()
}

func (m *tokMock) NewOpaque() (string, error) {
	return m.newOpaqueFn()
}

func (m *tokMock) NewFamilyID() (string, error) {
	return m.newFamilyIDFn()
}

// mailerモック。
type mailerMock struct {
	sendVerifyFn func(email string, token string) error
	sendResetFn  func(email string, token string) error
}

func (m *mailerMock) SendVerify(email string, token string) error {
	return m.sendVerifyFn(email, token)
}

func (m *mailerMock) SendReset(email string, token string) error {
	return m.sendResetFn(email, token)
}

// rate limitモック。
type rateLimMock struct {
	allowSignupFn     func(ip string) (bool, int, error)
	allowLoginFn      func(emailHash string) (bool, int, error)
	allowRefreshFn    func(userID int64) (bool, int, error)
	allowResendIPFn   func(ip string) (bool, int, error)
	allowResendMailFn func(emailHash string) (bool, int, error)
	allowForgotIPFn   func(ip string) (bool, int, error)
	allowForgotMailFn func(emailHash string) (bool, int, error)
}

func (m *rateLimMock) AllowSignup(ip string) (bool, int, error) {
	return m.allowSignupFn(ip)
}

func (m *rateLimMock) AllowLogin(emailHash string) (bool, int, error) {
	return m.allowLoginFn(emailHash)
}

func (m *rateLimMock) AllowRefresh(userID int64) (bool, int, error) {
	return m.allowRefreshFn(userID)
}

func (m *rateLimMock) AllowResendIP(ip string) (bool, int, error) {
	return m.allowResendIPFn(ip)
}

func (m *rateLimMock) AllowResendMail(emailHash string) (bool, int, error) {
	return m.allowResendMailFn(emailHash)
}

func (m *rateLimMock) AllowForgotIP(ip string) (bool, int, error) {
	return m.allowForgotIPFn(ip)
}

func (m *rateLimMock) AllowForgotMail(emailHash string) (bool, int, error) {
	return m.allowForgotMailFn(emailHash)
}
