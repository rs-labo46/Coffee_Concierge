package gemini

import (
	"coffee-spa/config"
	"coffee-spa/usecase"
	"fmt"
)

func NewClient(c config.Cfg) (usecase.GeminiClient, error) {
	if c.GeminiUseMock {
		return NewMockClient(), nil
	}

	client, err := NewService(c.GeminiAPIKey, c.GeminiModel)
	if err != nil {
		return nil, fmt.Errorf("new gemini service: %w", err)
	}

	return client, nil
}
