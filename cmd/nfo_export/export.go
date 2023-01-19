package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/odwrtw/polochon/lib"
	"github.com/odwrtw/polochon/lib/nfo"
	"google.golang.org/api/youtube/v3"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: export <in: meta.txt> <out: oufiles/>")
		os.Exit(-1)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("Error opeining '%s': %s\n", os.Args[1], err)
	}

	defer f.Close()

	out := os.Args[2]

	dec := json.NewDecoder(f)

	var line int64

	var channel *polochon.Show

	for {
		var vid youtube.Video

		if err := dec.Decode(&vid); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			log.Fatalf("Decode failed: %s\n", err)
		}

		if channel == nil {
			channel = &polochon.Show{
				Title: vid.Snippet.ChannelTitle,
				URL:   fmt.Sprintf("youtube.com/channel/%s", vid.Snippet.ChannelId),
			}
		}

		opath := path.Join(out, vid.Id) + ".nfo"
		ep := nfo.NewEpisode(&polochon.ShowEpisode{
			BaseVideo: polochon.BaseVideo{
				File: *polochon.NewFile(path.Join(out, vid.Id+".webm")),
			},

			Title:     vid.Snippet.Title,
			ShowTitle: vid.Snippet.ChannelTitle,
			Season:    0,
			Episode:   int(line),

			Aired:         vid.Snippet.PublishedAt,
			Plot:          vid.Snippet.Description,
			Runtime:       int(parseDur(vid.ContentDetails.Duration).Seconds()),
			EpisodeImdbID: vid.Id,

			Show: channel,
		})

		f, err := os.OpenFile(opath, os.O_CREATE|os.O_RDWR, 0777)
		if err != nil {
			log.Fatalf("Error opening file: %s\n", err)
		}

		defer f.Close()

		enc := xml.NewEncoder(f)

		if err := enc.Encode(&ep); err != nil {
			log.Fatalf("Error encoding XML: %s\n", err)
		}

		if err := enc.Flush(); err != nil {
			log.Fatalf("Error encoding XML: %s\n", err)
		}

		log.Printf("%s: %s \t%d\n", opath, ep.ShowEpisode.Title, line)
		line++
	}
}

func shortTitle(str string) string {
	if len(str) > 20 {
		return str[0:18] + "..."
	}

	return str
}

func parseDur(str string) time.Duration {
	var dur string

	if strings.HasPrefix(str, "PT") {
		dur = strings.ToLower(str[2:])
	} else {
		dur = strings.ToLower(str[1:])
	}

	d, err := time.ParseDuration(dur)
	if err != nil {
		log.Printf("[ERROR] Parseduration failed! %s\n", d)
		return time.Second * 0
	}

	return d
}
