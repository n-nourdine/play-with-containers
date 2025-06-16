package util

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
)

// func Tojson(i any, rw io.Writer) error {
// 	e := json.NewEncoder(rw)
// 	return e.Encode(i)
// }

//OR

func Tojson(i any, rw http.ResponseWriter) error {
	rw.Header().Set("Content-Type", "application/json")
	e := json.NewEncoder(rw)
	return e.Encode(i)
}

func FromJson(i any, r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(i)
}

func NewUUID() string {
	return uuid.NewString()
}

// func Load(path string) error {
// 	file, err := os.Open(path)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()
// 	scanner := bufio.NewScanner(file)
// 	for scanner.Scan() {
// 		line := scanner.Text()
// 		if len(line) == 0 || line[0] == '#' {
// 			continue
// 		}
// 		parts := strings.SplitN(line, "=", 2)
// 		if len(parts) != 2 {
// 			continue
// 		}
// 		key := strings.TrimSpace(parts[0])
// 		value := strings.TrimSpace(parts[1])
// 		if key != "" {
// 			os.Setenv(key, value)
// 		}
// 	}
// 	return scanner.Err()
// }
