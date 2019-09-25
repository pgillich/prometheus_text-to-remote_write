package util //nolint:golint

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"encoding/json"

	"github.com/golang/glog"
)

//nolint:golint
func FunctionName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

//nolint:golint
func FunctionNameShort() string {
	pc, _, _, _ := runtime.Caller(1)
	longName := runtime.FuncForPC(pc).Name()
	return longName[strings.LastIndex(longName, "/")+1:]
}

//nolint:golint
func CallerFunctionName() string {
	pc, _, _, _ := runtime.Caller(2)
	return runtime.FuncForPC(pc).Name()
}

//nolint:golint
func PrintFatalf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	os.Stdout.Sync() //nolint:gosec,errcheck

	glog.Fatalf(format, args...)
}

//nolint:golint
func LogObjAsJSON(level glog.Level, obj interface{}, name string, indent bool) {
	var objJSON []byte
	var err error

	if indent {
		objJSON, err = json.MarshalIndent(obj, "", "  ")
	} else {
		objJSON, err = json.Marshal(obj)
	}

	if err != nil {
		glog.V(level).Infof("%s: %s\n", name, err)
	} else {
		glog.V(level).Infof("%s: %s\n", name, objJSON)
	}
}
