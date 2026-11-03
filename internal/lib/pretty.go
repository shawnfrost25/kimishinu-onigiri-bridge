package lib

import (
	"net/http"

	"github.com/bytedance/sonic"
)

type Error struct {
	Status     string
	Message    string
	Additional map[string]string
}

func Pretty(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	s := sonic.ConfigDefault.NewEncoder(w)
	s.SetIndent("", "    ")
	_ = s.Encode(data)
}
