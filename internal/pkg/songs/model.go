package songs

import (
	"context"

	"cloud.google.com/go/logging"
	"golang.org/x/oauth2"
)

type Client struct {
	clientID     string
	clientSecret string
	playlist     string
	logger       *logging.Logger
	cacheToken   *oauth2.Token
}

type Song struct {
	ID         string   `json:"id"`
	Artists    []Artist `json:"artists"`
	Images     []string `json:"images"`
	Endpoint   string   `json:"endpoint"`
	Name       string   `json:"name"`
	Popularity int      `json:"popularity"`
}

type Artist struct {
	Name string `json:"name"`
}

type API interface {
	Search(ctx context.Context, term string) ([]Song, error)
}
