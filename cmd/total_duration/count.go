package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"google.golang.org/api/youtube/v3"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: count <in: meta.txt>")
		os.Exit(-1)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("Error opeining '%s': %s\n", os.Args[1], err)
	}

	defer f.Close()

	dec := json.NewDecoder(f)

	var line int64
	var duration time.Duration

	for {
		var vid youtube.Video

		if err := dec.Decode(&vid); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatalf("Decode failed: %s\n", err)
		}

		var dur string
		if strings.HasPrefix(vid.ContentDetails.Duration, "PT") {
			dur = strings.ToLower(vid.ContentDetails.Duration[2:])
		} else {
			dur = strings.ToLower(vid.ContentDetails.Duration[1:])
		}

		dur = strings.ReplaceAll(dur, "d", "s")

		d, err := time.ParseDuration(dur)
		if err != nil {
			log.Fatalf("ParseDuration failed with '%s': %s <- %s\n", vid.ContentDetails.Duration, err, vid.ContentDetails.Duration)
		}

		duration += d

		log.Printf("%6d, %6s <- %s\n", line, dur, vid.ContentDetails.Duration)
		line++
	}

	fmt.Printf("Total Duration: %s (%fm)", duration.String(), duration.Minutes())
}
