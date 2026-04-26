package gemini

import (
	"coffee-spa/config"
	"coffee-spa/usecase"
)

func NewClient(c config.Cfg) usecase.GeminiClient {
	if c.GeminiUseMock {
		return NewMockClient()
	}

	client, err := NewService(c.GeminiAPIKey, c.GeminiModel)
	if err != nil {
		return NewMockClient()
	}

	return client
}
