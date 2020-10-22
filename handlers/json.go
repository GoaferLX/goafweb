package handlers

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

// writeJSON is a helper function for writing a JSON response to http.ResponseWriter.
func writeJson(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	switch data := data.(type) {
	case error:
		w.Write([]byte(data.Error()))
		return
	default:
		stream, err := json.Marshal(data)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(stream)
	}
	w.WriteHeader(code)
}

// readJson is a helper method to read a JSON request.
func readJson(r *http.Request, dest interface{}) error {
	if header := r.Header.Get("Content-Type"); header != "" {
		if header != "application/json" {
			return errors.New("Media Type Not Supported: Content-Type header is not application/json")
		}
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return err
	}
	return nil

}
