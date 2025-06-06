package spreadsheet

import (
	"bytes"
	"encoding/json"
	"time"

	"cloud.google.com/go/logging"
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
	Songs           []string   `json:"songs,omitempty"`
	Comments        string     `json:"comments"`
	Confirmed       bool       `json:"confirmed"`
	Alergies        []Alergy   `json:"alergies"`
	ConfirmedAt     *time.Time `json:"confirmed_at,omitempty"`
}

type Alergy struct {
	Id    string `json:"id"`
	Count int    `json:"count"`
}

type API interface {
	GetFamily(code string) (*Family, int, error)
	ConfirmFamily(code string, updatedFamily Family) (*Family, error)
}

func (f *Family) GetAlergies() (string, error) {
	alergies := bytes.NewBuffer([]byte(""))
	err := json.
		NewEncoder(alergies).
		Encode(f.Alergies)
	return alergies.String(), err
}
