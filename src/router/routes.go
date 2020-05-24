package router

import (
	"net/http"

	"github.com/combina/src/router/routes"
	"github.com/combina/src/storage/types"
	"github.com/gorilla/mux"
)

func CreateRoutes(cb types.Combination) *mux.Router {
	r := mux.NewRouter()

	// r.Path("/combinations").
	// 	Methods(http.MethodGet).
	// 	Name("ListCombinations").
	// 	Handler(routes.NewListCombinationsHandler(cb))
	r.Path("/combinations").
		Methods(http.MethodPost).
		Name("CreateCombination").
		Handler(routes.NewCreateComboHandler(cb))
	r.Path("/combinations/{id}").
		Methods(http.MethodGet).
		Name("ReadCombination").
		Handler(routes.NewReadComboHandler(cb))
	// r.Path("/combinations/{id}").
	// 	Methods(http.MethodDelete).
	// 	Name("DeleteCombination").
	// 	Handler(routes.NewDeleteCombinationHandler(cb))
	// r.Path("/combinations/evaluate/{id}").
	// 	Methods(http.MethodGet).
	// 	Name("EvaluateCombination").
	// 	Handler(routes.NewEvaluateCombinationHandler(cb))

	return r
}
