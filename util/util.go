package util

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"

	"encoding/json"

	"github.com/golang/glog"
	"github.com/pkg/errors"
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

	const depth = 64
	var pcs [depth]uintptr
	callers := runtime.Callers(3, pcs[:])

	for err != nil {
		if stackErr, ok := err.(stackTracer); ok {
			trace = buildStackTraceList(stackErr, callers)
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

	if msg := getFieldValue(err, "msg"); msg != nil {
		return *msg
	}

	return err.Error()
}

func getFieldValue(input interface{}, keyName string) *string {
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

	dataVal := reflect.Indirect(reflect.ValueOf(input))
	if dataVal.Kind() == reflect.Struct {
		typ := dataVal.Type()
		for i := 0; i < typ.NumField(); i++ {
			if typ.Field(i).Name == keyName {
				msg := dataVal.Field(i).String()
				return &msg
			}
		}
	}

	return nil
}
