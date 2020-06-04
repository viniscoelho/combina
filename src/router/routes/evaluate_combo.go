package routes

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/combina/src/types"
	"github.com/gorilla/mux"
)

type evalCombo struct {
	cb types.LottoCombinator
}

func NewEvaluateComboHandler(cb types.LottoCombinator) *evalCombo {
	return &evalCombo{cb}
}

func (h evalCombo) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetID := vars[idRouteVar]

	// alternative w/ long url: ?values=1&values2...
	// r.ParseForm()
	// values := r.Form["values"]

	// alternative w/ one param as string: ?values=[1,2,...]
	queryValues := r.FormValue("values")
	var values []int
	err := json.Unmarshal([]byte(queryValues), &values)
	if err != nil {
		log.Printf("An error occured: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("internal server error"))
		return
	}

	result, err := h.cb.EvaluateCombination(targetID, values)
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

	content, err := json.Marshal(result)
	if err != nil {
		log.Printf("An error occured: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("internal server error"))
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.Write(content)
}
