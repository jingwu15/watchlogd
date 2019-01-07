package cmd

import (
	//"fmt"
	"github.com/spf13/cobra"
	"github.com/jingwu15/watchlogd/watch"
)

var runCmd = &cobra.Command {
	Use:   "run",
	Short: "run the watchlogd server",
	Long:  `run the watchlogd server`,
	Run: func(cmd *cobra.Command, args []string) {
		mergeViperServer()
		watch.Run()
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start the watchlogd server, use nohup ... &",
	Long:  `start the watchlogd server, use nohup ... &`,
	Run: func(cmd *cobra.Command, args []string) {
		mergeViperServer()
		watch.Start()
	},
}

//var restartCmd = &cobra.Command{
//	Use:   "restart",
//	Short: "restart the watchlogd server",
//	Long:  `restart the watchlogd server `,
//	Run: func(cmd *cobra.Command, args []string) {
//		mergeViperServer()
//		watch.Restart()
//	},
//}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop the watchlogd server",
	Long:  `stop the watchlogd server`,
	Run: func(cmd *cobra.Command, args []string) {
		mergeViperServer()
		watch.Stop()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	//rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(runCmd)
}
