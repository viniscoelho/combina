package types

func NewLottoInput(dto LottoInputDTO) (LottoInput, error) {
	if err := validateInputDTO(dto); err != nil {
		return LottoInput{}, err
	}

	if dto.Alias == nil {
		dto.Alias = new(string)
		*dto.Alias = "default"
	}

	li := LottoInput{
		NumGames:          *dto.NumGames,
		NumEachGame:       *dto.NumEachGame,
		FixedNumbers:      dto.FixedNumbers,
		MostSortedNumbers: dto.MostSortedNumbers,
		GameType:          *dto.GameType,
		Alias:             *dto.Alias,
	}

	return li, nil
}
