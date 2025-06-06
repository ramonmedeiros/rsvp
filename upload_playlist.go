package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/logging"
	"github.com/caarlos0/env"
	"github.com/ramonmedeiros/rsvp/internal/pkg/songs"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

// redirectURI is the OAuth redirect URI for the application.
// You must register an application at Spotify's developer portal
// and enter this value.
const redirectURI = "http://127.0.0.1:8080/callback"

var (
	auth        *spotifyauth.Authenticator
	state       = "abc123"
	songService *songs.Client
)

type config struct {
	ServiceAccount      string `env:"SERVICE_ACCOUNT,required"`
	SpreadsheetID       string `env:"SPREADSHEET_ID,required"`
	ClientID            string `env:"CLIENT_ID,required"`
	Port                string `env:"PORT" envDefault:"8080"`
	ProjectID           string `env:"PROJECT_ID,required"`
	SpotifyClientID     string `env:"SPOTIFY_CLIENT_ID,required"`
	SpotifyClientSecret string `env:"SPOTIFY_CLIENT_SECRET,required"`
	SpotifyPlaylist     string `env:"SPOTIFY_PLAYLIST" envDefault:"3r69De7EiLiRNUqM2xeST1"`
}

func main() {
	http.HandleFunc("/callback", completeAuth)
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("%+v\n", err)
	}

	ctx := context.Background()
	client, err := logging.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	logger := client.Logger(cfg.ProjectID)
	if err != nil {
		log.Fatalf("could not start logger")
	}

	songService = songs.New(context.Background(), cfg.SpotifyClientID, cfg.SpotifyClientSecret, logger)

	auth := songService.GetUserAuthLink(redirectURI)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", auth.AuthURL(state))
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	// use the token to get an authenticated client
	client := spotify.New(auth.Client(r.Context(), tok))

	// copy & paste from spreadshet
	for _, song := range []string{} {
		_, err := client.AddTracksToPlaylist(context.Background(), spotify.ID("3r69De7EiLiRNUqM2xeST1"), spotify.ID(song))
		if err != nil {
			log.Fatal(err)
		}
	}

}
