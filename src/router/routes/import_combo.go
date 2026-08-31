package routes

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"combina/src/types"
)

type importCombo struct {
	cb types.LottoCombinator
}

func NewImportComboHandler(cb types.LottoCombinator) *importCombo {
	return &importCombo{cb}
}

type importResponse struct {
	Games     [][]int `json:"games"`
	GameType  string  `json:"game_type"`
	Discarded int     `json:"discarded"`
}

func (h importCombo) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("invalid multipart form"))
		return
	}

	gameType := r.FormValue("game_type")
	if _, ok := types.Games[gameType]; !ok {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("invalid or missing game_type"))
		return
	}

	separator := r.FormValue("separator")
	validSeparators := map[string]bool{",": true, ";": true, " ": true, "\t": true}
	if !validSeparators[separator] {
		separator = ","
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("missing file field"))
		return
	}
	defer file.Close()

	gameRange := types.Games[gameType]
	var games [][]int
	discarded := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Split(line, separator)
		nums := make([]int, 0, len(parts))
		valid := true

		for _, p := range parts {
			n, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil {
				valid = false
				break
			}
			if n < gameRange.Min || n > gameRange.Max {
				valid = false
				break
			}
			nums = append(nums, n)
		}

		if !valid || !types.IsNumEachValid(len(nums), gameType) {
			discarded++
			continue
		}

		// check for duplicate numbers within the game
		seen := make(map[int]bool, len(nums))
		for _, n := range nums {
			if seen[n] {
				valid = false
				break
			}
			seen[n] = true
		}
		if !valid {
			discarded++
			continue
		}

		sort.Ints(nums)
		games = append(games, nums)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("file scan error: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("internal server error"))
		return
	}

	resp := importResponse{
		Games:     games,
		GameType:  gameType,
		Discarded: discarded,
	}
	content, err := json.Marshal(resp)
	if err != nil {
		log.Printf("marshal error: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("internal server error"))
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.Write(content)
}
