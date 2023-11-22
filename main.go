package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type HNItem struct {
	By          string `json:"by,omitempty"`
	Descendants int    `json:"descendants,omitempty"`
	ID          int    `json:"id,omitempty"`
	Kids        []int  `json:"kids,omitempty"`
	Score       int    `json:"score,omitempty"`
	Text        string `json:"text,omitempty"`
	Time        int    `json:"time,omitempty"`
	Title       string `json:"title,omitempty"`
	Type        string `json:"type,omitempty"`
}

func getTopStories() []HNItem {
    resp, err := http.Get("https://hacker-news.firebaseio.com/v0/topstories.json")
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    var topStoryIDs []int
    err = json.NewDecoder(resp.Body).Decode(&topStoryIDs)
    if err != nil {
        log.Fatal(err)
    }

    var topStoryItems []HNItem
    for _, id := range topStoryIDs {
        topStoryItems = append(topStoryItems, getItem(id))
    }

    return topStoryItems
}

func getItem(id int) HNItem {
    itemEndpoint := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", id)
    resp, err := http.Get(itemEndpoint)

    log.Printf("Calling endpoint %s", itemEndpoint)
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    var item HNItem

    err = json.NewDecoder(resp.Body).Decode(&item)
    if err != nil {
        log.Fatal(err)
    }

    return item
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		data := getTopStories()
		tmpl.Execute(w, data)
	})

	http.ListenAndServe(":8080", nil)
}

