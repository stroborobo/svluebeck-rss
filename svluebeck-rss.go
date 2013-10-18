package main

import (
	"fmt"
	"os"
	"io"
	/*"os/exec"*/
	"time"
	"strings"
	rss	"github.com/jteeuwen/go-pkg-rss"
)

var (
	items []*rss.Item = make([]*rss.Item, 0, 10)
	lastTimeFile string = "/var/tmp/svluebeck-status"
	newsFile string = "/var/tmp/svluebeck"
)

type News struct {
	time	time.Time
	link	string
	title	string
	//text	string
}

func main() {
	feed := rss.New(10, true, nil, itemHandler)
	url := "http://www.sv-luebeck.de/de/beeintraechtigungen.feed?type=rss"
	if err := feed.Fetch(url, nil); err != nil {
		fmt.Fprintf(os.Stderr, "[e] %s: %s", url, err)
		return
	}

	var last time.Time	// inits with Mon, 01 Jan 0001 00:00:00 +0000
	{
		file, err := os.Open(lastTimeFile)
		if err == nil {
			data := make([]byte, 32)
			count, err := file.Read(data)
			if err == nil {
				tmp, err := time.Parse(time.RFC1123Z, strings.TrimSpace(string(data[:count])))
				if err != nil {
					fmt.Fprintf(os.Stderr, "[Error] parse last:\n%s\n%s\n", err, string(data[:count]))
					os.Exit(2)
				}
				last = tmp
			} else if err != io.EOF {
				fmt.Fprintf(os.Stderr, "[Error] read: %s\n", err)
				os.Exit(2)
			}
		}
		if file != nil {
			file.Close()
		}
	}

	news := make([]*News, 0, len(items))
	for _, item := range items {
		n := new(News)
		{
			tmp, err := time.Parse(time.RFC1123Z, item.PubDate)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[Error] can't parse pubDate: %s\n", err)
				return
			}
			// PubDate <= last
			if !tmp.After(last) {
				continue
			}
			n.time = tmp
		}
		if len(item.Links) >= 1 {
			n.link = item.Links[0].Href
		}
		n.title = item.Title
		// XXX: Why did I add this line?
		//fmt.Println(item.Title);
		// TODO. strip HTML from News.text (item.Description)
		news = append(news, n)
	}

	file, err := os.OpenFile(newsFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprint(os.Stderr, "[Error] open: %s\n", err)
		return
	}
	for _, n := range news {
		file.WriteString(n.time.Format("02.01.06 15:04:05 -> ") + n.title + "\n")
	}
	file.Close()

	for _, n := range news {
		//fmt.Println(n.time.Format("02.01.06 15:04:05 ->"), n.title);
		if n.time.After(last) {
			last = n.time
		}
	}

	file, err = os.OpenFile(lastTimeFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Error] open: %s\n", err)
		return
	}
	file.WriteString(last.Format(time.RFC1123Z))
	file.Close()
}

func itemHandler(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
	l := len(items)
	minlen := l + len(newitems)
	// Do we need more space?
	if minlen > cap(items) {
		slice := make([]*rss.Item, minlen, minlen * 2)
		copy(slice, items)
		items = slice
	}
	items = items[0:minlen]
	for i, item := range newitems {
		items[l+i] = item
	}
}
