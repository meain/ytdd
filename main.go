package main

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/iancoleman/orderedmap"
)

func formatPercentage(info string, url string) string {
	re1, err := regexp.Compile(`\[download\][\s]+([\d]+.[\d]+)%`)
	if err != nil {
		log.Fatal(err)
	}
	result := re1.FindStringSubmatch(info)
	if len(result) > 1 {
		return result[1] + "% " + url
	}
	return "... " + url
}

func download(url string, c chan string) {
	cmd := exec.Command("youtube-dl", "--no-playlist", url)
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	content := make([]byte, 50)
	for {
		_, err := stdout.Read(content)
		if err != nil {
			break
		}
		info := formatPercentage(string(content), url)
		c <- info
	}

	cmd.Wait()
}

func combineLogs(c chan string) {
	perc := orderedmap.New()
	// perc := make(map[string]string)
	for {
		info := <-c
		splits := strings.Split(info, " ")
		if splits[0] != "..." {
			perc.Set(splits[1], splits[0])
			// perc[splits[1]] = splits[0]
			// perc[splits[1]] = "done"
			// delete(perc, splits[1])
		}
		var count = 0
		for _, k := range perc.Keys() {
			v, _ := perc.Get(k)
			// s := strings.Split(k, "=")
			// fmt.Printf("= %v: %v      \n", s[1], v)
			fmt.Printf("= %v: %v      \n", k, v)
			count += 1
		}
		CURSOR_UP_ONE := "\x1b[1A"
		// ERASE_LINE := "\x1b[2K"
		fmt.Printf(strings.Repeat(CURSOR_UP_ONE, count))
	}
}

// https://www.youtube.com/watch?v=OxA0XQ06rrw
// https://www.youtube.com/watch?v=2WXNY1ppTzY
func main() {
	var prevClip = ""
	c := make(chan string)
	go combineLogs(c)
	for {
		var clip, err = clipboard.ReadAll()
		clip = strings.Trim(clip, " ")
		if err != nil {
			log.Fatal(err)
		}
		if prevClip != clip && strings.HasPrefix(clip, "https://") {
			prevClip = clip
			// fmt.Println(clip)
			go download(clip, c)
		}
		time.Sleep(1 * time.Second)
	}
}
