package routes

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/combina/src/storage/types"
)

type createCombo struct {
	cb types.LottoCombinator
}

func NewCreateComboHandler(cb types.LottoCombinator) *createCombo {
	return &createCombo{cb}
}

func (h createCombo) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	log.Printf("%s request received at %s", http.MethodPost, r.URL.RequestURI())

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("An error occured: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("internal server error"))
		return
	}

	newInputDTO := types.LottoInputDTO{}
	err = json.Unmarshal(body, &newInputDTO)
	if err != nil {
		log.Printf("An error occured during unmarshal: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("internal server error"))
		return
	}

	err = validateInputDTO(newInputDTO)
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

	var rgg types.RandomGameGenerator
	if len(newInputDTO.MostSortedNumbers) != 0 {
		rgg = types.NewMostSortedShuffle(newInputDTO)
	} else {
		rgg = types.NewRandomGameGenerator(newInputDTO)
	}

	lotto := rgg.GenerateLottoCombination()
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

	log.Printf("Combination %s successfully created", lotto.ID)
	rw.Header().Add("Location", lotto.ID)
	rw.WriteHeader(http.StatusCreated)
}
