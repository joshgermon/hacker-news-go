package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type HNStorySearchResult struct {
	Hits         []HNStoryHit `json:"hits"`
	HitsPerPage  int          `json:"hitsPerPage"`
	NbHits       int          `json:"nbHits"`
	NbPages      int          `json:"nbPages"`
	Page         int          `json:"page"`
	Params       string       `json:"params"`
	Query        string       `json:"query"`
	ServerTimeMS int          `json:"serverTimeMS"`
}

type StoryText struct {
	MatchLevel   string `json:"matchLevel"`
	MatchedWords []any  `json:"matchedWords"`
	Value        string `json:"value"`
}

type HNStoryHit struct {
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
	Author      string `json:"author"`
	TimeAgo     string `json:"timeAgo"`
	NumComments int    `json:"num_comments"`
	Points      int    `json:"points"`
	StoryID     int    `json:"story_id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
}

func timeAgo(t time.Time) string {
	duration := time.Now().Sub(t)

	switch {
	case duration.Hours() >= 24:
		days := int(duration.Hours() / 24)
		return fmt.Sprintf("%d days ago", days)
	case duration.Minutes() >= 60:
		hours := int(duration.Minutes() / 60)
		return fmt.Sprintf("%d hours ago", hours)
	default:
		minutes := int(duration.Minutes())
		return fmt.Sprintf("%d minutes ago", minutes)
	}
}

func transformStoryData(story HNStoryHit) HNStory {
	timeAgo := timeAgo(story.CreatedAt)

	return HNStory{
		Author:      story.Author,
		TimeAgo:     timeAgo,
		NumComments: story.NumComments,
		Points:      story.Points,
		StoryID:     story.StoryID,
		Title:       story.Title,
		URL:         story.URL,
	}
}

func getTopStories() []HNStory {
	resp, err := http.Get("https://hn.algolia.com/api/v1/search?tags=front_page")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var result HNStorySearchResult
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

// Comment's

type HNCommentSearchResult struct {
	Hits         []HNCommentHit `json:"hits,omitempty"`
	HitsPerPage  int            `json:"hitsPerPage,omitempty"`
	NbHits       int            `json:"nbHits,omitempty"`
	NbPages      int            `json:"nbPages,omitempty"`
	Page         int            `json:"page,omitempty"`
	Params       string         `json:"params,omitempty"`
	Query        string         `json:"query,omitempty"`
	ServerTimeMS int            `json:"serverTimeMS,omitempty"`
}

type HNCommentHit struct {
	Author      string    `json:"author,omitempty"`
	CommentText string    `json:"comment_text,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	ObjectID    string    `json:"objectID,omitempty"`
	ParentID    int       `json:"parent_id,omitempty"`
	Points      any       `json:"points,omitempty"`
	StoryID     int       `json:"story_id,omitempty"`
	StoryTitle  string    `json:"story_title,omitempty"`
	Title       string    `json:"title,omitempty"`
	StoryURL    string    `json:"story_url,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	Children    []int     `json:"children,omitempty"`
	Tags        []string  `json:"_tags"`
}

type Story struct {
	Title    string
	Author   string
	Comments []Comment
}

func getStory(storyID int) Story {
	endpoint := fmt.Sprintf("https://hn.algolia.com/api/v1/search?tags=story_%d&hitsPerPage=100", storyID)
	resp, err := http.Get(endpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var result HNCommentSearchResult
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}

	story := Story{}
	for _, hit := range result.Hits {
		if slices.Contains(hit.Tags, "story") {
			story.Title = hit.Title
			story.Author = hit.Author
		}
		// Is a parent comment
		if hit.ParentID == storyID {
			story.Comments = append(story.Comments, transformComment(hit))
		}
	}

	return story
}

type Comment struct {
	ID       string
	Text     template.HTML
	Author   string
	Children []Comment
}

func transformComment(comment HNCommentHit) Comment {
	return Comment{
		Text:   template.HTML(comment.CommentText),
		Author: comment.Author,
	}
}

type HNItem struct {
  Author    string `json:"author"`
  Children  []HNItem `json:"children"`
  CreatedAt string `json:"created_at"`
  CreatedAtI int `json:"created_at_i"`
  ID        int `json:"id"`
  Options   []string `json:"options"`
  ParentID  int `json:"parent_id"`
  Points    interface{} `json:"points"`
  StoryID   int `json:"story_id"`
  Text      string `json:"text"`
  Title     string `json:"title"`
  Type      string `json:"type"`
  URL       interface{} `json:"url"`
}

func getItem(storyID int) HNItem {
	endpoint := fmt.Sprintf("https://hn.algolia.com/api/v1/items/%d", storyID)
	resp, err := http.Get(endpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var result HNItem
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}
    return result
}

func main() {
	r := chi.NewRouter()

	r.Get("/post/{id}", func(w http.ResponseWriter, r *http.Request) {
		storyID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil {
			log.Fatal(err)
		}
		tmpl := template.Must(template.ParseFiles("templates/post.html"))
		data := getItem(storyID)
		b, err := json.Marshal(data)
		if err != nil {
			fmt.Println("error:", err)
		}
		log.Printf("comments %s", b)
		tmpl.Execute(w, data)
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		data := getTopStories()
		tmpl.Execute(w, data)
	})

	http.ListenAndServe(":8080", r)
}
