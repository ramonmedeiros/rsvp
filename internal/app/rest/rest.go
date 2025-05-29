package rest

import (
	"github.com/pkg/errors"

	"cloud.google.com/go/logging"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/ramonmedeiros/rsvp/internal/pkg/songs"
	"github.com/ramonmedeiros/rsvp/internal/pkg/spreadsheet"
)

var (
	ErrGetEvents   = errors.New("could not retrieve events")
	ErrMarshalJSON = errors.New("could not marshal json")
)

type Server struct {
	spreadsheetService spreadsheet.API
	songService        songs.API
	port               string
	logger             *logging.Logger
	clientID           string
	authenticatedCodes []string
}

type API interface {
	Serve()
}

func New(
	spreadsheetService spreadsheet.API,
	songService songs.API,
	port string, clientID string,
	logger *logging.Logger) API {

	return &Server{
		spreadsheetService: spreadsheetService,
		songService:        songService,
		port:               port,
		logger:             logger,
		clientID:           clientID,
	}
}

func (s *Server) Serve() {
	router := gin.Default()

	// allow cors and authorization flow
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowCredentials = true
	router.Use(cors.New(config))

	// and endpoints
	router.GET("/family/:code", s.getFamily)
	router.POST("/family/:code", s.updateFamily)

	router.GET("/song/:code/:term", s.searchSong)

	router.Run("0.0.0.0:" + s.port)
}
