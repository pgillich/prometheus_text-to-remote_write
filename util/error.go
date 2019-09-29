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
	MessagesDetailsSeparator = " | "
	ContextKeyStackTrace     = "stacktrace"
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

// nolint:golint
func TrimModuleNamePrefix(path string) string {
	if len(moduleNamePrefix) == 0 {
		moduleNameTrimming := reflect.TypeOf(AdvancedTextFormatter{}).PkgPath()
		// module full name + '/'
		moduleNamePrefix = moduleNameTrimming[:strings.LastIndex(moduleNameTrimming, "/")+1]
	}

	return strings.TrimPrefix(path, moduleNamePrefix)
}

// nolint:gochecknoglobals
var modulePathPrefix string

// nolint:golint
func TrimModulePathPrefix(path string) string {
	if len(modulePathPrefix) == 0 {
		_, modulePathTrimming, _, _ := runtime.Caller(0)
		// this package is under the module level
		modulePathTrimming = modulePathTrimming[:strings.LastIndex(modulePathTrimming, "/")]
		// module path + '/'
		modulePathPrefix = modulePathTrimming[:strings.LastIndex(modulePathTrimming, "/")+1]
	}

	return strings.TrimPrefix(path, modulePathPrefix)
}

// nolint:golint
type AdvancedTextFormatter struct {
	log.TextFormatter

	/* StackLinesSkip
	If 0, stack traces are NOT printed
	If >0, stack traces are printed, skipping the first set lines
		so, the main() will never be printed.
	*/
	StackLinesSkip int
}

// nolint:golint
func NewAdvancedTextFormatter(stackLinesSkip int) *AdvancedTextFormatter {
	return &AdvancedTextFormatter{
		TextFormatter: log.TextFormatter{
			CallerPrettyfier: ModuleCallerPrettyfier,
			SortingFunc:      SortingFuncDecorator(AdvancedFieldOrder()),
		},
		StackLinesSkip: stackLinesSkip,
	}
}

// nolint:golint
func (f *AdvancedTextFormatter) Format(entry *log.Entry) ([]byte, error) {
	textPart, err := f.TextFormatter.Format(entry)

	if entry.Context != nil {
		if stackValue := entry.Context.Value(ContextKeyStackTrace); stackValue != nil {
			if stackList, ok := stackValue.([]string); ok {
				stackList = stackList[:len(stackList)-f.StackLinesSkip]
				textPart = append(textPart, '\t')
				textPart = append(textPart,
					[]byte(strings.Join(stackList, "\n\t"))...,
				)
				textPart = append(textPart, '\n')
			}
		}
	}

	return textPart, err
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

// nolint:golint
func AdvancedFieldOrder() map[string]int {
	return map[string]int{
		log.FieldKeyLevel:       100,
		log.FieldKeyTime:        90,
		log.FieldKeyFunc:        80,
		log.FieldKeyMsg:         70,
		log.FieldKeyLogrusError: 60,
		log.FieldKeyFile:        50,
	}
}

// nolint:golint
func SortingFuncDecorator(fieldOrder map[string]int) func([]string) {
	return func(keys []string) {
		sorter := EntryFieldSorter{keys, fieldOrder}
		sort.Sort(sorter)
	}
}

// nolint:golint
func ModuleCallerPrettyfier(frame *runtime.Frame) (function string, file string) {
	function = TrimModuleNamePrefix(frame.Function)
	file = fmt.Sprintf("%s:%d", TrimModulePathPrefix(frame.File), frame.Line)
	return
}

// nolint:golint
func ErrorsHandleLogrus(logger *log.Logger, level log.Level, err error) {
	var entry *log.Entry

	var trace stackTracer
	if errors.As(err, &trace) {
		traceCtx := context.WithValue(context.Background(),
			ContextKeyStackTrace, buildStackTraceList(trace),
		)
		entry = logger.WithContext(traceCtx).WithFields(log.Fields(keyval.ToMap(errors.GetDetails(err))))
	} else {
		entry = logger.WithFields(log.Fields(keyval.ToMap(errors.GetDetails(err))))
	}
	entry.Log(level, err)
}

// nolint:golint
func ErrorsFormatConsole(err error) string {
	var str strings.Builder

	str.WriteString(fmt.Sprintf("%s", err))                                      //nolint:gosec
	str.WriteString(MessagesDetailsSeparator)                                    //nolint:gosec
	str.WriteString(strings.Join(buildDetailsList(errors.GetDetails(err)), " ")) //nolint:gosec

	var trace stackTracer
	if errors.As(err, &trace) {
		traceList := buildStackTraceList(trace)
		for _, line := range traceList {
			str.WriteString("\n\t") //nolint:gosec
			str.WriteString(line)   //nolint:gosec
		}
	}

	return str.String()
}

// nolint:golint
func ErrorsFormatRfc7807(err error, status int, addTrace bool) []byte {
	traceList := []string{}
	var trace stackTracer
	if addTrace && errors.As(err, &trace) {
		traceList = buildStackTraceList(trace)
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

func buildStackTraceList(trace stackTracer) []string {
	traceList := []string{}

	traces := trace.StackTrace()
	for _, t := range traces {
		ds0 := dummyState{flags: map[int]bool{'+': true}}
		t.Format(&ds0, 's')
		ds := dummyState{}
		ds.str.WriteString(strings.Split(ds0.str.String(), "\n")[0]) //nolint:gosec
		ds.str.WriteString("() ")                                    //nolint:gosec
		t.Format(&ds, 's')
		ds.str.WriteString(":") //nolint:gosec
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
