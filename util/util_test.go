package util

import (
	"fmt"
	"regexp"
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

func TestGetFieldString(t *testing.T) {
	err := getErrors(3)

	msg := getFieldString(err, "msg")
	assert.NotNil(t, msg)
	assert.Equal(t, "ERROR 3", *msg)
}

func TestGetFieldString_NotExists(t *testing.T) {
	err := getErrors(3)

	msg := getFieldString(err, "XXX")
	assert.Nil(t, msg)
}

func TestGetFieldString_Empty(t *testing.T) {
	err := getErrors(3)

	msg := getFieldString(err, "")
	assert.Nil(t, msg)
}

func TestGetFieldString_Nil(t *testing.T) {
	msg := getFieldString(nil, "msg")
	assert.Nil(t, msg)
}

func TestGetFieldString_NilError(t *testing.T) {
	var err error
	msg := getFieldString(err, "msg")
	assert.Nil(t, msg)
}

func TestGetFieldString_Error(t *testing.T) {
	err := fmt.Errorf("error")
	msg := getFieldString(err, "msg")
	assert.Nil(t, msg)
}

type sampleStruct struct {
	i int
	s string
}

type sampleFields struct {
	str           string
	strPtr        *string
	strPtrNil     *string
	intf          interface{}
	intfPtr       interface{}
	intfPtrNil    interface{}
	i0            int
	i             int
	mapIntInt0    map[int]int
	mapIntInt     map[int]int
	sampleStruct0 sampleStruct
	sampleStruct1 sampleStruct
}

func TestGetFieldString_Fields(t *testing.T) {
	strPtr := "strPtr"
	intfPtr := "intfPtr"

	testData := sampleFields{
		str:        "str",
		strPtr:     &strPtr,
		strPtrNil:  nil,
		intf:       "intf",
		intfPtr:    &intfPtr,
		intfPtrNil: nil,
		i:          1,
		mapIntInt: map[int]int{
			1: 10,
			2: 20,
		},
		sampleStruct1: sampleStruct{
			i: 2,
			s: "3",
		},
	}
	var msg *string

	msg = getFieldString(testData, "str")
	assert.NotNil(t, msg)
	assert.Equal(t, "str", *msg)

	msg = getFieldString(testData, "strPtr")
	assert.NotNil(t, msg)
	assert.Equal(t, "strPtr", *msg)

	msg = getFieldString(testData, "strPtrNil")
	assert.Nil(t, msg)

	msg = getFieldString(testData, "intf")
	assert.NotNil(t, msg)
	assert.Equal(t, "intf", *msg)

	msg = getFieldString(testData, "intfPtr")
	assert.NotNil(t, msg)
	assert.Equal(t, "intfPtr", *msg)

	msg = getFieldString(testData, "intfPtrNil")
	assert.Nil(t, msg)

	msg = getFieldString(testData, "sampleStruct0")
	assert.NotNil(t, msg)
	assert.Equal(t, "{i:0 s:}", *msg)

	msg = getFieldString(testData, "sampleStruct1")
	assert.NotNil(t, msg)
	assert.Equal(t, "{i:2 s:3}", *msg)

	msg = getFieldString(testData, "mapIntInt")
	assert.NotNil(t, msg)
	assert.Equal(t, "map[1:10 2:20]", *msg)

	msg = getFieldString(testData, "mapIntInt0")
	assert.NotNil(t, msg)
	assert.Equal(t, "map[]", *msg)

	msg = getFieldString(testData, "i0")
	assert.NotNil(t, msg)
	assert.Equal(t, "0", *msg)

	msg = getFieldString(testData, "i")
	assert.NotNil(t, msg)
	assert.Equal(t, "1", *msg)
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

func testStructMsg(t *testing.T,
	expected string,
	message string,
	keyValues ...interface{},
) {
	text := StructMsg(message, keyValues...)
	assert.Equal(t, expected, text)
}

func TestStructMsg(t *testing.T) {
	testStructMsg(t, `mess a ge{"s":"{i:4 s:str}","x":"3","y":"4"}`,
		"mess a ge",
		"x", 3,
		"y", "4",
		"s", sampleStruct{i: 4, s: "str"},
	)
}

func TestStructMsg_0(t *testing.T) {
	testStructMsg(t, `mess a ge{}`,
		"mess a ge",
	)
}

func TestStructMsg_1(t *testing.T) {
	testStructMsg(t, `mess a ge{"x":""}`,
		"mess a ge",
		"x",
	)
}

func TestStructMsg_1_nil(t *testing.T) {
	testStructMsg(t, `mess a ge{"<nil>":""}`,
		"mess a ge",
		nil,
	)
}

func TestStructMsg_2_nil(t *testing.T) {
	testStructMsg(t, `mess a ge{"x":"<nil>"}`,
		"mess a ge",
		"x", nil,
	)
}

func TestFormatErrorMessages_A(t *testing.T) {
	err := getErrors(2)
	err = errors.WithMessage(err, StructMsg("mess a ge",
		"x", 3,
		"y", "4",
		"n", nil,
		"s", sampleStruct{i: 4, s: "str"},
	))

	text := replaceCallLine(FormatErrorMessages(err))
	assert.Equal(t, `mess a ge{"n":"<nil>","s":"{i:4 s:str}","x":"3","y":"4"}; ERROR 2; ERROR 1; ERROR 0; 
    util_test.go#0:getErrors()
    util_test.go#0:getErrors()
    util_test.go#0:getErrors()
    util_test.go#0:TestFormatErrorMessages_A()`, text)
}

func TestFormatErrorMessages_B(t *testing.T) {
	err := getErrors(1)
	err = errors.WithMessage(err, StructMsg("mess a ge",
		"x", 3,
		"y", "4",
		"n", nil,
		"s", sampleStruct{i: 4, s: "str"},
	))
	err = errors.WithMessage(err, "ERROR 2")

	text := replaceCallLine(FormatErrorMessages(err))
	assert.Equal(t, `ERROR 2; mess a ge{"n":"<nil>","s":"{i:4 s:str}","x":"3","y":"4"}; ERROR 1; ERROR 0; 
    util_test.go#0:getErrors()
    util_test.go#0:getErrors()
    util_test.go#0:TestFormatErrorMessages_B()`, text)
}

func TestFormatErrorMessages_C(t *testing.T) {
	err := errors.New(StructMsg("mess a ge",
		"x", 3,
		"y", "4",
		"n", nil,
		"s", sampleStruct{i: 4, s: "str"},
	))
	err = errors.WithMessage(err, "ERROR 1")
	err = errors.WithMessage(err, "ERROR 2")

	text := replaceCallLine(FormatErrorMessages(err))
	assert.Equal(t, `ERROR 2; ERROR 1; mess a ge{"n":"<nil>","s":"{i:4 s:str}","x":"3","y":"4"}; 
    util_test.go#0:TestFormatErrorMessages_C()`, text)
}

func replaceCallLine(lines string) string {
	linePattern := regexp.MustCompile(`(?m)#\d*:`)
	return linePattern.ReplaceAllString(lines, "#0:")
}

func TestHttpProblem_A(t *testing.T) {
	err := getErrors(1)
	err = errors.WithMessage(err, StructMsg("mess a ge",
		"x", 3,
		"y", "4",
		"n", nil,
		"s", sampleStruct{i: 4, s: "str"},
	))
	err = errors.WithMessage(err, "ERROR 2")
	messages, trace := BuildErrorsLists(err)

	httpProblem := NewHttpProblem(404, messages, trace)

	resp, _ := httpProblem.MarshalPretty()
	respText := string(resp)

	assert.Equal(t, `{
  "type": "about:blank",
  "title": "Not Found",
  "status": 404,
  "detail": "ERROR 2",
  "Errors": [
    "ERROR 2",
    "mess a ge{\"n\":\"<nil>\",\"s\":\"{i:4 s:str}\",\"x\":\"3\",\"y\":\"4\"}",
    "ERROR 1",
    "ERROR 0"
  ]
}`, respText)
}
