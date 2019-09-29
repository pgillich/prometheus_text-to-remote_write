package util //nolint:golint

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"strings"

	"encoding/json"

	"emperror.dev/errors"
	"emperror.dev/errors/utils/keyval"
	"github.com/moogar0880/problems"
	log "github.com/sirupsen/logrus"
)

const (
	//nolint:golint
	MessagesDetailsSeparator = " | "
	//nolint:golint
	KeyCallStack = "callstack"
	//nolint:golint
	MaximumCallerDepth = 25
)

//nolint:golint
type StackTracer interface {
	StackTrace() errors.StackTrace
}

type contextLogFieldKey string

//nolint:golint
func NewTextLoggerFormatter(level log.Level) (*log.Logger, *AdvancedTextFormatter) {
	skipPackageNameForCaller := map[string]struct{}{}
	callerName := CallerFunctionName()
	if i := strings.LastIndex(callerName, "/"); i > 0 {
		skipPackageNameForCaller[callerName[:i+1]] = struct{}{}
	}

	textFormatter := NewAdvancedTextFormatter(skipPackageNameForCaller)
	logger := &log.Logger{
		Formatter:    textFormatter,
		Hooks:        make(log.LevelHooks),
		Level:        level,
		ReportCaller: true,
	}

	parentCallerHook := ParentCallerHook{1}
	logger.AddHook(&parentCallerHook)

	return logger, textFormatter
}

/*ErrorsHandleLogrus is a text formatter
callStackLinesSkip
If 0, call stack lines are NOT printed
If >0, call stack lines are printed, skipping the first set lines
	so, the main() will never be printed.
*/
func ErrorsHandleLogrus(logger *log.Logger, level log.Level, err error,
	callStackLinesSkip int, callStackInFields bool,
) {
	var entry *log.Entry
	fieldsRaw := keyval.ToMap(errors.GetDetails(err))
	fields := map[string]interface{}{}
	for key, value := range fieldsRaw {
		fields[key] = value
		if val := reflect.ValueOf(value); val.IsValid() {
			if val.Kind() != reflect.String {
				if json, jsonErr := json.Marshal(value); jsonErr == nil {
					fields[key] = string(json)
				}
			}
		}
	}

	var stackTracer StackTracer
	if callStackLinesSkip > 0 && errors.As(err, &stackTracer) {
		callStackLines := buildCallStackLines(stackTracer)
		if len(callStackLines) > callStackLinesSkip {
			callStackLines = callStackLines[:len(callStackLines)-callStackLinesSkip]
			if callStackInFields {
				if json, jsonErr := json.Marshal(callStackLines); jsonErr == nil {
					fields[KeyCallStack] = string(json)
				}
			} else {
				ctxCallStack := context.WithValue(context.Background(),
					contextLogFieldKey(KeyCallStack), callStackLines,
				)
				entry = logger.WithContext(ctxCallStack).WithFields(log.Fields(fields))
				entry.Log(level, err)
				return
			}
		}
	}

	entry = logger.WithFields(log.Fields(fields))
	entry.Log(level, err)
}

// nolint:golint
type AdvancedTextFormatter struct {
	log.TextFormatter
}

// nolint:golint
func NewAdvancedTextFormatter(skipPackageNameForCaller map[string]struct{}) *AdvancedTextFormatter {
	return &AdvancedTextFormatter{
		TextFormatter: log.TextFormatter{
			CallerPrettyfier: ModuleCallerPrettyfierDecorator(skipPackageNameForCaller),
			SortingFunc:      SortingFuncDecorator(AdvancedFieldOrder()),
			DisableColors:    true,
			QuoteEmptyFields: true,
		},
	}
}

// nolint:golint
func (f *AdvancedTextFormatter) Format(entry *log.Entry) ([]byte, error) {
	textPart, err := f.TextFormatter.Format(entry)

	if entry.Context != nil {
		if callStack := entry.Context.Value(contextLogFieldKey(KeyCallStack)); callStack != nil {
			if callStackLines, ok := callStack.([]string); ok {
				textPart = append(textPart, '\t')
				textPart = append(textPart,
					[]byte(strings.Join(callStackLines, "\n\t"))...,
				)
				textPart = append(textPart, '\n')
			}
		}
	}

	return textPart, err
}

// nolint:golint
func AdvancedFieldOrder() map[string]int {
	return map[string]int{
		log.FieldKeyLevel:       100, // first
		log.FieldKeyTime:        90,
		log.FieldKeyFunc:        80,
		log.FieldKeyMsg:         70,
		log.FieldKeyLogrusError: 60,
		log.FieldKeyFile:        50,
		KeyCallStack:            -2, // after normal fields (-1)
	}
}

// nolint:golint
func SortingFuncDecorator(fieldOrder map[string]int) func([]string) {
	return func(keys []string) {
		sorter := EntryFieldSorter{keys, fieldOrder}
		sort.Sort(sorter)
	}
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

/* Similar pull request:
https://github.com/sirupsen/logrus/pull/973
*/
// nolint:golint
type ParentCallerHook struct {
	ParentCount int
}

// nolint:golint
func (*ParentCallerHook) Levels() []log.Level {
	return log.AllLevels
}

// nolint:golint
func (h *ParentCallerHook) Fire(entry *log.Entry) error {
	if h.ParentCount <= 0 || entry.Caller == nil {
		return nil
	}

	pcs := make([]uintptr, MaximumCallerDepth)
	depth := runtime.Callers(2, pcs)
	frames := runtime.CallersFrames(pcs[:depth])

	var parentCount = MaximumCallerDepth - 1
	var f runtime.Frame
	var again bool
	for f, again = frames.Next(); again && parentCount > 0; f, again = frames.Next() {
		if f == *entry.Caller {
			parentCount = h.ParentCount
		}
		parentCount--
	}
	if again { // for loop exited by parentCount == 0
		entry.Caller = &f // nolint:scopelint
	}

	return nil
}

// nolint:golint
type EntryFieldSorter struct {
	items      []string
	fieldOrder map[string]int
}

func (sorter EntryFieldSorter) Len() int { return len(sorter.items) }
func (sorter EntryFieldSorter) Swap(i, j int) {
	sorter.items[i], sorter.items[j] = sorter.items[j], sorter.items[i]
}
func (sorter EntryFieldSorter) Less(i, j int) bool {
	iWeight := sorter.weight(i)
	jWeight := sorter.weight(j)
	if iWeight == jWeight {
		return sorter.items[i] < sorter.items[j]
	}
	return iWeight > jWeight
}
func (sorter EntryFieldSorter) weight(i int) int {
	if weight, ok := sorter.fieldOrder[sorter.items[i]]; ok {
		return weight
	}
	return -1
}

/*ModuleCallerPrettyfierDecorator similar pull requests:
https://github.com/sirupsen/logrus/pull/989
*/
func ModuleCallerPrettyfierDecorator(skipPackageNameForCaller map[string]struct{},
) func(frame *runtime.Frame) (string, string) {
	return func(frame *runtime.Frame) (string, string) {
		filePath := frame.File
		if i := strings.LastIndex(filePath, "/"); i >= 0 {
			filePath = filePath[i+1:]
		}

		functionName := frame.Function
		for prefix := range skipPackageNameForCaller {
			if strings.HasPrefix(functionName, prefix) {
				functionName = strings.TrimPrefix(functionName, prefix)
				break
			}
		}

		return functionName, fmt.Sprintf("%s:%d", filePath, frame.Line)
	}
}

// nolint:golint
func ErrorsFormatRfc7807(err error, status int, addCallStack bool) []byte {
	callStackLines := []string{}
	var stackTracer StackTracer
	if addCallStack && errors.As(err, &stackTracer) {
		callStackLines = buildCallStackLines(stackTracer)
	}

	httpProblem := NewHTTPProblem(
		status,
		digErrorsString(err),
		fmt.Sprintf("%s", err),
		buildDetailsList(errors.GetDetails(err)),
		callStackLines,
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
	Details   []string `json:"details,omitempty"`
	CallStack []string `json:"callstack,omitempty"`
}

// NewHTTPProblem makes a HTTPProblem instance
func NewHTTPProblem(status int, title string, message string, details []string, callStack []string) *HTTPProblem {
	p := HTTPProblem{
		DefaultProblem: problems.DefaultProblem{
			Type:   problems.DefaultURL,
			Title:  title,
			Status: status,
			Detail: message,
		},
		Details:   details,
		CallStack: callStack,
	}
	return &p
}

// nolint:golint
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

func buildCallStackLines(stackTracer StackTracer) []string {
	callStackLines := []string{}

	stackTrace := stackTracer.StackTrace()
	for _, t := range stackTrace {
		dsFunction := dummyState{flags: map[int]bool{'+': true}}
		t.Format(&dsFunction, 's')
		//functionName := TrimModuleNamePrefix(strings.Split(dsFunction.str.String(), "\n")[0])
		functionName := strings.Split(dsFunction.str.String(), "\n")[0]

		dsPath := dummyState{}
		t.Format(&dsPath, 's')
		path := dsPath.str.String()

		dsLine := dummyState{}
		t.Format(&dsLine, 'd')
		line := dsLine.str.String()

		callStackLines = append(callStackLines, fmt.Sprintf("%s() %s:%s", functionName, path, line))
	}

	return callStackLines
}

func digErrorsString(err error) string {
	if msg := GetFieldString(err, "msg"); msg != nil {
		return *msg
	}

	return err.Error()
}

// nolint:golint
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
