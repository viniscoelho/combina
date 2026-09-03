package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"combina/src/types"
	"github.com/gorilla/mux"
)

type addSetHandler struct {
	cb types.LottoCombinator
}

func NewAddSetHandler(cb types.LottoCombinator) *addSetHandler {
	return &addSetHandler{cb}
}

type addSetRequest struct {
	Numbers []int `json:"numbers"`
}

type addSetResponse struct {
	Overlap types.OverlapReport `json:"overlap"`
	SetIndex int                `json:"set_index"`
}

func (h addSetHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)[idRouteVar]

	lotto, err := h.cb.FetchCombination(id)
	if err != nil {
		switch err.(type) {
		case types.CombinationDoesNotExistError:
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte("combination not found"))
		default:
			log.Printf("fetch error: %s", err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte("internal server error"))
		}
		return
	}

	var req addSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("invalid JSON body"))
		return
	}

	if len(req.Numbers) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("numbers must not be empty"))
		return
	}

	gameRange := types.Games[lotto.GameType]
	seen := make(map[int]bool, len(req.Numbers))
	for _, n := range req.Numbers {
		if n < gameRange.Min || n > gameRange.Max {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("number out of range for game type"))
			return
		}
		if seen[n] {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("duplicate number in set"))
			return
		}
		seen[n] = true
	}

	sorted := make([]int, len(req.Numbers))
	copy(sorted, req.Numbers)
	sort.Ints(sorted)

	overlap := types.AnalyseNewSet(sorted, lotto.Numbers.Combination)

	lotto.Numbers.Combination = append(lotto.Numbers.Combination, sorted)
	lotto.Numbers.Rows = len(lotto.Numbers.Combination)

	if err := h.cb.UpdateCombination(lotto); err != nil {
		log.Printf("update error: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("internal server error"))
		return
	}

	resp := addSetResponse{
		Overlap:  overlap,
		SetIndex: lotto.Numbers.Rows - 1,
	}
	content, err := json.Marshal(resp)
	if err != nil {
		log.Printf("marshal error: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("internal server error"))
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	rw.Write(content)
}

type overlapHandler struct {
	cb types.LottoCombinator
}

func NewOverlapHandler(cb types.LottoCombinator) *overlapHandler {
	return &overlapHandler{cb}
}

func (h overlapHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)[idRouteVar]

	lotto, err := h.cb.FetchCombination(id)
	if err != nil {
		switch err.(type) {
		case types.CombinationDoesNotExistError:
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte("combination not found"))
		default:
			log.Printf("fetch error: %s", err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte("internal server error"))
		}
		return
	}

	// optional filter: only report findings involving a specific set index
	filterParam := strings.TrimSpace(r.URL.Query().Get("set"))
	var filterIdx *int
	if filterParam != "" {
		n, err := strconv.Atoi(filterParam)
		if err == nil {
			filterIdx = &n
		}
	}

	report := types.AnalyseOverlaps(lotto.Numbers.Combination)

	if filterIdx != nil {
		var filtered []types.OverlapFinding
		for _, f := range report.Findings {
			if f.SetA == *filterIdx || f.SetB == *filterIdx {
				filtered = append(filtered, f)
			}
		}
		report.Findings = filtered
		report.Clean = len(filtered) == 0
	}

	content, err := json.Marshal(report)
	if err != nil {
		log.Printf("marshal error: %s", err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("internal server error"))
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.Write(content)
}
