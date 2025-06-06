package main

import (
	"context"
	"log"

	"cloud.google.com/go/logging"
	"github.com/caarlos0/env"
	"github.com/ramonmedeiros/rsvp/internal/app/rest"
	"github.com/ramonmedeiros/rsvp/internal/pkg/songs"
	"github.com/ramonmedeiros/rsvp/internal/pkg/spreadsheet"
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
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("%+v\n", err)
	}

	// Creates a client.
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

	spreadsheetService, err := spreadsheet.New(ctx, cfg.ServiceAccount, cfg.SpreadsheetID, logger)
	if err != nil {
		log.Fatalf("could not start spreadsheet service")
	}

	songService := songs.New(
		ctx,
		cfg.SpotifyClientID,
		cfg.SpotifyClientSecret,
		logger)

	restService := rest.New(
		spreadsheetService,
		songService,
		cfg.Port,
		cfg.ClientID,
		logger)
	restService.Serve()
}
