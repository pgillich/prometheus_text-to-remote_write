package util

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"encoding/json"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	
	"github.com/mitchellh/mapstructure"	
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

type dummyState struct {
	str   strings.Builder
	flags map[int]bool
}

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

// Flag not implemented
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

		result := map[string]interface{}{}

		errX := mapstructure.Decode(err, &result)
		fmt.Printf("Decode: %+x\n", errX)

		if stackErr, ok := err.(stackTracer); ok {
			trace = buildStackTraceList(stackErr)
		}

		if goStringerErr, ok := err.(fmt.GoStringer); ok {
			fmt.Println("GoStringer: ", goStringerErr.GoString())
		} else {
			fmt.Println("GoStringer: NO")
		}

		if stringerErr, ok := err.(fmt.Stringer); ok {
			fmt.Println("Stringer: ", stringerErr.String())
		} else {
			fmt.Println("Stringer: NO")
		}

		err2 := error(err)
		fmt.Println("err2: ", err2)

		if formatterErr, ok := err.(fmt.Formatter); ok {
			ds := dummyState{}
			/*
				formatterErr.Format(&ds, 's')
				fmt.Printf("Formatter: Format s; %s\n.\n", ds.str.String())

				ds = dummyState{}
				formatterErr.Format(&ds, 'v')
				fmt.Printf("Formatter: Format v; %s\n.\n", ds.str.String())
			*/
			ds = dummyState{flags: map[int]bool{
				'+': true,
			}}
			formatterErr.Format(&ds, 'v')
			fmt.Printf("Formatter: Format +v; %s\n.\n", ds.str.String())

			ds = dummyState{flags: map[int]bool{
				'#': true,
			}}
			formatterErr.Format(&ds, 'v')
			fmt.Printf("Formatter: Format #v; %s\n.\n", ds.str.String())

			fmt.Printf("Formatter: s %s\n.\n", err)
			fmt.Printf("Formatter: v %v\n.\n", err)
			fmt.Printf("Formatter: +v %+v\n.\n", err)
			fmt.Printf("Formatter: #v %#v\n.\n", err)
			fmt.Printf("Formatter: q %q\n.\n", err)
			fmt.Printf("Formatter: d %d\n.\n", err)
		} else {
			fmt.Println("Formatter: OK")
		}

		fmt.Println("Error: ", err.Error())

		m := err.Error()
		messages = append(messages, m)

		if causeErr, ok := err.(causer); ok {
			err = causeErr.Cause()
		} else {
			err = nil
		}
	}

	return messages, trace
}

func buildStackTraceList(trace stackTracer) []string {
	var traceList []string

	for t := range trace.StackTrace() { // Format
		traceList = append(traceList, fmt.Sprintf("%+v", t))
	}

	return traceList
}
