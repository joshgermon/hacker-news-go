package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
)

type Todo struct {
	Id    int
	Label string
}


func getTopStories() []int {
    resp, err := http.Get("https://hacker-news.firebaseio.com/v0/topstories.json")
    if err != nil {
        log.Fatal(err)
    }
    defer resp.Body.Close()

    var topStories []int
    err = json.NewDecoder(resp.Body).Decode(&topStories)
    if err != nil {
        log.Fatal(err)
    }
    return topStories
}

func getStory() {

}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
        getTopStories()
		data := struct { TopStories []int }{ TopStories: getTopStories() }
		tmpl.Execute(w, data)
	})

	http.ListenAndServe(":8080", nil)
}

