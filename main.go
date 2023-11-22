package main

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"html/template"
	"log"
	"net/http"
	"time"
)

type HNSearchResult struct {
	Hits         []HNItem `json:"hits"`
	HitsPerPage  int      `json:"hitsPerPage"`
	NbHits       int      `json:"nbHits"`
	NbPages      int      `json:"nbPages"`
	Page         int      `json:"page"`
	Params       string   `json:"params"`
	Query        string   `json:"query"`
	ServerTimeMS int      `json:"serverTimeMS"`
}

type Author struct {
	MatchLevel   string `json:"matchLevel"`
	MatchedWords []any  `json:"matchedWords"`
	Value        string `json:"value"`
}

type Title struct {
	MatchLevel   string `json:"matchLevel"`
	MatchedWords []any  `json:"matchedWords"`
	Value        string `json:"value"`
}

type URL struct {
	MatchLevel   string `json:"matchLevel"`
	MatchedWords []any  `json:"matchedWords"`
	Value        string `json:"value"`
}

type HighlightResult struct {
	Author Author `json:"author"`
	Title  Title  `json:"title"`
	URL    URL    `json:"url"`
}

type StoryText struct {
	MatchLevel   string `json:"matchLevel"`
	MatchedWords []any  `json:"matchedWords"`
	Value        string `json:"value"`
}

type HNItem struct {
	Tags        []string  `json:"_tags"`
	Author      string    `json:"author"`
	Children    []int     `json:"children"`
	CreatedAt   time.Time `json:"created_at"`
	CreatedAtI  int       `json:"created_at_i"`
	NumComments int       `json:"num_comments"`
	ObjectID    string    `json:"objectID"`
	Points      int       `json:"points"`
	StoryID     int       `json:"story_id"`
	Title       string    `json:"title"`
	UpdatedAt   time.Time `json:"updated_at"`
	URL         string    `json:"url"`
	StoryText   string    `json:"story_text,omitempty"`
}

type HNStory struct {
	Author          string `json:"author"`
	CreatedAt       string `json:"created_at"`
	NumComments     int    `json:"num_comments"`
	Points          int    `json:"points"`
	StoryID         int    `json:"story_id"`
	Title           string `json:"title"`
	LinkDescription string `json:"link_description"`
	URL             string `json:"url"`
}

func getLinkDescription(url string) (string, error) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Println(err)
	}

    log.Printf("getting description for %s", url)

	var description string
	doc.Find("meta").Each(func(index int, item *goquery.Selection) {
		if item.AttrOr("name", "") == "description" {
			description = item.AttrOr("content", "")
		}
	})

	return description, nil
}

func transformStoryData(story HNItem) HNStory {
	layout := "January 2, 2006"
	createdAt := story.CreatedAt.Format(layout)

	linkDescription, err := getLinkDescription(story.URL)
	if err != nil {
		log.Fatal(err)
	}

	return HNStory{
		Author:          story.Author,
		CreatedAt:       createdAt,
		NumComments:     story.NumComments,
		Points:          story.Points,
		StoryID:         story.StoryID,
		Title:           story.Title,
		LinkDescription: linkDescription,
		URL:             story.URL,
	}
}

func getTopStories() []HNStory {
	resp, err := http.Get("https://hn.algolia.com/api/v1/search?tags=front_page")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var result HNSearchResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}

    var topStories []HNStory
    for _, item := range result.Hits {
        topStories = append(topStories, transformStoryData(item))
    }

	return topStories
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		data := getTopStories()
		tmpl.Execute(w, data)
	})

	http.ListenAndServe(":8080", nil)
}
