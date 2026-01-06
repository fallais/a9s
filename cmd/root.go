package cmd

import (
	"fmt"
	"os"

	"a9s/internal"
	"a9s/pkg/log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "a9s",
	Short: "A k9s-like terminal UI for AWS resources",
	Long:  `a9s is a terminal user interface for browsing and managing AWS resources, inspired by k9s for Kubernetes.`,
	Run:   internal.Run,
}

func init() {
	cobra.OnInitialize(initLogger)

	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug mode")

	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	viper.SetDefault("debug", false)
}

func initLogger() {
	log.InitLogger(viper.GetBool("debug"))
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
