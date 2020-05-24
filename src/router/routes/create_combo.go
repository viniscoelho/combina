package routes

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/combina/src/storage/types"
)

type createCombo struct {
	cb types.Combination
}

func NewCreateComboHandler(cb types.Combination) *createCombo {
	return &createCombo{cb}
}

func (h createCombo) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// tkn := r.Header.Get("Authorization")
	// if len(tkn) == 0 {
	// 	log.Printf("Unauthorized request to resource: missing authorization header")
	// 	rw.WriteHeader(http.StatusUnauthorized)
	// 	rw.Write([]byte("unauthorized"))
	// 	return
	// }

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("An error occured: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("internal server error"))
		return
	}

	newDTO := types.LottoInputDTO{}
	err = json.Unmarshal(body, &newDTO)
	if err != nil {
		log.Printf("An error occured: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("internal server error"))
		return
	}

	err = validateLottoDTO(newDTO)
	if err != nil {
		log.Printf("An error occured: %s", err)

		switch err.(type) {
		case types.MissingFieldsError, types.InvalidDTOError:
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(err.Error()))
		default:
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte("internal server error"))
		}

		return
	}

	lotto := newLottoCombination(newDTO)
	err = h.cb.CreateCombination(lotto)
	if err != nil {
		log.Printf("An error occured: %s", err)

		switch err.(type) {
		case types.CombinationAlreadyExistsError:
			rw.WriteHeader(http.StatusConflict)
			rw.Write([]byte("combination already exists"))
		default:
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte("internal server error"))
		}

		return
	}

	rw.Header().Add("Location", lotto.ID)
	rw.WriteHeader(http.StatusCreated)
}
