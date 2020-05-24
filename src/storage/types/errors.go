package types

import "fmt"

type EmptyStorageError struct{}

func (e EmptyStorageError) Error() string {
	return "storage is empty -- no combination was created yet"
}

type CombinationAlreadyExistsError struct{}

func (e CombinationAlreadyExistsError) Error() string {
	return "combination generated already"
}

type CombinationDoesNotExistError struct{}

func (e CombinationDoesNotExistError) Error() string {
	return "no combination registered with this ID"
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
