package main

import (
	"flag"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/go-shiori/go-readability"
	"github.com/mmcdole/gofeed"
	"github.com/rivo/tview"
	"jaytaylor.com/html2text"
)

var articleCache = make(map[string]string)

func fetchWebsiteWithCache(url string, updateTextChan chan<- string) {
	if text, ok := articleCache[url]; ok {
		updateTextChan <- text
	} else {
		fetchWebsite(url, updateTextChan)
	}
}
func main() {
	app := tview.NewApplication()

	rssUrl := flag.String("rss", "https://feeds.bbci.co.uk/news/rss.xml", "Specify the rss url")
	flag.Parse()

	urls, titles := fetchRSSArticles(*rssUrl)

	list := tview.NewList().ShowSecondaryText(false).SetShortcutColor(tcell.ColorDarkGray)
	for i, title := range titles {
		list.AddItem(title, "", rune('1'+i), nil)
	}
	list.AddItem("Quit", "Press to exit", 'q', func() {
		app.Stop()
	})

	updateTextChan := make(chan string)

	list.SetBorder(true)

	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	textView.SetBorder(true).SetTitle(titles[0]).SetTitleAlign(0)
	go fetchWebsite(urls[0], updateTextChan)

	flex := tview.NewFlex().
		AddItem(list, 0, 1, true).
		AddItem(textView, 0, 1, false)

	offset := 0

	list.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if index < len(urls) {
			// Reset offset when a new article is selected
			offset = 0
			textView.ScrollTo(offset, 0)

			// Fetch current website
			go fetchWebsiteWithCache(urls[index], updateTextChan)
			textView.SetTitle(titles[index])

			// Pre-fetch next and previous website if they exist
			if index-1 >= 0 {
				go fetchWebsite(urls[index-1], nil)
			}
			if index+1 < len(urls) {
				go fetchWebsite(urls[index+1], nil)
			}
		}
	})

	go func() {
		for {
			select {
			case response := <-updateTextChan:
				app.QueueUpdateDraw(func() {
					textView.SetText(response).SetTitleAlign(tview.AlignLeft)
				})
			}
		}
	}()

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			openURL(urls[list.GetCurrentItem()])
			return nil
		case tcell.KeyRight:
			offset += 1
			textView.ScrollTo(offset, 0)
			return nil
		case tcell.KeyLeft:
			offset -= 1
			if offset < 0 {
				offset = 0
			}
			textView.ScrollTo(offset, 0)
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
			case 'k':
				return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
			case 'l':
				offset += 1
				textView.ScrollTo(offset, 0)
				return nil
			case 'h':
				offset -= 1
				if offset < 0 {
					offset = 0
				}
				textView.ScrollTo(offset, 0)
				return nil
			default:
				return event
			}
		default:
			return event
		}
	})

	app.SetRoot(flex, true)

	if err := app.Run(); err != nil {
		panic(err)
	}
}

// TODO: revisit this in case of unneccessary update to TextChan and make sure text in textview corresponds to selected article
func fetchWebsite(url string, updateTextChan chan<- string) {
	sanitized, _ := readability.FromURL(url, 30*time.Second)
	what, _ := html2text.FromString(sanitized.Content)
	articleCache[url] = string(what)
	updateTextChan <- string(what)
}

func fetchRSSArticles(feedURL string) (urls []string, titles []string) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		fmt.Println("Error fetching or parsing feed:", err)
		return
	}

	for _, item := range feed.Items {
		urls = append(urls, item.Link)
		titles = append(titles, item.Title)
	}
	return
}

func openURL(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	exec.Command(cmd, args...).Start()
}
