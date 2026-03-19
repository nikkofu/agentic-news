package analyze

import "context"

type Client interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

type NoopClient struct{}

func (NoopClient) Generate(ctx context.Context, prompt string) (string, error) {
	return "", nil
}
