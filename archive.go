package archive

import (
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	"context"
	"errors"
	"os"
)

var (
	ErrTokenNotSet = errors.New("The GOOGLE_API_KEY environment variable is not set!")
)

func NewService() (*youtube.Service, error) {
	key, err := APIKey()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	service, err := youtube.NewService(ctx,
		option.WithAPIKey(key),
	)

	return service, err
}

func APIKey() (string, error) {
	tkn := os.Getenv("GOOGLE_API_KEY")
	if tkn == "" {
		return tkn, ErrTokenNotSet
	} else {
		return tkn, nil
	}
}
