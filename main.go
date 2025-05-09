package main

import (
	"flag"
	"fmt"
	"net/url"
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

func fetchWebsiteWithCache(url string, updateTextChan chan<- string, finally func()) {
	// finally is used to warm up the cache for the adjacent article once first article is fetched
	defer func() {
		if finally != nil {
			go finally()
		}
	}()
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
	lastItemShortcut := rune(list.GetItemCount() - 1 + '1')
	if lastItemShortcut != rune('*') {
		list.AddItem("Settings", "Open settings", rune('*'), nil)
	} else {
		list.AddItem("Settings", "Open settings", rune(list.GetItemCount() + '1'), nil)
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

	go fetchWebsiteWithCache(urls[0], updateTextChan, func() {
		fetchWebsite(urls[1], nil) // warm up below article
	})

	flex := tview.NewFlex().
		AddItem(list, 0, 1, true).
		AddItem(textView, 0, 1, false)

	offset := 0

	inputCapture := defaultInputCapture(app, textView, list, urls, &offset)
	settingsView := buildSettingsMenu(app, flex, list, textView, inputCapture)

	list.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if(index == len(urls)) {
			app.SetInputCapture(buildSettingsInputCapture())
			app.SetRoot(settingsView, true)
			offset = 0
			textView.ScrollTo(offset, 0)
			list.SetCurrentItem(index - 1)
		}
		if index < len(urls) {
			// Reset offset when a new article is selected
			offset = 0
			textView.ScrollTo(offset, 0)

			// Fetch current website
			go fetchWebsiteWithCache(urls[index], updateTextChan, nil)
			textView.SetTitle(titles[index])

			// Pre-fetch next and previous website if they exist
			if index-1 >= 0 {
				go fetchWebsiteWithCache(urls[index-1], nil, nil)
			}
			if index+1 < len(urls) {
				go fetchWebsiteWithCache(urls[index+1], nil, nil)
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

	app.SetInputCapture(inputCapture)
	app.SetRoot(flex, true)

	if err := app.Run(); err != nil {
		panic(err)
	}
}

func defaultInputCapture(app *tview.Application, textView *tview.TextView, list *tview.List, urls []string, offset *int) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			index := list.GetCurrentItem()
			itemText, _ := list.GetItemText(index)
			if index >= 0 && index < len(urls) {
				openURL((urls)[index])
			} else if itemText == "Quit" {
				app.Stop()
			}
			return nil
		case tcell.KeyRight:
			*offset += 1
			textView.ScrollTo(*offset, 0)
			return nil
		case tcell.KeyLeft:
			*offset -= 1
			if *offset < 0 {
				*offset = 0
			}
			textView.ScrollTo(*offset, 0)
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
			case 'k':
				return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
			case 'l':
				*offset += 1
				textView.ScrollTo(*offset, 0)
				return nil
			case 'h':
				*offset -= 1
				if *offset < 0 {
					*offset = 0
				}
				textView.ScrollTo(*offset, 0)
				return nil
			default:
				return event
			}
		default:
			return event
		}
	}
}

func buildSettingsMenu(app *tview.Application, appFlex *tview.Flex, newsList *tview.List, newsText * tview.TextView, defaultInputCapture func(event *tcell.EventKey) *tcell.EventKey) *tview.Flex {
	settingsMenu := tview.NewFlex().SetDirection(tview.FlexColumn)
	settingsMenu.SetBorder(true).SetTitle("Settings").SetTitleAlign(0)

	optionsList := tview.NewList().ShowSecondaryText(false).SetShortcutColor(tcell.ColorDarkGray)
	widgetList := tview.NewFlex().SetDirection(tview.FlexRow)

	// TODO: add more settings
	var colorOptions = []string{tcell.ColorRed.String(), tcell.ColorGreen.String(), tcell.ColorBlue.String(), tcell.ColorYellow.String(), tcell.ColorWhite.String()}
	var settings = map[string]tview.Primitive{
		"Select text color": tview.NewDropDown().
		SetLabel("Select Color (Press enter): ").
		SetLabelColor(tcell.ColorWhite).
		SetOptions(colorOptions, func(option string, index int) {
			newsList.SetMainTextColor(tcell.ColorNames[option])
			newsText.SetTextColor(tcell.ColorNames[option])
			optionsList.SetMainTextColor(tcell.ColorNames[option])
			app.SetFocus(optionsList)
		}).
		SetFieldTextColor(tcell.ColorDarkGray).
		SetFieldBackgroundColor(tcell.ColorWhite),
	}
	
	shortcutIndex := 0
	for option, widget := range settings {
		optionsList.AddItem(option, "", rune('1' + shortcutIndex), func() {
			optionsList.SetCurrentItem(0)
			app.SetFocus(widget)
		})
		widgetList.AddItem(widget, 0, 1, false)
		shortcutIndex++
	}
	optionsList.AddItem("Quit", "", 'q', func() {
		optionsList.SetCurrentItem(0)
		app.SetInputCapture(defaultInputCapture)
		app.SetRoot(appFlex, true)
	})

	settingsMenu.AddItem(optionsList, 0, 1, true)
	settingsMenu.AddItem(widgetList, 0, 1, false)
	return settingsMenu
}

func buildSettingsInputCapture() func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'j':
				return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
			case 'k':
				return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
			default:
				return event
			}
		default:
			return event
		}
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
	parsedUrl, err := url.Parse(feedURL)
	if err != nil || parsedUrl.Scheme == "" {
		feedURL = "https://" + feedURL
	}

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
