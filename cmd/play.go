package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/spf13/cobra"
)

// playCmd represents the play command
var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Plays an MP3 file",
	Long:  "Plays the specified MP3 file using the Beep library.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Please provide the path to the MP3 file.")
			return
		}

		filePath := args[0]

		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				fmt.Println("Error closing file:", err)
			}
		}(file)

		streamer, format, err := mp3.Decode(file)
		if err != nil {
			fmt.Println("Error decoding MP3:", err)
			return
		}
		defer func(streamer beep.StreamSeekCloser) {
			err := streamer.Close()
			if err != nil {
				fmt.Println("Error closing streamer:", err)
			}
		}(streamer)

		err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
		if err != nil {
			fmt.Println("Error initializing speaker:", err)
			return
		}

		done := make(chan bool)
		speaker.Play(beep.Seq(streamer, beep.Callback(func() {
			done <- true
		})))

		fmt.Printf("Playing %s...\n", filePath)
		<-done
		fmt.Println("Playback finished")
	},
}

func init() {
	rootCmd.AddCommand(playCmd)
}
