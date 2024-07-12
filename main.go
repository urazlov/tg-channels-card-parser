package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sqweek/dialog"
	"golang.org/x/net/html"
)

type Channel struct {
	Title       string `json:"title"`
	Subscribers string `json:"subscribers"`
	Views       string `json:"views"`
	Rating      string `json:"rating"`
	ER          string `json:"er"`
	FullPrice   string `json:"full_price"`
}

func ParseChannels(r io.Reader) ([]Channel, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	nodes := channelNodes(doc)

	var channels []Channel

	for _, node := range nodes {
		channels = append(channels, buildChannel(node))
	}

	return channels, nil
}

func buildChannel(n *html.Node) Channel {
	var channel Channel

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, attr := range n.Attr {
				if attr.Key == "class" {
					if strings.Contains(attr.Val, "channel_title") {
						channel.Title = getText(n)
					} else if strings.Contains(attr.Val, "channel-users-count") {
						channel.Subscribers = getText(n)
					} else if strings.Contains(attr.Val, "arating") {
						channel.Rating = getText(n)
					} else if strings.Contains(attr.Val, "item") && strings.Contains(attr.Val, "_3") {
						channel.ER = getText(findChildWithClass(n, "js-err"))
					} else if strings.Contains(attr.Val, "current_price") {
						channel.FullPrice = getText(n)
					} else if strings.Contains(attr.Val, "item") && strings.Contains(attr.Val, "_2") {
						channel.Views = getText(findChildWithClass(n, "js-view"))
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return channel
}

func findChildWithClass(n *html.Node, class string) *html.Node {
	var f func(*html.Node) *html.Node
	f = func(n *html.Node) *html.Node {
		if n.Type == html.ElementNode {
			for _, attr := range n.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, class) {
					return n
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if result := f(c); result != nil {
				return result
			}
		}
		return nil
	}
	return f(n)
}

func getText(n *html.Node) string {
	if n == nil {
		return ""
	}
	if n.Type == html.TextNode {
		return n.Data
	}
	var ret string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		ret += getText(c) + " "
	}
	return strings.TrimSpace(ret)
}

func channelNodes(n *html.Node) []*html.Node {
	if n.Type == html.ElementNode && n.Data == "div" {
		for _, attr := range n.Attr {
			if attr.Key == "class" && strings.Contains(attr.Val, "channels-item") {
				return []*html.Node{n}
			}
		}
	}
	var ret []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		ret = append(ret, channelNodes(c)...)
	}
	return ret
}

func saveToJSON(channels []Channel, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(channels)
}

func main() {
	filePath, err := dialog.File().Filter("HTML files", "html").Load()
	if err != nil {
		fmt.Println("Error selecting file:", err)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	channels, err := ParseChannels(file)
	if err != nil {
		fmt.Println("Error parsing channels:", err)
		return
	}

	err = saveToJSON(channels, "channels.json")
	if err != nil {
		fmt.Println("Error saving JSON:", err)
		return
	}

	fmt.Println("Все должно быть спаршено ебано рот")
}
