package usecase

import (
	"log"
)

// ログへ出して token を確認できるようにする。
type LogMailer struct {
	feURL string
}

// LogMailerを作る
func NewLogMailer(feURL string) *LogMailer {
	return &LogMailer{
		feURL: feURL,
	}
}

// verify用メールをログへ出す
func (m *LogMailer) SendVerify(email string, token string) error {
	link := m.feURL + "/verify-email?token=" + token

	log.Printf("[MAIL][VERIFY] to=%s link=%s", email, link)

	return nil
}

// reset用メールをログへ出す
func (m *LogMailer) SendReset(email string, token string) error {
	link := m.feURL + "/reset-password?token=" + token

	log.Printf("[MAIL][RESET] to=%s link=%s", email, link)

	return nil
}
