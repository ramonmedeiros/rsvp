package rest

import (
	"net/http"

	"github.com/pkg/errors"

	"cloud.google.com/go/logging"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/ramonmedeiros/rsvp/internal/pkg/spreadsheet"
)

var (
	ErrGetEvents   = errors.New("could not retrieve events")
	ErrMarshalJSON = errors.New("could not marshal json")
)

type Server struct {
	spreadsheetService spreadsheet.API
	port               string
	logger             *logging.Logger
	clientID           string
}

type API interface {
	Serve()
}

func New(
	spreadsheetService spreadsheet.API,
	port string, clientID string,
	logger *logging.Logger) API {

	return &Server{
		spreadsheetService: spreadsheetService,
		port:               port,
		logger:             logger,
		clientID:           clientID,
	}
}

func (s *Server) getFamily(c *gin.Context) {
	code := c.Param("code")

	if code == "" {
		c.AbortWithError(
			http.StatusBadRequest,
			errors.New("code is required"))
		return
	}

	family, _, err := s.spreadsheetService.GetFamily(code)
	if err != nil {
		s.logger.Log(logging.Entry{
			Severity: logging.Error,
			Payload: map[string]interface{}{
				"message": "could not retrieve event",
				"code":    code,
			}},
		)
		c.AbortWithError(
			http.StatusInternalServerError,
			errors.New("could not retrieve family"))
		return
	}

	c.IndentedJSON(http.StatusOK, family)
}

func (s *Server) updateFamily(c *gin.Context) {
	code := c.Param("code")
	confirmed := c.Query("confirmed") == "true"
	comments := c.Query("comments")

	updatedFamily, err := s.spreadsheetService.ConfirmFamily(code, confirmed, comments)
	if err != nil {
		if err == spreadsheet.ErrAlreadyConfirmed {
			c.AbortWithError(
				http.StatusForbidden,
				errors.Wrap(err, "already confirmed"))
			return
		}

		c.AbortWithError(
			http.StatusInternalServerError,
			errors.Wrap(err, "could not update presence"))
		return
	}
	c.IndentedJSON(http.StatusCreated, updatedFamily)
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

	router.Run("0.0.0.0:" + s.port)
}
