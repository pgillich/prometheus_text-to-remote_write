package cmd

import (
	"github.com/spf13/cobra"

	"github.com/pgillich/prometheus_text-to-remote_write/util"
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert text to binary, see more info: `prometheus_text-to-remote_write convert -h`",
	Long: `Convert text from stdin to binary to stdout.
Example commands:
`,
	Run: func(cmd *cobra.Command, args []string) {
		startConvert()
	},
}

func init() {
	RootCmd.AddCommand(convertCmd)
}

func startConvert() {
	util.PrintFatalf("NOT IMPLEMENTED\n")
}
