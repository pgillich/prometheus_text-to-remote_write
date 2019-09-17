package util

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"encoding/json"

	"github.com/golang/glog"
)

func FunctionName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

func FunctionNameShort() string {
	pc, _, _, _ := runtime.Caller(1)
	longName := runtime.FuncForPC(pc).Name()
	return longName[strings.LastIndex(longName, "/")+1:]
}

func CallerFunctionName() string {
	pc, _, _, _ := runtime.Caller(2)
	return runtime.FuncForPC(pc).Name()
}

func PrintFatalf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
	os.Stdout.Sync()

	glog.Fatalf(format, args...)
}

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
