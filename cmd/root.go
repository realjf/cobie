package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

func init() {
	rootCmd.PersistentFlags().Bool("viper", true, "Use Viper for configuration")
	viper.SetDefault("author", "Real JF <real_jf@hotmail.com>")
	viper.SetDefault("license", "apache")
}

var rootCmd = &cobra.Command{
	Use:   "cobie",
	Short: "cobie is a ping-like command line tools",
}

func Exec() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
