package cmd

import (
	"os"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/jingwu15/golib/misc"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "./watchlogd",
	Short: "watch log and write to queue",
	Long:  `watch log and write to queue`,
	//	Run: func(cmd *cobra.Command, args []string) { },
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "watchlogd version",
	Long:  `watchlogd version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("v0.1.0")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)

	rootCmd.PersistentFlags().StringP("config", "c", "./watchlogd.json", "JSON format configuration file")
	viper.BindPFlag("configfile", rootCmd.PersistentFlags().Lookup("config"))

	viper.SetDefault("configfile", "./watchlogd.json")
    mergeViperServer()
}

func mergeViperServer() {
	//加载配置文件
	configfile := viper.Get("configfile").(string)
	if misc.File_exists(configfile) {
		viper.SetConfigFile(configfile)
	} else {
		//设置配置文件查找路径
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/watchlogd")
		//设置配置文件名称，无后缀
		viper.SetConfigName("watchlogd")
	}
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
