package util

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"

	"encoding/json"

	"github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/moogar0880/problems"
)

func FUNCTION_NAME() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

func FUNCTION_NAME_SHORT() string {
	pc, _, _, _ := runtime.Caller(1)
	longName := runtime.FuncForPC(pc).Name()
	return longName[strings.LastIndex(longName, "/")+1:]
}

func CALLER_FUNCTION_NAME() string {
	pc, _, _, _ := runtime.Caller(2)
	return runtime.FuncForPC(pc).Name()
}

func PrintFatalf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	os.Stdout.Sync()

	glog.Fatalf(format, args...)
}

func LogObjAsJson(level glog.Level, obj interface{}, name string, indent bool) {
	var obj_json []byte
	var err error

	if indent {
		obj_json, err = json.MarshalIndent(obj, "", "  ")
	} else {
		obj_json, err = json.Marshal(obj)
	}

	if err != nil {
		glog.V(level).Infof("%s: %s\n", name, err)
	} else {
		glog.V(level).Infof("%s: %s\n", name, obj_json)
	}
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

type causer interface {
	Cause() error
}

// dummyState is a dummy fmt.State implementation
type dummyState struct {
	str   strings.Builder
	flags map[int]bool
}

// Write pass trough strings.Builder
func (ds *dummyState) Write(b []byte) (n int, err error) {
	return ds.str.Write(b)
}

// Width not implemented
func (*dummyState) Width() (wid int, ok bool) {
	return 0, false
}

// Precision not implemented
func (*dummyState) Precision() (prec int, ok bool) {
	return 0, false
}

// Flag returns dummyState.flags
func (ds *dummyState) Flag(c int) bool {
	if f, ok := ds.flags[c]; ok {
		return f
	}

	return false
}

// BuildErrorsLists makes
func BuildErrorsLists(err error) ([]string, []string) {
	var messages []string
	var trace []string

	for err != nil {
		if stackErr, ok := err.(stackTracer); ok {
			trace = buildStackTraceList(stackErr, 2)
		}

		m := getErrMsg(err)
		messages = append(messages, m)

		if causeErr, ok := err.(causer); ok {
			err = causeErr.Cause()
		} else {
			err = nil
		}
	}

	return messages, trace
}

func buildStackTraceList(trace stackTracer, skip int) []string {
	var traceList []string

	tracesAll := trace.StackTrace()
	traces := tracesAll[:len(tracesAll)-skip]
	for _, t := range traces {
		ds := dummyState{}
		t.Format(&ds, 's')
		ds.str.WriteString("#")
		t.Format(&ds, 'd')
		ds.str.WriteString(":")
		t.Format(&ds, 'n')
		ds.str.WriteString("()")

		traceList = append(traceList, ds.str.String())
	}

	return traceList
}

func getErrMsg(err error) string {
	if err == nil {
		return ""
	}

	if msg := getFieldString(err, "msg"); msg != nil {
		return *msg
	}

	return err.Error()
}

func getFieldString(input interface{}, keyName string) *string {
	var inputVal reflect.Value
	if input != nil {
		inputVal = reflect.ValueOf(input)

		if inputVal.Kind() == reflect.Ptr && inputVal.IsNil() {
			input = nil
		}
	}

	if input == nil || !inputVal.IsValid() {
		return nil
	}

	if inputValue := reflect.ValueOf(input); inputValue.IsValid() {
		dataVal := reflect.Indirect(inputValue)
		if dataVal.Kind() == reflect.Struct {
			typ := dataVal.Type()
			for i := 0; i < typ.NumField(); i++ {
				if typ.Field(i).Name == keyName {
					return getValueString(dataVal.Field(i))
				}
			}
		}
	}

	return nil
}

func getValueString(field reflect.Value) *string {
	if field.IsValid() {
		if field.Kind() == reflect.Interface {
			field = field.Elem()
			if !field.IsValid() {
				return nil
			}
		}
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				return nil
			}
			field = reflect.Indirect(field)
			if !field.IsValid() {
				return nil
			}
		}
		msgValue := fmt.Sprintf("%+v", field)
		if len(msgValue) > 2 {
			c := msgValue[:1]
			if c == "\u003c" {
				msgValue = msgValue[1:]
			}
		}
		return &msgValue
	}

	return nil
}

// StructMsg builds a message{JSON} format from structured list
func StructMsg(message string, keyValues ...interface{}) string {
	values := map[string]string{}

	for k := 0; k < len(keyValues); k += 2 {
		value := ""
		if k+1 < len(keyValues) {
			value = fmt.Sprintf("%+v", keyValues[k+1])
		}
		values[fmt.Sprintf("%v", keyValues[k])] = value
	}

	jsonBytes, _ := JSONMarshalNoEscapeHTML(values)
	return fmt.Sprintf("%s%s", message, jsonBytes)
}

// JSONMarshalNoEscapeHTML marshalles without escaping '<' and '>', no endline
func JSONMarshalNoEscapeHTML(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	byteSlice := buffer.Bytes()
	if len(byteSlice) > 0 && byteSlice[len(byteSlice)-1] == '\n' {
		byteSlice = byteSlice[:len(byteSlice)-1]
	}
	return byteSlice, err
}

// FormatErrorMessages builds console warning message
func FormatErrorMessages(err error) string {
	sb := strings.Builder{}

	messages, trace := BuildErrorsLists(err)

	for _, message := range messages {
		sb.WriteString(message)
		sb.WriteString("; ")
	}
	if len(trace) > 0 {
		for _, frame := range trace {
			sb.WriteString("\n")
			sb.WriteString("    ")
			sb.WriteString(frame)
		}
	}

	return sb.String()
}

// WarnErrorMessages logs console warning message from errors
func WarnErrorMessages(err error) {
	glog.WarningDepth(1, FormatErrorMessages(err))
}

// HttpProblem is RFC-7807 comliant response
type HttpProblem struct {
	problems.DefaultProblem
	Errors []string
	// TODO provide it at debug level
	// Stack []string
}

// NewHttpProblem makes a HttpProblem instance
func NewHttpProblem(status int, messages []string, trace []string) *HttpProblem {
	p := HttpProblem{DefaultProblem: *problems.NewStatusProblem(status)}
	if len(messages) > 0 {
		p.DefaultProblem.Detail = messages[0]
	}
	p.Errors = messages

	// TODO provide it at debug level
	// p.Stack = trace

	return &p
}

func (httpProblem *HttpProblem) MarshalPretty() ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(httpProblem)
	byteSlice := buffer.Bytes()
	if len(byteSlice) > 0 && byteSlice[len(byteSlice)-1] == '\n' {
		byteSlice = byteSlice[:len(byteSlice)-1]
	}
	return byteSlice, err
}
