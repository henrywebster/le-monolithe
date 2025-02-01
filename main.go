package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// TODO
// - set timeouts on http requests
// - use async

func main() {

	homeHandler, err := newHomeHandler()
	if err != nil {
		panic(err)
	}

	http.HandleFunc("GET /{$}", homeHandler)
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

type Options struct {
	DefaultCacheTTL time.Duration
	LetterboxdURL   string
	GoodreadsURL    string
	GithubToken     string
	GithubQuery     string
	StatusCafeURL   string
}

func readOptions() (Options, error) {
	var options Options
	defaultCacheTTLStr := os.Getenv("CACHE_TTL")
	defaultCacheTTL, err := strconv.ParseInt(defaultCacheTTLStr, 10, 64)
	if err != nil {
		return options, err
	}
	options.DefaultCacheTTL = time.Duration(defaultCacheTTL) * time.Second

	options.LetterboxdURL = os.Getenv("LETTERBOXD_URL")
	options.GoodreadsURL = os.Getenv("GOODREADS_URL")
	options.GithubToken = os.Getenv("GITHUB_TOKEN")
	options.GithubQuery = os.Getenv("GITHUB_GRAPHQL_QUERY")
	options.StatusCafeURL = os.Getenv("STATUS_CAFE_URL")

	return options, nil
}

func newHomeHandler() (http.HandlerFunc, error) {
	tmpl, err := template.New("").ParseFiles("template/base.html", "template/home.html", "template/commits.html", "template/status.html", "template/watched.html", "template/reading.html")
	if err != nil {
		return nil, err
	}

	options, err := readOptions()
	if err != nil {
		return nil, err
	}

	return func(w http.ResponseWriter, r *http.Request) {

		commits, err := getCommits(options.GithubToken, options.GithubQuery)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		status, err := getStatus(options.StatusCafeURL)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		recentlyWatched, err := getRss(options.LetterboxdURL, mapLetterboxd)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		currentlyReading, err := getRss(options.GoodreadsURL, mapGoodreads)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.ExecuteTemplate(w, "base", Data{Status: status, RecentlyWatched: recentlyWatched, CurrentlyReading: currentlyReading, Commits: commits})

	}, nil
}
