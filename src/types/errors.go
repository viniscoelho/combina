package types

import "fmt"

type CombinationAlreadyExistsError struct{}

func (e CombinationAlreadyExistsError) Error() string {
	return "combination generated already"
}

type CombinationDoesNotExistError struct{}

func (e CombinationDoesNotExistError) Error() string {
	return "no combination registered with this ID"
}

type GameTypeDoesNotExistError struct{}

func (e GameTypeDoesNotExistError) Error() string {
	return "no such game type registered"
}

type InvalidDTOError struct {
	Message string
}

func (e InvalidDTOError) Error() string {
	return fmt.Sprintf("dto contains errors: %s", e.Message)
}

type MissingFieldsError struct{}

func (e MissingFieldsError) Error() string {
	return "dto contains one or more missing fields"
}
