package main

import (
	"html/template"
	"log"
	"net/http"
	"time"
)

// TODO
// - set timeouts on http requests
// - use async

func main() {
	http.HandleFunc("GET /{$}", handler)
	http.HandleFunc("GET /blog", blogHandler)
	http.HandleFunc("GET /blog/{slug}", blogPostHandler)

	//http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
	//http.ServeFile(w, r, "favicon.ico")

	//})

	log.Println("Starting Le Monolithe")
	http.ListenAndServe(":8080", nil)
}

type Data struct {
	Status           map[string]string
	RecentlyWatched  []map[string]string
	CurrentlyReading []map[string]string
	Commits          []map[string]interface{}
}

func formatTime(inputFormat string, t string) (string, error) {
	date, err := time.Parse(inputFormat, t)
	if err != nil {
		return "", err
	}

	return date.Format("Jan 2, 2006"), nil
}

func formatDateTime(inputFormat string, t string) (string, error) {
	date, err := time.Parse(inputFormat, t)
	if err != nil {
		return "", err
	}

	return date.Format("Jan 2, 2006 15:04"), nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	tmpl, err := template.New("").ParseFiles("template/base.html", "template/home.html", "template/commits.html", "template/status.html", "template/watched.html", "template/reading.html")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	commits, err := getCommits()
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	status, err := getStatus()
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	recentlyWatched, err := getRss("https://letterboxd.com/hwebs/rss/", mapLetterboxd)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	currentlyReading, err := getRss("https://www.goodreads.com/review/list_rss/159263337?key=qDjiqflyhso0h4tUk8bW2USB19csqQ3NW32j7SBIIf6FFVG8&shelf=currently-reading", mapGoodreads)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "base", Data{Status: status, RecentlyWatched: recentlyWatched, CurrentlyReading: currentlyReading, Commits: commits})
}
