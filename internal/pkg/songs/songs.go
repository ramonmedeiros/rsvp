package songs

import (
	"context"
	"errors"
	"log"
	"slices"
	"time"

	"cloud.google.com/go/logging"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

var (
	ErrAlreadyConfirmed = errors.New("already confirmed")
)

func New(ctx context.Context, clientID string, clientSecret string, logger *logging.Logger) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		logger:       logger,
		cacheToken:   nil,
	}
}

func (c *Client) auth(ctx context.Context) error {
	if c.cacheToken != nil && time.Now().Before(c.cacheToken.Expiry) {
		return nil
	}

	config := &clientcredentials.Config{
		ClientID:     c.clientID,
		ClientSecret: c.clientSecret,
		TokenURL:     spotifyauth.TokenURL,
	}

	token, err := config.Token(ctx)
	if err != nil {
		return err
	}
	c.cacheToken = token
	return nil
}

func (c *Client) Search(ctx context.Context, term string) ([]Song, error) {
	err := c.auth(ctx)
	if err != nil {
		return nil, err
	}

	client := spotify.New(spotifyauth.New().Client(ctx, c.cacheToken))

	result, err := client.Search(ctx, term, spotify.SearchTypeTrack, spotify.Limit(40))
	if err != nil {
		return nil, err
	}

	var allSongs []Song
	for page := 1; ; page++ {
		log.Printf("Page %d has %d tracks", page, len(result.Tracks.Tracks))

		allSongs = append(allSongs, MapSearchResultToSongs(result.Tracks)...)

		err = client.NextPage(ctx, result.Tracks)
		if err == spotify.ErrNoMorePages {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
	}

	slices.SortFunc(allSongs, func(a, b Song) int {
		return b.Popularity - a.Popularity
	})

	return allSongs, nil
}
