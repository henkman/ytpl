package ytpl_test

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"

	"github.com/henkman/ytpl"
)

func Test(t *testing.T) {
	const pl = `https://www.youtube.com/playlist?list=PLVHGe9JLVncE4La4kosPvlrSF1qcuiOHW`
	jar, err := cookiejar.New(nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	cli := http.Client{Timeout: time.Second * 10, Jar: jar}
	vids, err := ytpl.GetVideos(&cli, pl)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(vids))
	for _, vid := range vids {
		fmt.Println(vid)
	}
}
