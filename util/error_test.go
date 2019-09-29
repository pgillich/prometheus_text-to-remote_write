package util

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"net/http"

	"github.com/stretchr/testify/assert"

	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
)

func replaceCallLine(lines string) string {
	linePattern := regexp.MustCompile(`(?m)\.go:\d*`)
	return linePattern.ReplaceAllString(lines, ".go:0")
}

func newWithDetails() error {
	_, err := strconv.Atoi("NO_NUMBER")
	return errors.WrapWithDetails(err, "MESSAGE%0", "K0_1", "V0_1", "K0_2", "V0_2")
}

func makeDeepErrors() error {
	type complexStruct struct {
		Text    string
		Integer int
		Bool    bool
		hidden  string
	}

	err := newWithDetails()
	err = errors.WithDetails(err, "K1_1", "V1_1", "K1_2", "V1_2")
	err = errors.WithMessage(err, "MESSAGE:2")
	err = errors.WithDetails(err,
		"K3=1", "V3=equal",
		"K3 2", "V3 space",
		"K3;3", "V3;semicolumn",
		"K3:3", "V3:column",
		"K3\"5", "V3\"doublequote",
		"K3%6", "V3%percent",
	)
	err = errors.WithMessage(err, "MESSAGE 4")
	err = errors.WithDetails(err,
		"K5_int", 12,
		"K5_bool", true,
		"K5_struct", complexStruct{Text: "text", Integer: 42, Bool: true, hidden: "hidden"},
	)

	return err
}

func TestGetFieldString(t *testing.T) {
	err := fmt.Errorf("fmt.Errorf")
	errStringPtr := GetFieldString(err, "s")
	assert.NotNil(t, errStringPtr)
	assert.Equal(t, "fmt.Errorf", *errStringPtr)

	err = errors.NewPlain("errors.NewPlain")
	errStringPtr = GetFieldString(err, "msg")
	assert.NotNil(t, errStringPtr)
	assert.Equal(t, "errors.NewPlain", *errStringPtr)

	type textStruct struct {
		text string
	}
	txt := textStruct{text: "textStruct.text"}
	txtStrPtr := GetFieldString(txt, "text")
	assert.NotNil(t, txtStrPtr)
	assert.Equal(t, "textStruct.text", *txtStrPtr)
}

func TestMessages(t *testing.T) {
	err := makeDeepErrors()

	text := fmt.Sprintf("%s", err)
	assert.Equal(t, `MESSAGE 4: MESSAGE:2: MESSAGE%0: strconv.Atoi: parsing "NO_NUMBER": invalid syntax`, text)
}

func TestDetails(t *testing.T) {
	err := makeDeepErrors()

	details := buildDetailsList(errors.GetDetails(err))
	assert.Equal(t, []string{
		`K0_1="V0_1"`,
		`K0_2="V0_2"`,
		`K1_1="V1_1"`,
		`K1_2="V1_2"`,
		`K3%3D1="V3%3Dequal"`,
		`K3%202="V3 space"`,
		`K3;3="V3;semicolumn"`,
		`K3:3="V3:column"`,
		`K3"5="V3\"doublequote"`,
		`K3%256="V3%25percent"`,
		`K5_int=12`,
		`K5_bool=true`,
		`K5_struct={"Text":"text","Integer":42,"Bool":true}`,
	}, details)
}

type LoggerMock struct {
	*log.Logger
	outBuf   *bytes.Buffer
	exitCode int
}

func (l *LoggerMock) exit(code int) {
	l.exitCode = code
}

func newTextLoggerMock() (*LoggerMock, *AdvancedTextFormatter) {
	logger, formatter := NewTextLoggerFormatter(log.InfoLevel)
	buf := new(bytes.Buffer)
	loggerMock := &LoggerMock{
		Logger:   logger,
		outBuf:   buf,
		exitCode: -1,
	}
	loggerMock.Out = buf
	loggerMock.ExitFunc = loggerMock.exit

	return loggerMock, formatter
}

func testSortingFuncDecorator(t *testing.T,
	fieldOrder map[string]int,
	expected []string,
	items []string,
) {
	sorter := SortingFuncDecorator(fieldOrder)
	sorter(items)
	assert.Equal(t, expected, items)
}

func TestSortingFuncDecorator(t *testing.T) {
	type testCase struct {
		fieldOrder map[string]int
		expected   []string
		items      []string
	}

	testCases := []testCase{
		{
			map[string]int{},
			[]string{"a", "b", "c", "x", "y", "z"},
			[]string{"a", "b", "c", "x", "y", "z"},
		},
		{
			map[string]int{},
			[]string{"a", "b", "c", "x", "y", "z"},
			[]string{"x", "b", "y", "a", "c", "z"},
		},
		{
			map[string]int{},
			[]string{"a", "b", "c", "x", "y", "z"},
			[]string{"z", "y", "x", "c", "b", "a"},
		},
		{
			map[string]int{"c": 1, "z": 2},
			[]string{"z", "c", "a", "b", "x", "y"},
			[]string{"z", "y", "x", "c", "b", "a"},
		},
		{
			map[string]int{"c": 1, "z": 2, "b": 3, "N": 5},
			[]string{"b", "z", "c", "a", "x", "y"},
			[]string{"z", "y", "x", "c", "b", "a"},
		},
		{
			map[string]int{"c": 2, "z": 2, "b": 2},
			[]string{"b", "c", "z", "a", "x", "y"},
			[]string{"z", "y", "x", "c", "b", "a"},
		},
	}

	for _, test := range testCases {
		testSortingFuncDecorator(t, test.fieldOrder, test.expected, test.items)
	}
}

func TestErrorsHandleLogrus(t *testing.T) {
	funcName := FunctionNameShort()
	loggerMock, formatter := newTextLoggerMock()
	formatter.TimestampFormat = "8008-08-08T08:08:08Z"

	err := makeDeepErrors()

	ErrorsHandleLogrus(loggerMock.Logger, log.ErrorLevel, err, 2, false)
	fmt.Printf("###\n%s\n###\n", loggerMock.outBuf.String())
	// nolint:lll
	assert.Equal(t, `level=error time="8008-08-08T08:08:08Z" func=`+funcName+` msg="MESSAGE 4: MESSAGE:2: MESSAGE%0: strconv.Atoi: parsing \"NO_NUMBER\": invalid syntax" file="error_test.go:0" K0_1=V0_1 K0_2=V0_2 K1_1=V1_1 K1_2=V1_2 K3 2="V3 space" K3"5="V3\"doublequote" K3%6="V3%percent" K3:3="V3:column" K3;3="V3;semicolumn" K3=1="V3=equal" K5_bool=true K5_int=12 K5_struct="{\"Text\":\"text\",\"Integer\":42,\"Bool\":true}"
	github.com/pgillich/prometheus_text-to-remote_write/util.newWithDetails() error_test.go:0
	github.com/pgillich/prometheus_text-to-remote_write/util.makeDeepErrors() error_test.go:0
	github.com/pgillich/prometheus_text-to-remote_write/`+funcName+`() error_test.go:0
`, replaceCallLine(loggerMock.outBuf.String()))
}

func TestErrorsHandleLogrus_CallStackInFields(t *testing.T) {
	funcName := FunctionNameShort()
	loggerMock, formatter := newTextLoggerMock()
	formatter.TimestampFormat = "8008-08-08T08:08:08Z"

	err := makeDeepErrors()

	ErrorsHandleLogrus(loggerMock.Logger, log.ErrorLevel, err, 2, true)
	fmt.Printf("###\n%s\n###\n", loggerMock.outBuf.String())
	// nolint:lll
	assert.Equal(t, `level=error time="8008-08-08T08:08:08Z" func=`+funcName+` msg="MESSAGE 4: MESSAGE:2: MESSAGE%0: strconv.Atoi: parsing \"NO_NUMBER\": invalid syntax" file="error_test.go:0" K0_1=V0_1 K0_2=V0_2 K1_1=V1_1 K1_2=V1_2 K3 2="V3 space" K3"5="V3\"doublequote" K3%6="V3%percent" K3:3="V3:column" K3;3="V3;semicolumn" K3=1="V3=equal" K5_bool=true K5_int=12 K5_struct="{\"Text\":\"text\",\"Integer\":42,\"Bool\":true}" callstack="[\"github.com/pgillich/prometheus_text-to-remote_write/util.newWithDetails() error_test.go:0\",\"github.com/pgillich/prometheus_text-to-remote_write/util.makeDeepErrors() error_test.go:0\",\"github.com/pgillich/prometheus_text-to-remote_write/util.TestErrorsHandleLogrus_CallStackInFields() error_test.go:0\"]"
`, replaceCallLine(loggerMock.outBuf.String()))
}

// nolint:lll
/*
func TestConsoleFormatting(t *testing.T) {
	err := makeDeepErrors()

	text := ErrorsFormatConsole(err)
	// nolint: lll
	assert.Equal(t,
		`MESSAGE 4: MESSAGE:2: MESSAGE%0 | K0_1="V0_1" K0_2="V0_2" K1_1="V1_1" K1_2="V1_2" K3%3D1="V3%3Dequal" K3%202="V3 space" K3;3="V3;semicolumn" K3:3="V3:column" K3"5="V3\"doublequote" K3%256="V3%25percent" K5_int=12 K5_bool=true K5_struct={"Text":"text","Integer":42,"Bool":true}
	github.com/pgillich/prometheus_text-to-remote_write/util.newWithDetails() error_test.go:0
	github.com/pgillich/prometheus_text-to-remote_write/util.makeDeepErrors() error_test.go:0
	github.com/pgillich/prometheus_text-to-remote_write/util.TestConsoleFormatting() error_test.go:0`,
		replaceCallLine(text),
	)
}
*/
func TestRfc7807Formatting(t *testing.T) {
	err := makeDeepErrors()

	bytes := ErrorsFormatRfc7807(err, http.StatusBadRequest, true)
	text := string(bytes)
	assert.Equal(t,
		`{
  "type": "about:blank",
  "title": "MESSAGE 4",
  "status": 400,
  "detail": "MESSAGE 4: MESSAGE:2: MESSAGE%0",
  "details": [
    "K0_1=\"V0_1\"",
    "K0_2=\"V0_2\"",
    "K1_1=\"V1_1\"",
    "K1_2=\"V1_2\"",
    "K3%3D1=\"V3%3Dequal\"",
    "K3%202=\"V3 space\"",
    "K3;3=\"V3;semicolumn\"",
    "K3:3=\"V3:column\"",
    "K3\"5=\"V3\\\"doublequote\"",
    "K3%256=\"V3%25percent\"",
    "K5_int=12",
    "K5_bool=true",
    "K5_struct={\"Text\":\"text\",\"Integer\":42,\"Bool\":true}"
  ],
  "stack": [
    "github.com/pgillich/prometheus_text-to-remote_write/util.newWithDetails() error_test.go:0",
    "github.com/pgillich/prometheus_text-to-remote_write/util.makeDeepErrors() error_test.go:0",
    "github.com/pgillich/prometheus_text-to-remote_write/util.TestRfc7807Formatting() error_test.go:0"
  ]
}`,
		replaceCallLine(text),
	)
}
