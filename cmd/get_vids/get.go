package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	//	"google.golang.org/api/googleapi"
	"github.com/derzombiiie/yt-archive"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: get_vids <channelID>")
		os.Exit(-1)
	}

	s := strings.SplitN(os.Args[1], ":", 2)
	if len(s) != 2 {
		os.Exit(-1)
	}

	var videos []string

	if s[0] == "yt" {
		videos = getVideosDL(s[1])
	} else if s[0] == "file" {
		f, err := os.Open(s[1])
		if err != nil {
			log.Fatalf("Error opening %s: %s\n", s[1], err)
		}
		defer f.Close()

		r := bufio.NewScanner(f)
		r.Scan()

		for r.Scan() {
			videos = append(videos, r.Text())
		}
	}

	fmt.Printf("Got a total of %d videos!\n", len(videos))

	// download all videos metadata:
	// yay!
}

func getVideosDL(query string) []string {

	fmt.Printf("Querying channel '%s'\n", query)

	service, err := archive.NewService()
	if err != nil {
		log.Fatalf("Error creating Service: %s\n", err)
	}

	scall := service.Search.List([]string{"snippet"})
	scall = scall.Q(query)

	sres, err := scall.Do()

	if err != nil {
		log.Fatalf("Error Search: %s\n", err)
	}

	var channelId string

	for _, itm := range sres.Items {
		if itm.Id.Kind == "youtube#channel" {
			fmt.Printf("Using channel:\n")
			fmt.Printf("  ChannelID:    %s\n", itm.Id.ChannelId)
			fmt.Printf("  ChannelTitle: %s\n", itm.Snippet.ChannelTitle)
			fmt.Printf("  Title:        %s\n", itm.Snippet.Title)

			channelId = itm.Snippet.ChannelId

			break
		}
	}

	if channelId == "" {
		log.Fatalf("Couldn't find any channel by Q(%s)\n", query)
	}

	ccall := service.Channels.List([]string{"snippet", "contentDetails"})
	ccall.Id(channelId)

	cres, err := ccall.Do()
	if err != nil {
		log.Fatalf("Error getting channel: %s\n", err)
	}

	if len(cres.Items) == 0 {
		log.Fatalf("len(cres.Items) == 0!\n")
	}

	channel := cres.Items[0]

	fmt.Printf("upload playlist: %v\n", channel.ContentDetails.RelatedPlaylists.Uploads)

	pcall := service.PlaylistItems.List([]string{"snippet", "contentDetails"})
	pcall = pcall.MaxResults(50)
	pcall = pcall.PlaylistId(channel.ContentDetails.RelatedPlaylists.Uploads)
	pres, err := pcall.Do()
	if err != nil {
		log.Fatalf("Error getting uploads-playlist: %s\n", err)
	}

	var videos []string
	f, err := os.OpenFile("video.txt", os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(f, "Videos for %s:\n", channel.Id)

	tmp := make([]string, len(pres.Items))
	for k := range pres.Items {
		tmp[k] = pres.Items[k].ContentDetails.VideoId
		fmt.Fprintln(f, pres.Items[k].ContentDetails.VideoId)
	}

	videos = append(videos, tmp...)

	// get all tokens
	nextToken := pres.NextPageToken
	for nextToken != "" {
		fmt.Printf("NextPageToken: %s\n", nextToken)

		pcall.PageToken(nextToken)
		pres, err := pcall.Do()
		if err != nil {
			log.Fatalf("Error nexttoken: %s: %s\n", nextToken, err)
		}

		nextToken = pres.NextPageToken

		tmp := make([]string, len(pres.Items))
		for k := range pres.Items {
			tmp[k] = pres.Items[k].ContentDetails.VideoId
			fmt.Fprintln(f, pres.Items[k].ContentDetails.VideoId)
		}

		videos = append(videos, tmp...)
	}

	for i := 0; i < len(videos); i++ {
		fmt.Printf("%s, ", videos[i])
	}

	return videos
}
