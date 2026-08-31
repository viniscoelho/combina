package routes

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"combina/src/types"
)

type generateFrom struct {
	cb types.LottoCombinator
}

func NewGenerateFromHandler(cb types.LottoCombinator) *generateFrom {
	return &generateFrom{cb}
}

type generateFromDTO struct {
	GameType          *string `json:"game_type"`
	NumGames          *int    `json:"num_games"`
	NumEach           *int    `json:"num_each"`
	ExistingGames     [][]int `json:"existing_games"`
	CarryGames        [][]int `json:"carry_games,omitempty"`
	FixedNumbers      []int   `json:"fixed_numbers,omitempty"`
	MostSortedNumbers []int   `json:"most_sorted,omitempty"`
	SaveImported      *bool   `json:"save_imported,omitempty"`
	Alias             *string `json:"alias,omitempty"`
}

func (h generateFrom) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("internal server error"))
		return
	}

	var dto generateFromDTO
	if err := json.Unmarshal(body, &dto); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("invalid json body"))
		return
	}

	if dto.GameType == nil || dto.NumGames == nil || dto.NumEach == nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(types.MissingFieldsError{}.Error()))
		return
	}

	alias := "default"
	if dto.Alias != nil {
		alias = *dto.Alias
	}

	inputDTO := types.LottoInputDTO{
		NumGames:          dto.NumGames,
		NumEachGame:       dto.NumEach,
		FixedNumbers:      dto.FixedNumbers,
		MostSortedNumbers: dto.MostSortedNumbers,
		GameType:          dto.GameType,
		Alias:             &alias,
	}

	lottoInput, err := types.NewLottoInput(inputDTO)
	if err != nil {
		log.Printf("validation error: %s", err)
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

	// Both imported (ExistingGames) and previously generated (CarryGames) games
	// seed the dedup map so newly generated games never collide with either set.
	seed := make([][]int, 0, len(dto.ExistingGames)+len(dto.CarryGames))
	seed = append(seed, dto.ExistingGames...)
	seed = append(seed, dto.CarryGames...)

	var rgg types.GameGenerator
	if len(lottoInput.MostSortedNumbers) != 0 {
		rgg = types.NewMostSortedGenerator(lottoInput, seed)
	} else {
		rgg = types.NewRandomGameGenerator(lottoInput, seed)
	}
	lotto := rgg.GenerateLottoCombination()

	// Persisted set = imported (only if save_imported, default true) + carried
	// (always) + newly generated. Order: imported, carried, generated.
	saveImported := dto.SaveImported == nil || *dto.SaveImported
	saved := make([][]int, 0, len(seed)+len(lotto.Numbers.Combination))
	if saveImported {
		saved = append(saved, dto.ExistingGames...)
	}
	saved = append(saved, dto.CarryGames...)
	saved = append(saved, lotto.Numbers.Combination...)
	lotto.Numbers.Combination = saved
	lotto.Numbers.Rows = len(saved)

	if err := h.cb.AddCombination(lotto); err != nil {
		log.Printf("storage error: %s", err)
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

	log.Printf("generate-from combination %s created", lotto.ID)

	content, err := json.Marshal(lotto)
	if err != nil {
		log.Printf("marshal error: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("internal server error"))
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Add("Location", lotto.ID)
	rw.WriteHeader(http.StatusCreated)
	rw.Write(content)
}
