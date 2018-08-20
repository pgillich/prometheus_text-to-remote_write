package util

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"encoding/json"

	"github.com/golang/glog"
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
