package routes

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/combina/src/types"
	"github.com/gorilla/mux"
)

type readCombo struct {
	cb types.LottoCombinator
}

func NewReadComboHandler(cb types.LottoCombinator) *readCombo {
	return &readCombo{cb}
}

func (h readCombo) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetID := vars[idRouteVar]

	lotto, err := h.cb.FetchCombination(targetID)
	if err != nil {
		log.Printf("An error occured: %s", err)

		switch err.(type) {
		case types.CombinationDoesNotExistError:
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte("combination does not exist"))
		default:
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte("internal server error"))
		}

		return
	}

	content, err := json.Marshal(lotto)
	if err != nil {
		log.Printf("An error occured: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("internal server error"))
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.Write(content)
}
