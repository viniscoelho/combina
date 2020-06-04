package routes

import (
	"log"
	"net/http"

	"github.com/combina/src/types"
	"github.com/gorilla/mux"
)

type deleteCombo struct {
	cb types.LottoCombinator
}

func NewDeleteComboHandler(cb types.LottoCombinator) *deleteCombo {
	return &deleteCombo{cb}
}

func (h deleteCombo) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetID := vars[idRouteVar]

	err := h.cb.DeleteCombination(targetID)
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

	rw.WriteHeader(http.StatusNoContent)
}
