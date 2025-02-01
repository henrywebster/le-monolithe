package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type Post struct {
	Title              string
	CreatedAt          string
	FormattedCreatedAt string
	Slug               string
	MetaDescription    string
	Content            string
}

func getPosts() ([]Post, error) {
	db, err := sql.Open("sqlite3", "data.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var posts []Post
	rows, err := db.Query("SELECT title, slug, meta_description, created_at FROM posts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.Title, &post.Slug, &post.MetaDescription, &post.CreatedAt); err != nil {
			return nil, err
		}
		formatted, err := formatDateTime("2006-01-02 15:04:05", post.CreatedAt)
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
	var post Post

	db, err := sql.Open("sqlite3", "data.db")
	if err != nil {
		return post, err
	}
	defer db.Close()

	row := db.QueryRow("SELECT title, slug, meta_description, content, created_at FROM posts WHERE slug = ?", slug)
	if err := row.Scan(&post.Title, &post.Slug, &post.MetaDescription, &post.Content, &post.CreatedAt); err != nil {
		return post, err
	}
	post.FormattedCreatedAt, err = formatDateTime("2006-01-02 15:04:05", post.CreatedAt)
	if err != nil {
		return post, err
	}

	return post, nil
}

func blogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	tmpl, err := template.New("").ParseFiles("template/base.html", "template/blog.html")

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	posts, err := getPosts()
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "base", posts)
}

func blogPostHandler(w http.ResponseWriter, r *http.Request) {
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

	tmpl, err := template.New("").ParseFiles("template/base.html", "template/post.html")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "base", post)
}
