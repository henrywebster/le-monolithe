package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// TODO
// - set timeouts on http requests
// - use async
// - favicon
// - 404 page

type HandlerCreator = func(*Options) (func(http.ResponseWriter, *http.Request), error)

func main() {

	options, err := readOptions()
	if err != nil {
		panic(err)
	}

	// TODO: better way to handle templates
	homeFiles := []string{"base.html", "home.html", "commits.html", "status.html", "watched.html", "reading.html", "top-artists.html"}
	var paths []string
	for _, file := range homeFiles {
		paths = append(paths, filepath.Join(options.TemplateDir, file))
	}
	log.Println(paths)
	tmplHome, err := template.New("").ParseFiles(paths...)
	if err != nil {
		log.Println(err)
	}
	homeHandler := newHomeHandler(tmplHome, &options)

	blogFiles := []string{"base.html", "blog.html"}
	var blogPaths []string
	for _, file := range blogFiles {
		blogPaths = append(blogPaths, filepath.Join(options.TemplateDir, file))
	}
	tmplBlog, err := template.New("").ParseFiles(blogPaths...)
	if err != nil {
		log.Println(err)
	}
	blogHandler := newBlogHandler(tmplBlog, &options)

	blogPostFiles := []string{"base.html", "post.html"}
	var blogPostPaths []string
	for _, file := range blogPostFiles {
		blogPostPaths = append(blogPostPaths, filepath.Join(options.TemplateDir, file))
	}
	tmplBlogPost, err := template.New("").ParseFiles(blogPostPaths...)
	if err != nil {
		log.Println(err)
	}
	blogPostHandler := newBlogPostHandler(tmplBlogPost, &options)

	musicFiles := []string{"base.html", "music.html"}
	var musicPaths []string
	for _, file := range musicFiles {
		musicPaths = append(musicPaths, filepath.Join(options.TemplateDir, file))
	}
	tmplMusic, err := template.New("").ParseFiles(musicPaths...)
	if err != nil {
		log.Println(err)
	}
	musicHandler := newMusicHandler(tmplMusic, &options)

	staticHandler := newStaticHandler(&options)

	http.HandleFunc("GET /{$}", homeHandler)
	http.HandleFunc("GET /blog", blogHandler)
	http.HandleFunc("GET /blog/{slug}", blogPostHandler)
	http.HandleFunc("GET /music", musicHandler)
	http.HandleFunc("GET /image/{file}", staticHandler)
	http.HandleFunc("GET /style/{file}", staticHandler)

	log.Printf("Starting Le Monolithe on :%d\n", options.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", options.Port), nil)
}

type Data struct {
	Status           map[string]string
	RecentlyWatched  []map[string]string
	CurrentlyReading []map[string]string
	Commits          []map[string]interface{}
	TopArtists       []map[string]interface{}
}

func formatTime(inputFormat string, t string) (string, error) {
	date, err := time.Parse(inputFormat, t)
	if err != nil {
		return "", err
	}

	return date.Format("Jan 02, 2006"), nil
}

func formatDateTime(inputFormat string, t string) (string, error) {
	date, err := time.Parse(inputFormat, t)
	if err != nil {
		return "", err
	}

	return date.Format("Jan 02, 2006 15:04"), nil
}

type Options struct {
	DefaultCacheTTL time.Duration
	LetterboxdURL   string
	GoodreadsURL    string
	GithubToken     string
	GithubQuery     string
	StatusCafeURL   string
	TemplateDir     string
	Port            int
	StaticDir       string
}

func readOptions() (Options, error) {
	var options Options
	defaultCacheTTLStr := os.Getenv("CACHE_TTL")
	defaultCacheTTL, err := strconv.ParseInt(defaultCacheTTLStr, 10, 64)
	if err != nil {
		return options, err
	}
	options.DefaultCacheTTL = time.Duration(defaultCacheTTL) * time.Second

	portStr := os.Getenv("PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return options, err
	}
	options.Port = port

	options.LetterboxdURL = os.Getenv("LETTERBOXD_URL")
	options.GoodreadsURL = os.Getenv("GOODREADS_URL")
	options.GithubToken = os.Getenv("GITHUB_TOKEN")
	options.GithubQuery = os.Getenv("GITHUB_GRAPHQL_QUERY")
	options.StatusCafeURL = os.Getenv("STATUS_CAFE_URL")
	options.TemplateDir = os.Getenv("TEMPLATE_DIR")
	options.StaticDir = os.Getenv("STATIC_DIR")

	return options, nil
}

func newHomeHandler(tmpl *template.Template, options *Options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		commits, err := getCommits(options)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		status, err := getStatus(options.StatusCafeURL, options)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		recentlyWatched, err := getRss(options.LetterboxdURL, mapLetterboxd, options)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		currentlyReading, err := getRss(options.GoodreadsURL, mapGoodreads, options)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		topArtists, err := getTopArtists(options)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.ExecuteTemplate(w, "base", Data{Status: status, RecentlyWatched: recentlyWatched, CurrentlyReading: currentlyReading, Commits: commits, TopArtists: topArtists})

	}
}
