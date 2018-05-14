package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"

	"github.com/henkman/ytpl"
)

func main() {
	var opts struct {
		PrintUrls bool
		Playlist  string
		Out       string
	}
	flag.BoolVar(&opts.PrintUrls, "u", false, "print urls to the videos")
	flag.StringVar(&opts.Playlist, "p", "", "playlist url")
	flag.StringVar(&opts.Out, "o", "", "out file, if ommited use stdout")
	flag.Parse()
	if opts.Playlist == "" {
		flag.Usage()
		return
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	cli := http.Client{Timeout: time.Second * 10, Jar: jar}
	vids, err := ytpl.GetVideos(&cli, opts.Playlist)
	if err != nil {
		fmt.Println(err)
		return
	}
	var out io.Writer
	if opts.Out != "" {
		fd, err := os.OpenFile(opts.Out, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0750)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer fd.Close()
		out = fd
	} else {
		out = os.Stdout
	}
	if opts.PrintUrls {
		for _, vid := range vids {
			fmt.Fprintf(out, "https://www.youtube.com?v=%s\n", vid.ID)
		}
		return
	}
	raw, err := json.MarshalIndent(vids, "", "\t")
	if err != nil {
		fmt.Println(err)
		return
	}
	out.Write(raw)
}
