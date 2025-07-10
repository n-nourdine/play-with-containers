package util

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
)

func ToJSON(i any, rw http.ResponseWriter) error {
	rw.Header().Set("Content-Type", "application/json")
	e := json.NewEncoder(rw)
	return e.Encode(i)
}

func FromJSON(i any, r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(i)
}

func NewUUID() string {
	return uuid.NewString()
}
