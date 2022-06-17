package router

import (
	"net/http"

	"combina/src/router/routes"
	"combina/src/types"
	"github.com/gorilla/mux"
)

func CreateRoutes(cb types.LottoCombinator) *mux.Router {
	r := mux.NewRouter()

	r.Path("/combinations").
		Methods(http.MethodGet).
		Name("ListCombinations").
		Handler(routes.NewListComboHandler(cb))
	r.Path("/combinations").
		Methods(http.MethodPost).
		Name("CreateCombination").
		Handler(routes.NewCreateComboHandler(cb))
	r.Path("/combinations/{id}").
		Methods(http.MethodGet).
		Name("ReadCombination").
		Handler(routes.NewReadComboHandler(cb))
	r.Path("/combinations/{id}").
		Methods(http.MethodDelete).
		Name("DeleteCombination").
		Handler(routes.NewDeleteComboHandler(cb))
	r.Path("/combinations/evaluate/{id}").
		Methods(http.MethodGet).
		Name("EvaluateCombination").
		Handler(routes.NewEvaluateComboHandler(cb))

	return r
}
