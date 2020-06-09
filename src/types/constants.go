package types

const (
	DatabaseName = "/lotto"
)

var (
	Games = map[string]MinMaxRange{
		"Lotofacil":    {1, 25},
		"Lotomania":    {0, 99},
		"Quina":        {1, 80},
		"Mega-Sena":    {1, 60},
		"Quina-Brasil": {1, 80},
		"Quininha":     {1, 80},
		"Seninha":      {1, 60},
	}

	Prizes = map[string][]int{
		"Lotofacil":    {11, 12, 13, 14, 15},
		"Lotomania":    {0, 15, 16, 17, 18, 19, 20},
		"Quina":        {2, 3, 4, 5},
		"Mega-Sena":    {4, 5, 6},
		"Quina-Brasil": {3, 4, 5},
		"Quininha":     {5},
		"Seninha":      {6},
	}
)
