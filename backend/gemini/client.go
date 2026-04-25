package gemini

import (
	"log"

	"coffee-spa/config"
	"coffee-spa/usecase"
)

func NewClient(c config.Cfg) usecase.GeminiClient {
	if c.GeminiUseMock {
		return NewMockClient()
	}

	client, err := NewService(c.GeminiAPIKey, c.GeminiModel)
	if err != nil {
		log.Printf("gemini service disabled. fallback to mock: %v", err)
		return NewMockClient()
	}

	return client
}
