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
	re2, err := regexp.Compile(`\[download\] Destination: (.*)`)
	if err != nil {
		log.Fatal(err)
	}
	maybeName := re2.FindStringSubmatch(info)
	if len(maybeName) > 1 {
		return url + " " + maybeName[1]
	}
	result := re1.FindStringSubmatch(info)
	if len(result) > 1 {
		return result[1] + "% " + url
	}
	return "... " + url
}

func download(url string, c chan string) {
	cmd := exec.Command("youtube-dl", "--no-playlist", url)
	// cmd := exec.Command("./test/eclogs")
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	content := make([]byte, 5000)
	for {
		_, err := stdout.Read(content)
		if err != nil {
			break
		}
		info := formatPercentage(string(content), url)
		c <- info
	}

	cmd.Wait()
	c <- "complete " + url
}

func combineLogs(c chan string) {
	ERASE_LINE := "\x1b[2K"
	CURSOR_UP_ONE := "\x1b[1A"
	perc := orderedmap.New()
	names := make(map[string]string)
	for {
		info := <-c
		splits := strings.Split(info, " ")
		if strings.HasPrefix(splits[0], "https://") {
			names[splits[0]] = strings.Join(splits[1:], " ")
			continue
		}
		if splits[0] != "..." {
			perc.Set(splits[1], splits[0])
		}
		var count = 0
		for _, k := range perc.Keys() {
			v, _ := perc.Get(k)

			name, ok := names[k]
			if ok {
				fmt.Printf("%v= %v: %v\n", ERASE_LINE, name, v)
			} else {
				fmt.Printf("%v= %v: %v\n", ERASE_LINE, k, v)
			}
			count += 1
		}
		fmt.Printf(strings.Repeat(CURSOR_UP_ONE, count))
	}
}

func bufferUrls(url string, c chan string, guard chan struct{}) {
	c <- "added " + url
	guard <- struct{}{}
	download(url, c)
	<-guard
}

func main() {
	maxParallel := 10
	prevClip := ""

	c := make(chan string)
	guard := make(chan struct{}, maxParallel)

	go combineLogs(c)
	for {
		var clip, err = clipboard.ReadAll()
		clip = strings.Trim(clip, " ")
		if err != nil {
			log.Fatal(err)
		}
		if prevClip != clip && (strings.HasPrefix(clip, "https://www.youtube.com/") || strings.HasPrefix(clip, "https://youtu.be/") || strings.HasPrefix(clip, "https://youtube.com/")) {
			prevClip = clip
			go bufferUrls(clip, c, guard)
		}
		time.Sleep(333 * time.Millisecond)
	}
}
