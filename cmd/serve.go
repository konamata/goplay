package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serves a web page to play music from a link",
	Long:  "Starts an HTTP server that serves a web page to play music from a specified link.",
	Run: func(cmd *cobra.Command, args []string) {
		http.HandleFunc("/", servePage)
		http.HandleFunc("/audio", streamAudio)

		go func() {
			if err := http.ListenAndServe(":8080", nil); err != nil {
				fmt.Println("Error starting server:", err)
			}
		}()

		fmt.Println("Server started at http://localhost:8080")

		// Handle graceful shutdown
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		fmt.Println("Shutting down server...")
	},
}

func servePage(w http.ResponseWriter, r *http.Request) {
	html := `
 <!DOCTYPE html>
 <html lang="en">
 <head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Music Player</title>
 </head>
 <body>
  <h1>Music Player</h1>
  <audio controls autoplay>
   <source src="/audio" type="audio/mpeg">
   Your browser does not support the audio element.
  </audio>
 </body>
 </html>
 `
	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(html))
}

func streamAudio(w http.ResponseWriter, r *http.Request) {
	audioURL := "https://github.com/konamata/goplay/raw/refs/heads/main/dist/xizm.mp3"
	resp, err := http.Get(audioURL)
	if err != nil {
		http.Error(w, "Error fetching audio", http.StatusInternalServerError)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}(resp.Body)

	w.Header().Set("Content-Type", "audio/mpeg")
	_, _ = io.Copy(w, resp.Body)
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
