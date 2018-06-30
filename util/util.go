package util

import (
	"fmt"
	"os"
	"runtime"

	"github.com/golang/glog"
)

func FUNCTION_NAME() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
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
