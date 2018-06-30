package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	serviceCmd.PersistentFlags().String(OPT_RECEIVE_ON, DEFAULT_RECEIVE_ON, "Receive text on address:port")
	viper.BindPFlag(OPT_RECEIVE_ON, serviceCmd.PersistentFlags().Lookup(OPT_RECEIVE_ON))

	serviceCmd.PersistentFlags().String(OPT_WRITE_TO, DEFAULT_WRITE_TO, "Send binary to URL")
	viper.BindPFlag(OPT_WRITE_TO, serviceCmd.PersistentFlags().Lookup(OPT_WRITE_TO))
}

func startListening() {

}
