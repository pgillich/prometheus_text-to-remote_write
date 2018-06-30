package cmd

import (
	goflag "flag"

	"github.com/golang/glog"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/pgillich/prometheus_text-to-remote_write/util"
)

const (
	OPT_RECEIVE_ON = "receive-on"
	OPT_WRITE_TO   = "write-to"

	OPT_COPYSTANDARDLOGTO      = "copystandardlogto"
	OPT_GLOG_COPYSTANDARDLOGTO = "glog." + OPT_COPYSTANDARDLOGTO

	DEFAULT_RECEIVE_ON = ":9099"
	DEFAULT_WRITE_TO   = "http://influxdb:8086/api/v1/prom/write?u=prom&p=prom&db=prometheus"
)

var RootCmd = &cobra.Command{
	Use:   "prometheus_text-to-remote_write",
	Short: "Prometheus text to remote_write",
	Long: `It's a microservice receiving Prometheus text exposition format and sending it to Prometheus remote_write.
It can be run as a service and as a converter.

Activating logging to stderr:
./prometheus_text-to-remote_write (...) --logtostderr=1
./prometheus_text-to-remote_write (...) --logtostderr
`,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		util.PrintFatalf("Runtime error: %+v\n", err)
	}
}

var goflagLogtostderr *goflag.Flag
var goflagV *goflag.Flag

func init() {
	goflag.CommandLine.VisitAll(func(goflag *goflag.Flag) {
		switch goflag.Name {
		case "logtostderr", "alsologtostderr", "v", "stderrthreshold", "vmodule", "log_backtrace_at", "log_dir":
			goflag.Usage = goflag.Usage + " (glog)"
			pflag.CommandLine.AddGoFlag(goflag)

			if goflag.Name == "v" {
				goflagV = goflag
			} else {
				pflag.CommandLine.MarkHidden(goflag.Name)
			}

			viper.BindPFlag("glog."+goflag.Name, pflag.CommandLine.Lookup(goflag.Name))
		default:
			glog.Warningln("Not handled CLI option:", goflag.Name)
		}
	})

	cobra.OnInitialize(initConfig)

	// see https://godoc.org/github.com/golang/glog#CopyStandardLogTo
	RootCmd.PersistentFlags().String(OPT_COPYSTANDARDLOGTO, "INFO", "Calling CopyStandardLogTo with this option (glog)")
	copystandardlogtoFlag := RootCmd.PersistentFlags().Lookup(OPT_COPYSTANDARDLOGTO)
	copystandardlogtoFlag.Hidden = true
	viper.BindPFlag(OPT_GLOG_COPYSTANDARDLOGTO, copystandardlogtoFlag)

	cobra.OnInitialize()

	goflag.CommandLine.Usage = func() {
		RootCmd.Usage()
	}
	goflag.Parse()
}

func initConfig() {
	viper.AutomaticEnv() // read in environment variables that match

	// Apply if set
	if copyStandardLogTo := viper.GetString("glog.copystandardlogto"); copyStandardLogTo != "" {
		glog.CopyStandardLogTo(copyStandardLogTo)
	}
}
