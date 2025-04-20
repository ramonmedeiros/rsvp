package spreadsheet

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/logging"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	sheets "google.golang.org/api/sheets/v4"
)

type Client struct {
	service       *sheets.Service
	spreadsheetID string
	logger        *logging.Logger
}

type Family struct {
	Name            string     `json:"name"`
	ExpectedGuests  []string   `json:"expected_guests,omitempty"`
	ConfirmedGuests []string   `json:"confirmed_guests,omitempty"`
	Comments        string     `json:"comments"`
	Confirmed       bool       `json:"confirmed"`
	ConfirmedAt     *time.Time `json:"confirmed_at,omitempty"`
}

type API interface {
	GetFamily(code string) (*Family, int, error)
	ConfirmFamily(code string, confirmedGuests []string, confirmed bool, comments string) (*Family, error)
}

const (
	ReadRange = "Sheet1!A:F"
)

var (
	ErrAlreadyConfirmed = errors.New("already confirmed")
)

func New(ctx context.Context, serviceAccount string, spreadsheetId string, logger *logging.Logger) (*Client, error) {
	serviceAccountString, err := base64.RawStdEncoding.DecodeString(serviceAccount)
	if err != nil {
		return nil, err
	}

	credentials, err := google.CredentialsFromJSON(ctx, []byte(serviceAccountString),
		sheets.DriveFileScope,
		sheets.DriveReadonlyScope,
		sheets.DriveScope,
		sheets.SpreadsheetsScope,
		sheets.SpreadsheetsReadonlyScope)
	if err != nil {
		return nil, err
	}

	service, err := sheets.NewService(
		ctx,
		option.WithCredentials(credentials),
	)
	if err != nil {
		logger.Log(logging.Entry{
			Severity: logging.Error,
			Payload: map[string]any{
				"message": "unable to retrieve sheets client",
				"error":   err,
			}},
		)
		return nil, err
	}

	return &Client{
		spreadsheetID: spreadsheetId,
		service:       service,
		logger:        logger,
	}, nil

}

func (c *Client) GetFamily(code string) (*Family, int, error) {
	resp, err := c.service.Spreadsheets.Values.Get(c.spreadsheetID, ReadRange).Do()
	if err != nil {
		c.logger.Log(logging.Entry{
			Severity: logging.Error,
			Payload: map[string]any{
				"message": "unable to retrieve data from sheet",
				"error":   err,
			}},
		)
	}

	for line, row := range resp.Values[1:] {

		if len(row) < 3 {
			c.logger.Log(logging.Entry{
				Severity: logging.Error,
				Payload: map[string]any{
					"message": "row does not have minimum number of columns",
					"row":     row,
				}})
			continue
		}

		if row[0].(string) == code {
			expectedGuests, ok := row[2].(string)
			if !ok {
				c.logger.Log(logging.Entry{
					Severity: logging.Error,
					Payload: map[string]any{
						"message": "could not cast expected guests to string",
						"row2":    row[2],
					}})
			}

			confirmedGuests := []string{}
			if len(row) >= 4 {
				confirmedGuestsString, ok := row[3].(string)
				if ok {
					confirmedGuests = strings.Split(confirmedGuestsString, ",")
				}
			}

			comments := ""
			if len(row) >= 5 {
				comments = row[4].(string)
			}

			confirmed := false
			if len(row) >= 6 {
				confirmedString := row[5].(string)
				if strings.EqualFold(confirmedString, "true") {
					confirmed = true
				}
			}

			var confirmedAt *time.Time
			if len(row) >= 7 {
				timeString, ok := row[6].(string)
				if ok {
					confirmedAtTime, err := time.Parse(time.DateTime, timeString)
					if err == nil {
						confirmedAt = &confirmedAtTime
					}
				}
			}

			return &Family{
				Name:            row[1].(string),
				ExpectedGuests:  strings.Split(expectedGuests, ","),
				ConfirmedGuests: confirmedGuests,
				Comments:        comments,
				Confirmed:       confirmed,
				ConfirmedAt:     confirmedAt,
			}, line + 2, nil
		}
	}

	return nil, 0, errors.New("not found")
}

func (c *Client) ConfirmFamily(code string, confirmedGuests []string, confirmed bool, comments string) (*Family, error) {
	family, line, err := c.GetFamily(code)
	if err != nil {
		c.logger.Log(logging.Entry{
			Severity: logging.Error,
			Payload: map[string]any{
				"message": "error while retrieving family",
				"error":   err,
			}},
		)
		return nil, err
	}

	if family.ConfirmedAt != nil {
		return nil, ErrAlreadyConfirmed
	}

	confirmedString := "false"
	if confirmed {
		confirmedString = "true"
	}

	vr := sheets.ValueRange{
		Range: getUpdateRange(line),
		Values: [][]any{{
			strings.Join(confirmedGuests, ","),
			comments,
			confirmedString,
			time.Now().Format(time.DateTime),
		}},
	}

	_, err = c.service.Spreadsheets.Values.BatchUpdate(c.spreadsheetID,
		&sheets.BatchUpdateValuesRequest{
			ValueInputOption: "USER_ENTERED",
			Data:             []*sheets.ValueRange{&vr},
		}).
		Do()
	if err != nil {
		c.logger.Log(logging.Entry{
			Severity: logging.Error,
			Payload: map[string]any{
				"message": "unable to update data from sheet",
				"error":   err,
			}},
		)
	}

	return nil, err
}

func getUpdateRange(lineNumber int) string {
	return fmt.Sprintf(`Sheet1!D%d:G%d`, lineNumber, lineNumber)
}
