package main

import (
	"github.com/pgillich/prometheus_text-to-remote_write/cmd"

	"github.com/golang/glog"
)

func main() {
	defer glog.Flush()
	cmd.Execute()
}
