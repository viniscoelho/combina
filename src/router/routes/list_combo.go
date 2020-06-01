package routes

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/combina/src/storage/types"
)

type listCombo struct {
	cb types.LottoCombinator
}

func NewListComboHandler(cb types.LottoCombinator) *listCombo {
	return &listCombo{cb}
}

func (h listCombo) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	gameType := r.FormValue("type")
	lotto, err := h.cb.ListCombinations(gameType)
	if err != nil {
		log.Printf("An error occured: %s", err)

		switch err.(type) {
		case types.GameTypeDoesNotExistError:
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(err.Error()))
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
