package util

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/pkg/errors"
)

func TestBuildErrorsLists(t *testing.T) {
	err := getErrors(3)

	messages, trace := BuildErrorsLists(err)
	fmt.Printf("messages=%+v trace=%+v\n", messages, trace)
}

func getErrors(e int) error {
	message := "ERROR " + strconv.Itoa(e)

	if e == 0 {
		return errors.New(message)
	}

	return errors.WithMessage(getErrors(e-1), message)
}
