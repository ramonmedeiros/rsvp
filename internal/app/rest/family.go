package rest

import (
	"encoding/json"
	"net/http"
	"slices"

	"cloud.google.com/go/logging"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/ramonmedeiros/rsvp/internal/pkg/spreadsheet"
)

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

	if !slices.Contains(s.authenticatedCodes, code) {
		s.authenticatedCodes = append(s.authenticatedCodes, code)
	}

	c.IndentedJSON(http.StatusOK, family)
}

func (s *Server) updateFamily(c *gin.Context) {
	code := c.Param("code")

	input := spreadsheet.Family{}
	err := json.NewDecoder(c.Request.Body).Decode(&input)
	if err != nil {
		c.AbortWithError(
			http.StatusBadRequest,
			errors.Wrap(err, "could not parse input"))
		return
	}

	updatedFamily, err := s.spreadsheetService.ConfirmFamily(code, input.ConfirmedGuests, input.Confirmed, input.Songs, input.Comments)
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

	if !slices.Contains(s.authenticatedCodes, code) {
		s.authenticatedCodes = append(s.authenticatedCodes, code)
	}

	c.IndentedJSON(http.StatusCreated, updatedFamily)
}
