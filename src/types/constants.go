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
		"Seninha":      {1, 60},
	}
)
