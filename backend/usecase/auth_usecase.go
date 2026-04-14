package usecase

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"coffee-spa/entity"
	"coffee-spa/repository"
)

type AuthUC interface {
	Signup(in SignupIn) (SignupOut, error)
	VerifyEmail(in VerifyEmailIn) error
	Login(in LoginIn) (LoginOut, error)
	Refresh(in RefreshIn) (RefreshOut, error)
	ResendVerify(in ResendVerifyIn) error
	Logout(actor entity.Actor, refreshToken string) error
	ForgotPw(in ForgotPwIn) error
	ResetPw(in ResetPwIn) error
	Me(actor entity.Actor) (entity.User, error)
}

// 新規登録入力。
type SignupIn struct {
	Email    string
	Password string
}

// 新規登録結果。
type SignupOut struct {
	User entity.User
}

// メール認証入力。
type VerifyEmailIn struct {
	Token string
}

// ログイン入力。
type LoginIn struct {
	Email    string
	Password string
	UA       string
	IP       string
}

// ログイン結果。
type LoginOut struct {
	User         entity.User
	AccessToken  string
	RefreshToken string
}

// refresh入力。
type RefreshIn struct {
	RefreshToken string
	UA           string
	IP           string
}

// refresh結果。
type RefreshOut struct {
	User         entity.User
	AccessToken  string
	RefreshToken string
}

// 再設定メール発行入力。
type ForgotPwIn struct {
	Email string
}

// 認証メール再送入力。
type ResendVerifyIn struct {
	Email string
}

// 再設定。
type ResetPwIn struct {
	Token    string
	Password string
}

// password hash / compare。
type PwHasher interface {
	Hash(raw string) (string, error)
	Compare(hash string, raw string) error
}

// access token発行。
type TokenSvc interface {
	SignAccess(user entity.User) (string, error)
}

// refresh tokenの生成とhash化。
type RefreshSvc interface {
	New() (plain string, tokenHash string, err error)
	Hash(token string) string
}

// 現在の時刻取得。
type Clock interface {
	Now() time.Time
}

// family ID などの生成
type IDGen interface {
	New() string
}

// 認証メール送信。
type Mailer interface {
	SendVerifyEmail(to string, token string) error
	SendResetPwEmail(to string, token string) error
}

type authUsecase struct {
	users    repository.UserRepository
	verifies repository.EmailVerifyRepository
	resets   repository.PwResetRepository
	rts      repository.RtRepository
	audits   repository.AuditRepository

	val      AuthVal
	hasher   PwHasher
	tokenSvc TokenSvc
	rtSvc    RefreshSvc
	clock    Clock
	idGen    IDGen
	mailer   Mailer

	verifyTTL time.Duration
	resetTTL  time.Duration
	rtTTL     time.Duration
}

func NewAuthUsecase(
	users repository.UserRepository,
	verifies repository.EmailVerifyRepository,
	resets repository.PwResetRepository,
	rts repository.RtRepository,
	audits repository.AuditRepository,
	val AuthVal,
	hasher PwHasher,
	tokenSvc TokenSvc,
	rtSvc RefreshSvc,
	clock Clock,
	idGen IDGen,
	mailer Mailer,
	verifyTTL time.Duration,
	resetTTL time.Duration,
	rtTTL time.Duration,
) AuthUC {
	return &authUsecase{
		users:     users,
		verifies:  verifies,
		resets:    resets,
		rts:       rts,
		audits:    audits,
		val:       val,
		hasher:    hasher,
		tokenSvc:  tokenSvc,
		rtSvc:     rtSvc,
		clock:     clock,
		idGen:     idGen,
		mailer:    mailer,
		verifyTTL: verifyTTL,
		resetTTL:  resetTTL,
		rtTTL:     rtTTL,
	}
}

// 新規登録を行う。
func (u *authUsecase) Signup(in SignupIn) (SignupOut, error) {
	if err := u.val.Signup(in.Email, in.Password); err != nil {
		return SignupOut{}, err
	}

	passHash, err := u.hasher.Hash(in.Password)
	if err != nil {
		return SignupOut{}, repository.ErrInternal
	}

	now := u.clock.Now()

	user := &entity.User{
		Email:         in.Email,
		PassHash:      passHash,
		Role:          entity.RoleUser,
		TokenVer:      1,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := u.users.Create(user); err != nil {
		return SignupOut{}, err
	}

	verifyPlain := u.idGen.New()
	verifyHash := hashToken(verifyPlain)

	verify := &entity.EmailVerify{
		UserID:    user.ID,
		TokenHash: verifyHash,
		ExpiresAt: now.Add(u.verifyTTL),
		UsedAt:    nil,
	}

	if err := u.verifies.Create(verify); err != nil {
		return SignupOut{}, err
	}

	u.writeAudit(
		"auth.signup",
		&user.ID,
		"",
		"",
		map[string]string{
			"user_id": uintToStr(user.ID),
			"email":   user.Email,
		},
	)

	if u.mailer != nil {
		if err := u.mailer.SendVerifyEmail(user.Email, verifyPlain); err != nil {
			u.writeAudit(
				"auth.verify_email.mail_failed",
				&user.ID,
				"",
				"",
				map[string]string{
					"user_id": uintToStr(user.ID),
					"email":   user.Email,
				},
			)
		}
	}

	return SignupOut{
		User: *user,
	}, nil
}

// メール認証を完了させる。
func (u *authUsecase) VerifyEmail(in VerifyEmailIn) error {
	if err := u.val.Token(in.Token); err != nil {
		return err
	}

	tokenHash := hashToken(in.Token)

	v, err := u.verifies.GetByTokenHash(tokenHash)
	if err != nil {
		return err
	}

	now := u.clock.Now()

	if v.UsedAt != nil {
		return repository.ErrConflict
	}
	if !v.ExpiresAt.After(now) {
		return repository.ErrInvalidState
	}

	user, err := u.users.GetByID(v.UserID)
	if err != nil {
		return err
	}

	user.EmailVerified = true
	user.UpdatedAt = now

	if err := u.users.Update(user); err != nil {
		return err
	}

	if err := u.verifies.MarkUsed(v.ID, now); err != nil {
		return err
	}

	u.writeAudit(
		"auth.verify_email",
		&user.ID,
		"",
		"",
		map[string]string{
			"user_id": uintToStr(user.ID),
		},
	)

	return nil
}

// email / password認証後にaccess tokenとrefresh tokenを発行する。
func (u *authUsecase) Login(in LoginIn) (LoginOut, error) {
	if err := u.val.Login(in.Email, in.Password); err != nil {
		return LoginOut{}, err
	}

	user, err := u.users.GetByEmail(in.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return LoginOut{}, repository.ErrUnauthorized
		}
		return LoginOut{}, err
	}

	if err := u.hasher.Compare(user.PassHash, in.Password); err != nil {
		u.writeAudit(
			"auth.login.failed",
			&user.ID,
			in.IP,
			in.UA,
			map[string]string{
				"user_id": uintToStr(user.ID),
				"reason":  "password_mismatch",
			},
		)
		return LoginOut{}, repository.ErrUnauthorized
	}

	accessToken, err := u.tokenSvc.SignAccess(*user)
	if err != nil {
		return LoginOut{}, repository.ErrInternal
	}

	refreshPlain, refreshHash, err := u.rtSvc.New()
	if err != nil {
		return LoginOut{}, repository.ErrInternal
	}

	now := u.clock.Now()
	familyID := u.idGen.New()

	rt := &entity.Rt{
		UserID:       user.ID,
		FamilyID:     familyID,
		TokenHash:    refreshHash,
		ExpiresAt:    now.Add(u.rtTTL),
		RevokedAt:    nil,
		UsedAt:       nil,
		ReplacedByID: nil,
		CreatedAt:    now,
	}

	if err := u.rts.Create(rt); err != nil {
		return LoginOut{}, err
	}

	u.writeAudit(
		"auth.login",
		&user.ID,
		in.IP,
		in.UA,
		map[string]string{
			"user_id":   uintToStr(user.ID),
			"family_id": familyID,
		},
	)

	return LoginOut{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: refreshPlain,
	}, nil
}

// refresh rotationを行う。
// used済みtokenの再利用を検知したら、family全失効 + token_ver更新。
func (u *authUsecase) Refresh(in RefreshIn) (RefreshOut, error) {
	if err := u.val.Token(in.RefreshToken); err != nil {
		return RefreshOut{}, err
	}

	tokenHash := u.rtSvc.Hash(in.RefreshToken)

	current, err := u.rts.GetByTokenHash(tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return RefreshOut{}, repository.ErrUnauthorized
		}
		return RefreshOut{}, err
	}

	now := u.clock.Now()

	if current.RevokedAt != nil {
		return RefreshOut{}, repository.ErrUnauthorized
	}
	if !current.ExpiresAt.After(now) {
		return RefreshOut{}, repository.ErrUnauthorized
	}
	if current.UsedAt != nil {
		_ = u.rts.RevokeFamily(current.FamilyID, now)

		user, getErr := u.users.GetByID(current.UserID)
		if getErr == nil {
			_ = u.users.UpdateTokenVer(user.ID, user.TokenVer+1)
			u.writeAudit(
				"auth.refresh.reuse_detected",
				&user.ID,
				in.IP,
				in.UA,
				map[string]string{
					"user_id":   uintToStr(user.ID),
					"family_id": current.FamilyID,
				},
			)
		}

		return RefreshOut{}, repository.ErrUnauthorized
	}

	user, err := u.users.GetByID(current.UserID)
	if err != nil {
		return RefreshOut{}, err
	}

	newPlain, newHash, err := u.rtSvc.New()
	if err != nil {
		return RefreshOut{}, repository.ErrInternal
	}

	next := &entity.Rt{
		UserID:       user.ID,
		FamilyID:     current.FamilyID,
		TokenHash:    newHash,
		ExpiresAt:    now.Add(u.rtTTL),
		RevokedAt:    nil,
		UsedAt:       nil,
		ReplacedByID: nil,
		CreatedAt:    now,
	}

	if err := u.rts.Create(next); err != nil {
		return RefreshOut{}, err
	}

	current.UsedAt = &now
	current.ReplacedByID = &next.ID

	if err := u.rts.Update(current); err != nil {
		return RefreshOut{}, err
	}

	accessToken, err := u.tokenSvc.SignAccess(*user)
	if err != nil {
		return RefreshOut{}, repository.ErrInternal
	}

	u.writeAudit(
		"auth.refresh",
		&user.ID,
		in.IP,
		in.UA,
		map[string]string{
			"user_id":   uintToStr(user.ID),
			"family_id": current.FamilyID,
		},
	)

	return RefreshOut{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: newPlain,
	}, nil
}

// refresh tokenを起点にfamilyを失効する。
func (u *authUsecase) Logout(actor entity.Actor, refreshToken string) error {
	if err := u.val.Token(refreshToken); err != nil {
		return err
	}

	tokenHash := u.rtSvc.Hash(refreshToken)

	rt, err := u.rts.GetByTokenHash(tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.ErrUnauthorized
		}
		return err
	}

	if rt.UserID != actor.UserID {
		return repository.ErrForbidden
	}

	now := u.clock.Now()

	if err := u.rts.RevokeFamily(rt.FamilyID, now); err != nil {
		return err
	}

	u.writeAudit(
		"auth.logout",
		&actor.UserID,
		"",
		"",
		map[string]string{
			"user_id":   uintToStr(actor.UserID),
			"family_id": rt.FamilyID,
		},
	)

	return nil
}

// 再設定tokenを発行する。
func (u *authUsecase) ForgotPw(in ForgotPwIn) error {
	if err := u.val.Email(in.Email); err != nil {
		return err
	}
	user, err := u.users.GetByEmail(in.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// 存在有無は外に漏らさない。
			return nil
		}
		return err
	}

	now := u.clock.Now()
	resetPlain := u.idGen.New()
	resetHash := hashToken(resetPlain)

	reset := &entity.PwReset{
		UserID:    user.ID,
		TokenHash: resetHash,
		ExpiresAt: now.Add(u.resetTTL),
		UsedAt:    nil,
		CreatedAt: now,
	}

	if err := u.resets.Create(reset); err != nil {
		return err
	}

	u.writeAudit(
		"auth.password.forgot",
		&user.ID,
		"",
		"",
		map[string]string{
			"user_id": uintToStr(user.ID),
			"email":   user.Email,
		},
	)

	if u.mailer != nil {
		if err := u.mailer.SendResetPwEmail(user.Email, resetPlain); err != nil {
			u.writeAudit(
				"auth.password.forgot.mail_failed",
				&user.ID,
				"",
				"",
				map[string]string{
					"user_id": uintToStr(user.ID),
					"email":   user.Email,
				},
			)
		}
	}

	return nil
}

// 認証メールを再送する。
func (u *authUsecase) ResendVerify(in ResendVerifyIn) error {
	if err := u.val.Email(in.Email); err != nil {
		return err
	}
	user, err := u.users.GetByEmail(in.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			// 存在有無は外に漏らさない。
			return nil
		}
		return err
	}

	// すでに認証済みでも存在有無を曖昧に返す。
	if user.EmailVerified {
		return nil
	}

	now := u.clock.Now()
	verifyPlain := u.idGen.New()
	verifyHash := hashToken(verifyPlain)

	verify := &entity.EmailVerify{
		UserID:    user.ID,
		TokenHash: verifyHash,
		ExpiresAt: now.Add(u.verifyTTL),
		UsedAt:    nil,
	}

	if err := u.verifies.Create(verify); err != nil {
		return err
	}

	u.writeAudit(
		"auth.verify_email.resend",
		&user.ID,
		"",
		"",
		map[string]string{
			"user_id": uintToStr(user.ID),
			"email":   user.Email,
		},
	)

	if u.mailer != nil {
		if err := u.mailer.SendVerifyEmail(user.Email, verifyPlain); err != nil {
			u.writeAudit(
				"auth.verify_email.resend.mail_failed",
				&user.ID,
				"",
				"",
				map[string]string{
					"user_id": uintToStr(user.ID),
					"email":   user.Email,
				},
			)
		}
	}

	return nil
}

// password reset tokenを使ってpasswordを更新する。
func (u *authUsecase) ResetPw(in ResetPwIn) error {
	if err := u.val.Token(in.Token); err != nil {
		return err
	}
	if err := u.val.NewPw(in.Password); err != nil {
		return err
	}

	tokenHash := hashToken(in.Token)

	reset, err := u.resets.GetByTokenHash(tokenHash)
	if err != nil {
		return err
	}

	now := u.clock.Now()

	if reset.UsedAt != nil {
		return repository.ErrConflict
	}
	if !reset.ExpiresAt.After(now) {
		return repository.ErrInvalidState
	}

	user, err := u.users.GetByID(reset.UserID)
	if err != nil {
		return err
	}

	passHash, err := u.hasher.Hash(in.Password)
	if err != nil {
		return repository.ErrInternal
	}

	user.PassHash = passHash
	user.TokenVer = user.TokenVer + 1
	user.UpdatedAt = now

	if err := u.users.Update(user); err != nil {
		return err
	}

	if err := u.resets.MarkUsed(reset.ID, now); err != nil {
		return err
	}

	u.writeAudit(
		"auth.password.reset",
		&user.ID,
		"",
		"",
		map[string]string{
			"user_id": uintToStr(user.ID),
		},
	)

	return nil
}

// actorから自分のuserを返す。
func (u *authUsecase) Me(actor entity.Actor) (entity.User, error) {
	user, err := u.users.GetByID(actor.UserID)
	if err != nil {
		return entity.User{}, err
	}

	return *user, nil
}

// 監査ログ保存
func (u *authUsecase) writeAudit(
	typ string,
	userID *uint,
	ip string,
	ua string,
	meta map[string]string,
) {
	if u.audits == nil {
		return
	}

	raw, err := json.Marshal(meta)
	if err != nil {
		raw = []byte(`{}`)
	}

	_ = u.audits.Create(&entity.AuditLog{
		Type:      typ,
		UserID:    userID,
		IP:        ip,
		UA:        ua,
		Meta:      raw,
		CreatedAt: u.clock.Now(),
	})
}

// verify/reset用 tokenをsha256で固定。
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
func uintToStr(v uint) string {
	return strconv.FormatUint(uint64(v), 10)
}
