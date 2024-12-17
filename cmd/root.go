package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "mp3player",
	Short: "A simple MP3 player CLI",
	Long:  "A simple CLI application to play MP3 files using Cobra and Beep library.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use the 'play' command to play an MP3 file. For help, use --help")
	},
}

// Execute adds all child commands to the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
