package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type Link struct {
	Platform string
	Url      string
}

type Album struct {
	Links       []Link
	Title       string
	ReleaseDate string
	Id          int
}

type Artist struct {
	Albums []Album
	Name   string
	Id     int
}

func getLinks(album_id int) ([]Link, error) {
	databaseFile := os.Getenv("DATABASE_FILE")

	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var links []Link
	rows, err := db.Query("SELECT platform, url FROM links WHERE album_id = ?", album_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var link Link
		if err := rows.Scan(&link.Platform, &link.Url); err != nil {
			return nil, err
		}

		//	formatted, err := formatDateTime("2006-01-02 15:04:05", post.CreatedAt)
		//	if err != nil {
		//		return nil, err
		//	}
		//	post.FormattedCreatedAt = formatted
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}

func getAlbums(artist_id int) ([]Album, error) {
	databaseFile := os.Getenv("DATABASE_FILE")

	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var albums []Album
	rows, err := db.Query("SELECT id, title, release_date FROM albums WHERE artist_id = ?", artist_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var album Album
		if err := rows.Scan(&album.Id, &album.Title, &album.ReleaseDate); err != nil {
			return nil, err
		}

		//	formatted, err := formatDateTime("2006-01-02 15:04:05", post.CreatedAt)
		//	if err != nil {
		//		return nil, err
		//	}
		//	post.FormattedCreatedAt = formatted
		links, err := getLinks(album.Id)
		if err != nil {
			return nil, err
		}
		album.Links = links

		albums = append(albums, album)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return albums, nil
}

func getArtists() ([]Artist, error) {
	databaseFile := os.Getenv("DATABASE_FILE")

	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var artists []Artist
	rows, err := db.Query("SELECT id, name FROM artists")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var artist Artist
		if err := rows.Scan(&artist.Id, &artist.Name); err != nil {
			return nil, err
		}

		albums, err := getAlbums(artist.Id)
		if err != nil {
			return nil, err
		}
		artist.Albums = albums
		//	formatted, err := formatDateTime("2006-01-02 15:04:05", post.CreatedAt)
		//	if err != nil {
		//		return nil, err
		//	}
		//	post.FormattedCreatedAt = formatted
		artists = append(artists, artist)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return artists, nil
}

func newMusicHandler(tmpl *template.Template, _ *Options) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		artists, err := getArtists()
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.ExecuteTemplate(w, "base", artists)
	}
}
