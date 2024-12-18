package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const defaultAudioURL = "https://github.com/konamata/goplay/raw/refs/heads/main/dist/xizm.mp3"
const tempFileName = "default.mp3"

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Plays an audio file",
	Long:  "Plays the specified audio file.",
	Run: func(cmd *cobra.Command, args []string) {
		var filePath string

		// Eğer dosya yolu verilmemişse, varsayılan URL'den dosyayı indir.
		if len(args) < 1 {
			fmt.Println("No audio file provided. Playing the default audio...")
			var err error
			filePath, err = downloadDefaultAudio()
			if err != nil {
				fmt.Println("Error downloading default audio:", err)
				return
			}
		} else {
			filePath = args[0]
		}

		ext := strings.ToLower(filepath.Ext(filePath))
		if ext != ".mp3" && ext != ".wav" && ext != ".flac" && ext != ".ogg" {
			fmt.Println("Unsupported file type. Please provide an MP3, WAV, FLAC, or OGG file.")
			return
		}

		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer func(file *os.File) {
			_ = file.Close()
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

		volume := &effects.Volume{
			Streamer: beep.Loop(-1, streamer),
			Base:     2,
			Volume:   0.0, // Start at 100% volume
			Silent:   false,
		}

		ctrl := &beep.Ctrl{Streamer: volume, Paused: false}
		speaker.Play(ctrl)

		done := make(chan bool)
		go func() {
			for {
				select {
				case <-done:
					return
				default:
					speaker.Lock()
					position := streamer.Position()
					length := streamer.Len()
					speaker.Unlock()
					printProgressBar(position, length, volume.Volume)
					time.Sleep(time.Second)
				}
			}
		}()

		handleKeyPress(streamer, volume, done)
	},
}

func downloadDefaultAudio() (string, error) {
	resp, err := http.Get(defaultAudioURL)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}(resp.Body)

	out, err := os.Create(tempFileName)
	if err != nil {
		return "", err
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			fmt.Println("Error closing output file:", err)
		}
	}(out)

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return tempFileName, nil
}

func printProgressBar(position, length int, volume float64) {
	progress := float64(position) / float64(length)
	barWidth := 50
	progressWidth := int(progress * float64(barWidth))

	currentTime := formatTime(position)
	totalTime := formatTime(length)
	volumePercentage := int((volume + 1) * 100) // Adjust volume percentage calculation

	fmt.Printf("\r[")
	for i := 0; i < barWidth; i++ {
		if i < progressWidth {
			fmt.Print("=")
		} else {
			fmt.Print(" ")
		}
	}
	fmt.Printf("] %d%% %s/%s Volume: %d%%", int(progress*100), currentTime, totalTime, volumePercentage)
}

func formatTime(frames int) string {
	seconds := frames / 44100
	minutes := seconds / 60
	seconds = seconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func handleKeyPress(streamer beep.StreamSeekCloser, volume *effects.Volume, done chan bool) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("Error setting terminal to raw mode:", err)
		return
	}
	defer func(fd int, oldState *term.State) {
		err := term.Restore(fd, oldState)
		if err != nil {
			fmt.Println("Error restoring terminal:", err)
		}
	}(int(os.Stdin.Fd()), oldState)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		done <- true
	}()

	for {
		var buf [1]byte
		_, err := os.Stdin.Read(buf[:])
		if err != nil {
			fmt.Println("Error reading from stdin:", err)
			return
		}
		switch buf[0] {
		case 27: // Arrow keys
			_, err := os.Stdin.Read(buf[:])
			if err != nil {
				fmt.Println("Error reading from stdin:", err)
				return
			}
			if buf[0] == 91 {
				_, err := os.Stdin.Read(buf[:])
				if err != nil {
					fmt.Println("Error reading from stdin:", err)
					return
				}
				switch buf[0] {
				case 67: // Right arrow
					speaker.Lock()
					err := streamer.Seek(streamer.Position() + streamer.Len()/10)
					if err != nil {
						fmt.Println("Error seeking streamer:", err)
						return
					}
					speaker.Unlock()
				case 68: // Left arrow
					speaker.Lock()
					err := streamer.Seek(streamer.Position() - streamer.Len()/10)
					if err != nil {
						fmt.Println("Error seeking streamer:", err)
						return
					}
					speaker.Unlock()
				case 65: // Up arrow
					speaker.Lock()
					volume.Volume += 0.1
					speaker.Unlock()
				case 66: // Down arrow
					speaker.Lock()
					volume.Volume -= 0.1
					speaker.Unlock()
				}
			}
		case 'q': // Quit
			done <- true
			return
		}
	}
}

func init() {
	rootCmd.AddCommand(playCmd)
}
