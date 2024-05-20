package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "eastloader",
	Short: "eastloader is a small cli tool to download the regional newsletter \"Ostfriesische Nachrichten\"",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("You have to specify a newspaper or api!")
		fmt.Println("Currently there is support for:")
		fmt.Println("\t1. Visiolink ApS, which supports e.g. the Ostfriesischen Nachrichten (visiolink)")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
