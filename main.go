package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

func main() {
	url := "https://github.com/konamata/goplay/raw/refs/heads/main/dist/xizm.mp3"

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error downloading MP3:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Failed to download MP3: Status", resp.Status)
		return
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		fmt.Println("Error reading MP3 data:", err)
		return
	}

	streamer, format, err := mp3.Decode(io.NopCloser(&buf))
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

	fmt.Println("Playing audio...")
	<-done
	fmt.Println("Playback finished.")
}
