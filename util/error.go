package util

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"encoding/json"

	"emperror.dev/errors"
	"emperror.dev/errors/utils/keyval"
	"github.com/moogar0880/problems"
	log "github.com/sirupsen/logrus"
)

const (
	SkipFirstStackLines      = 2
	MessagesDetailsSeparator = " | "
)

type stackTracer interface {
	StackTrace() errors.StackTrace
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

// nolint:gochecknoglobals
var moduleNamePrefix string

func TrimModuleNamePrefix(path string) string {
	if len(moduleNamePrefix) == 0 {
		moduleNamePrefix = reflect.TypeOf(TextFormatterTrace{}).PkgPath()
		moduleNamePrefix = moduleNamePrefix[:strings.LastIndex(moduleNamePrefix, "/")+1]

	}

	return strings.TrimPrefix(path, moduleNamePrefix)
}

// nolint:gochecknoglobals
var modulePathPrefix string

func TrimModulePathPrefix(path string) string {

	if len(modulePathPrefix) == 0 {
		_, modulePathPrefix, _, _ := runtime.Caller(0)
		modulePathPrefix = modulePathPrefix[:strings.LastIndex(modulePathPrefix, "/")]
		// nolint:staticcheck
		modulePathPrefix = modulePathPrefix[:strings.LastIndex(modulePathPrefix, "/")+1]
	}

	return strings.TrimPrefix(path, modulePathPrefix)
}

type TextFormatterTrace struct {
	log.TextFormatter
}

func (f *TextFormatterTrace) Format(entry *log.Entry) ([]byte, error) {
	textPart, err := f.TextFormatter.Format(entry)

	textPart = append(textPart, []byte("\n\tXXX\n\tYYY")...)

	return textPart, err
}

func CallerPrettyfierFunc(frame *runtime.Frame) (function string, file string) {
	type empty struct{}
	pkgPath := reflect.TypeOf(empty{}).PkgPath()
	fmt.Printf("%+v\n", pkgPath)

	pc, fileC, _, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	fmt.Printf("Entry=%+v pc=%+v file=%+v Name=%+v\n", f.Entry(), pc, fileC, f.Name())

	return frame.Function, pkgPath
}

func CallerPrettyfierFuncFile(frame *runtime.Frame) (function string, file string) {
	return frame.Function, fmt.Sprintf("%s:%d", frame.File, frame.Line)
}

func ErrorsHandleLogrus(logger *log.Logger, err error) {
	entry := logger.WithFields(log.Fields(keyval.ToMap(errors.GetDetails(err))))
	entry.Message = "MessagE"
	var trace stackTracer
	if errors.As(err, &trace) {
		traceList := buildStackTraceList(trace, SkipFirstStackLines)
		tracePart := strings.Join(traceList, "\n\t")
		fmt.Println(tracePart)
		entry.Infof("%s\n\t%s", err, tracePart)
	} else {
		entry.Info(err)
	}
}

func ErrorsFormatConsole(err error) string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("%s", err))
	str.WriteString(MessagesDetailsSeparator)
	str.WriteString(strings.Join(buildDetailsList(errors.GetDetails(err)), " "))

	var trace stackTracer
	if errors.As(err, &trace) {
		traceList := buildStackTraceList(trace, SkipFirstStackLines)
		for _, line := range traceList {
			str.WriteString("\n\t")
			str.WriteString(line)
		}
	}

	return str.String()
}

func ErrorsFormatRfc7807(err error, status int, addTrace bool) []byte {
	traceList := []string{}
	var trace stackTracer
	if addTrace && errors.As(err, &trace) {
		traceList = buildStackTraceList(trace, SkipFirstStackLines)
	}

	httpProblem := NewHTTPProblem(
		status,
		digErrorsString(err),
		fmt.Sprintf("%s", err),
		buildDetailsList(errors.GetDetails(err)),
		traceList,
	)

	resp, err := httpProblem.MarshalPretty()
	if err != nil {
		return []byte(err.Error())
	}
	return resp
}

// HTTPProblem is RFC-7807 comliant response
type HTTPProblem struct {
	problems.DefaultProblem
	Details []string `json:"details,omitempty"`
	Stack   []string `json:"stack,omitempty"`
}

// NewHTTPProblem makes a HTTPProblem instance
func NewHTTPProblem(status int, title string, message string, details []string, trace []string) *HTTPProblem {
	p := HTTPProblem{
		DefaultProblem: problems.DefaultProblem{
			Type:   problems.DefaultURL,
			Title:  title,
			Status: status,
			Detail: message,
		},
		Details: details,
		Stack:   trace,
	}
	return &p
}

func (httpProblem *HTTPProblem) MarshalPretty() ([]byte, error) {
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

// nolint:gochecknoglobals
var detailsKeyReplacer = strings.NewReplacer(
	"%", "%25",
	"=", "%3D",
	" ", "%20",
)

// nolint:gochecknoglobals
var detailsValueReplacer = strings.NewReplacer(
	"%", "%25",
	"=", "%3D",
)

func buildDetailsList(kvs []interface{}) []string {
	detailsList := []string{}
	if len(kvs)%2 == 1 {
		kvs = append(kvs, nil)
	}

	for i := 0; i < len(kvs); i += 2 {
		var value string
		valueJSON, valueErr := json.Marshal(kvs[i+1])
		if valueErr != nil {
			value = valueErr.Error()
		} else {
			value = string(valueJSON)
		}
		detailsList = append(detailsList, fmt.Sprintf("%s=%s",
			detailsKeyReplacer.Replace(fmt.Sprintf("%v", kvs[i])),
			detailsValueReplacer.Replace(value),
		))
	}

	return detailsList
}

func buildStackTraceList(trace stackTracer, skip int) []string {
	traceList := []string{}

	tracesAll := trace.StackTrace()
	traces := tracesAll[:len(tracesAll)-skip]
	for _, t := range traces {
		ds0 := dummyState{flags: map[int]bool{'+': true}}
		t.Format(&ds0, 's')
		ds := dummyState{}
		ds.str.WriteString(strings.Split(ds0.str.String(), "\n")[0])
		ds.str.WriteString("() ")
		t.Format(&ds, 's')
		ds.str.WriteString(":")
		t.Format(&ds, 'd')

		traceList = append(traceList, ds.str.String())
	}

	return traceList
}

func digErrorsString(err error) string {
	if msg := GetFieldString(err, "msg"); msg != nil {
		return *msg
	}

	return err.Error()
}

func GetFieldString(input interface{}, keyName string) *string {
	if input == nil {
		return nil
	}
	var inputVal reflect.Value
	if inputVal = reflect.ValueOf(input); !inputVal.IsValid() {
		return nil
	}

	if inputVal.Kind() == reflect.Ptr {
		if inputVal.IsNil() {
			return nil
		}
		inputVal = reflect.Indirect(inputVal)
		if !inputVal.IsValid() {
			return nil
		}
	}
	if inputVal.Kind() == reflect.Struct {
		typ := inputVal.Type()
		for i := 0; i < typ.NumField(); i++ {
			if typ.Field(i).Name == keyName {
				if field := inputVal.Field(i); field.IsValid() {
					fieldValue := fmt.Sprintf("%+v", field)
					return &fieldValue
				}
				break
			}
		}
	}

	return nil
}
