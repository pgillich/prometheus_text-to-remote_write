package util

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestBuildErrorsLists(t *testing.T) {
	err := getErrors(3)

	messages, trace := BuildErrorsLists(err)
	fmt.Printf("messages=%+v trace=%+v\n", messages, trace)
}

func TestGetFieldValue(t *testing.T) {
	err := getErrors(3)

	msg := getFieldValue(err, "msg")
	assert.NotNil(t, msg)
	assert.Equal(t, "ERROR 3", *msg)
}

func TestGetFieldValue_NotExists(t *testing.T) {
	err := getErrors(3)

	msg := getFieldValue(err, "XXX")
	assert.Nil(t, msg)
}

func TestGetFieldValue_Empty(t *testing.T) {
	err := getErrors(3)

	msg := getFieldValue(err, "")
	assert.Nil(t, msg)
}

func TestGetFieldValue_Nil(t *testing.T) {
	msg := getFieldValue(nil, "msg")
	assert.Nil(t, msg)
}

func TestGetFieldValue_NilError(t *testing.T) {
	var err error
	msg := getFieldValue(err, "msg")
	assert.Nil(t, msg)
}

func TestGetFieldValue_Error(t *testing.T) {
	err := fmt.Errorf("error")
	msg := getFieldValue(err, "msg")
	assert.Nil(t, msg)
}

func TestGetErrMsg(t *testing.T) {
	err := getErrors(3)

	msg := getErrMsg(err)
	assert.Equal(t, "ERROR 3", msg)
}

func TestGetErrMsg_Error(t *testing.T) {
	err := fmt.Errorf("error")

	msg := getErrMsg(err)
	assert.Equal(t, "error", msg)
}

func TestGetErrMsg_Nil(t *testing.T) {
	msg := getErrMsg(nil)
	assert.Equal(t, "", msg)
}

func TestGetErrMsg_NilError(t *testing.T) {
	var err error
	msg := getErrMsg(err)
	assert.Equal(t, "", msg)
}

func getErrors(e int) error {
	message := "ERROR " + strconv.Itoa(e)

	if e == 0 {
		return errors.New(message)
	}

	return errors.WithMessage(getErrors(e-1), message)
}
