package ytpl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// thumbnail is https://i.ytimg.com/vi/$ID/hqdefault.jpg
// url is https://www.youtube.com/watch?v=$ID
type Video struct {
	Title    string        `json:"title"`
	ID       string        `json:"id"`
	Duration time.Duration `json:"duration"`
}

func (v Video) String() string {
	return fmt.Sprintf("%s, %s, %v", v.Title, v.ID, v.Duration)
}

func parseVideo(s *goquery.Selection) Video {
	var vid Video
	if a, ok := s.Attr("data-video-id"); ok {
		vid.ID = a
	}
	if a, ok := s.Attr("data-title"); ok {
		vid.Title = a
	}
	{
		d := time.Duration(0)
		ps := strings.Split(s.Find(".timestamp span").Text(), ":")
		units := []time.Duration{time.Second, time.Minute, time.Hour}
		u := 0
		for i := len(ps) - 1; i >= 0; i-- {
			t, _ := strconv.ParseInt(ps[i], 10, 32)
			d += units[u] * time.Duration(t)
			u++
		}
		vid.Duration = d
	}
	return vid
}

func GetVideos(cli *http.Client, url string) ([]Video, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	res, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return nil, err
	}
	first := doc.Find("tr.pl-video")
	videos := make([]Video, 0, first.Size())
	first.Each(func(i int, s *goquery.Selection) {
		videos = append(videos, parseVideo(s))
	})
	loadmore := doc.Find("button[data-uix-load-more-href]")
	for loadmore.Size() > 0 {
		next := "https://youtube.com" +
			loadmore.AttrOr("data-uix-load-more-href", "")
		req, err := http.NewRequest("GET", next, nil)
		if err != nil {
			return nil, err
		}
		res, err := cli.Do(req)
		if err != nil {
			return nil, err
		}
		var r struct {
			ContentHtml        string `json:"content_html"`
			LoadMoreWidgetHtml string `json:"load_more_widget_html"`
		}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			res.Body.Close()
			return nil, err
		}
		res.Body.Close()
		doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(
			"<!DOCTYPE html><html><body><table>" +
				r.ContentHtml +
				"</table></body></html>"))
		if err != nil {
			return nil, err
		}
		doc.Find("tr.pl-video").Each(func(i int, s *goquery.Selection) {
			videos = append(videos, parseVideo(s))
		})
		if r.LoadMoreWidgetHtml == "" {
			break
		}
		doc, err = goquery.NewDocumentFromReader(bytes.NewBufferString(
			"<!DOCTYPE html><html><body>" +
				r.LoadMoreWidgetHtml +
				"</body></html>"))
		if err != nil {
			return nil, err
		}
		loadmore = doc.Find("button[data-uix-load-more-href]")
	}
	return videos, nil
}
