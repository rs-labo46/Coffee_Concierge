package middlewaremock

import "coffee-spa/entity"

type TokenVersionReader struct {
	GetByIDFn func(uint) (*entity.User, error)
}

func (m TokenVersionReader) GetByID(id uint) (*entity.User, error) {
	if m.GetByIDFn == nil {
		return &entity.User{ID: id, TokenVer: 1}, nil
	}
	return m.GetByIDFn(id)
}

type RateLimiter struct {
	Allowed    bool
	RetryAfter int
	Err        error
	LastKey    string
}

func (m *RateLimiter) AllowSignup(ip string) (bool, int, error)  { m.LastKey = ip; return m.allowed() }
func (m *RateLimiter) AllowLoginIP(ip string) (bool, int, error) { m.LastKey = ip; return m.allowed() }
func (m *RateLimiter) AllowLogin(emailHash string) (bool, int, error) {
	m.LastKey = emailHash
	return m.allowed()
}
func (m *RateLimiter) AllowResendIP(ip string) (bool, int, error) { m.LastKey = ip; return m.allowed() }
func (m *RateLimiter) AllowResendMail(emailHash string) (bool, int, error) {
	m.LastKey = emailHash
	return m.allowed()
}
func (m *RateLimiter) AllowForgotIP(ip string) (bool, int, error) { m.LastKey = ip; return m.allowed() }
func (m *RateLimiter) AllowForgotMail(emailHash string) (bool, int, error) {
	m.LastKey = emailHash
	return m.allowed()
}
func (m *RateLimiter) AllowRefreshToken(tokenHash string) (bool, int, error) {
	m.LastKey = tokenHash
	return m.allowed()
}
func (m *RateLimiter) AllowWS(key string) (bool, int, error) { m.LastKey = key; return m.allowed() }
func (m *RateLimiter) allowed() (bool, int, error) {
	if m.Err != nil {
		return false, 0, m.Err
	}
	if m.RetryAfter == 0 {
		m.RetryAfter = 60
	}
	return m.Allowed, m.RetryAfter, nil
}
