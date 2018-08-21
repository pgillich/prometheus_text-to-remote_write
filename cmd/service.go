package cmd

import (
	"net/http"

	"github.com/golang/glog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/pgillich/prometheus_text-to-remote_write/conf"
	"github.com/pgillich/prometheus_text-to-remote_write/handler"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Start service, see more info: `prometheus_text-to-remote_write service -h`",
	Long: `Start service.
Example commands:
`,
	Run: func(cmd *cobra.Command, args []string) {
		startListening()
	},
}

func init() {
	RootCmd.AddCommand(serviceCmd)

	serviceCmd.PersistentFlags().String(conf.OPT_RECEIVE_ON, conf.DEFAULT_RECEIVE_ON, "Receive text on address:port")
	viper.BindPFlag(conf.OPT_RECEIVE_ON, serviceCmd.PersistentFlags().Lookup(conf.OPT_RECEIVE_ON))

	serviceCmd.PersistentFlags().String(conf.OPT_RECEIVE_PATH_TEXT, conf.DEFAULT_RECEIVE_PATH_TEXT, "Receive path of text")
	viper.BindPFlag(conf.OPT_RECEIVE_PATH_TEXT, serviceCmd.PersistentFlags().Lookup(conf.OPT_RECEIVE_PATH_TEXT))

	serviceCmd.PersistentFlags().String(conf.OPT_WRITE_TO, conf.DEFAULT_WRITE_TO, "Send binary to URL")
	viper.BindPFlag(conf.OPT_WRITE_TO, serviceCmd.PersistentFlags().Lookup(conf.OPT_WRITE_TO))
}

func startListening() {
	receiveOnAddr := viper.GetString(conf.OPT_RECEIVE_ON)

	http.Handle(viper.GetString(conf.OPT_RECEIVE_PATH_TEXT), http.HandlerFunc(handler.HandlePush))

	glog.Infoln("Receiving on", receiveOnAddr)
	http.ListenAndServe(receiveOnAddr, nil)
}
