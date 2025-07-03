package main

import (
	"bytes"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/feeds"
	_ "github.com/mattn/go-sqlite3"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

type Post struct {
	Title              string
	CreatedAt          string
	FormattedCreatedAt string
	Slug               string
	MetaDescription    string
	Content            template.HTML
}

func getPosts() ([]Post, error) {
	databaseFile := os.Getenv("DATABASE_FILE")

	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var posts []Post
	rows, err := db.Query("SELECT title, slug, meta_description, created_at FROM posts ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.Title, &post.Slug, &post.MetaDescription, &post.CreatedAt); err != nil {
			return nil, err
		}
		formatted, err := formatTime("2006-01-02 15:04:05", post.CreatedAt)
		if err != nil {
			return nil, err
		}
		post.FormattedCreatedAt = formatted
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func getPost(slug string) (Post, error) {
	databaseFile := os.Getenv("DATABASE_FILE")
	var post Post

	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		return post, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT title, slug, meta_description, content, created_at FROM posts WHERE slug = ?", slug)
	var source []byte
	if err := row.Scan(&post.Title, &post.Slug, &post.MetaDescription, &source, &post.CreatedAt); err != nil {
		return post, err
	}
	post.FormattedCreatedAt, err = formatDateTime("2006-01-02 15:04:05", post.CreatedAt)
	if err != nil {
		return post, err
	}

	markdown := goldmark.New(goldmark.WithExtensions(extension.Footnote))

	var buf bytes.Buffer
	if err := markdown.Convert(source, &buf); err != nil {
		return post, err
	}

	post.Content = template.HTML(buf.String())

	return post, nil
}

func newBlogHandler(tmpl *template.Template, _ *Options) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		posts, err := getPosts()
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.ExecuteTemplate(w, "base", posts)
	}
}

func newBlogPostHandler(tmpl *template.Template, _ *Options) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")

		post, err := getPost(slug)
		if err != nil {
			log.Println(err.Error())
			if err == sql.ErrNoRows {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.ExecuteTemplate(w, "base", post)
	}
}

func postsToFeed(posts []Post) ([]*feeds.Item, error) {
	items := make([]*feeds.Item, len(posts))
	for i := range posts {
		created, err := time.Parse("2006-01-02 15:04:05", posts[i].CreatedAt)
		if err != nil {
			return nil, err
		}
		item := &feeds.Item{
			Title:       posts[i].Title,
			Link:        &feeds.Link{Href: "https://henz.world/blog/" + posts[i].Slug},
			Description: posts[i].MetaDescription,
			Created:     created,
		}
		items[i] = item
	}

	return items, nil
}
