package rest

import (
	"net/http"
	"slices"

	"cloud.google.com/go/logging"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func (s *Server) searchSong(c *gin.Context) {
	code := c.Param("code")
	term := c.Param("term")

	if code == "" {
		c.AbortWithError(
			http.StatusUnauthorized,
			errors.New("code is required"))
		return
	}

	if term == "" {
		c.AbortWithError(
			http.StatusUnauthorized,
			errors.New("term is empty"))
		return
	}

	if !slices.Contains(s.authenticatedCodes, code) {
		_, _, err := s.spreadsheetService.GetFamily(code)
		if err != nil {
			s.logger.Log(logging.Entry{
				Severity: logging.Error,
				Payload: map[string]any{
					"message": "could not retrieve event",
					"code":    code,
				}},
			)
			c.AbortWithError(
				http.StatusInternalServerError,
				errors.New("could not retrieve family"))
			return
		}
	}

	songs, err := s.songService.Search(c.Request.Context(), term)
	if err != nil {
		s.logger.Log(logging.Entry{
			Severity: logging.Error,
			Payload: map[string]any{
				"message": "could search songs",
				"code":    code,
			}},
		)
		c.AbortWithError(
			http.StatusInternalServerError,
			errors.New("could search songs"))
		return
	}

	c.IndentedJSON(http.StatusOK, songs)
}
