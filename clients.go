package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"

	"github.com/mmcdole/gofeed"
)

type ItemMapper = func([]*gofeed.Item) []map[string]string

var cache = NewCache(60 * time.Minute)

func getRss(url string, mapItems ItemMapper, options *Options) ([]map[string]string, error) {

	if data, found := cache.Get(url); found {
		return data.([]map[string]string), nil
	}

	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	log.Println("Fetching", url)

	if err != nil {
		return nil, err
	}

	items := mapItems(feed.Items)

	cache.Set(url, items, options.DefaultCacheTTL)

	return items, nil
}

func mapLetterboxd(items []*gofeed.Item) []map[string]string {
	// TODO - handle errors
	data := make([]map[string]string, len(items))
	for i := 0; i < len(items); i++ {
		formattedDate, _ := formatTime("2006-01-02", items[i].Extensions["letterboxd"]["watchedDate"][0].Value)

		data[i] = make(map[string]string)
		data[i]["title"] = items[i].Extensions["letterboxd"]["filmTitle"][0].Value
		data[i]["filmYear"] = items[i].Extensions["letterboxd"]["filmYear"][0].Value
		data[i]["watchedDate"] = items[i].Extensions["letterboxd"]["watchedDate"][0].Value
		data[i]["formattedWatchedDate"] = formattedDate
		data[i]["link"] = items[i].Link
	}
	return data[:5]
}

func mapGoodreads(items []*gofeed.Item) []map[string]string {
	// TODO - handle errors
	data := make([]map[string]string, len(items))
	for i := 0; i < len(items); i++ {
		data[i] = make(map[string]string)
		data[i]["title"] = items[i].Title
		data[i]["link"] = items[i].Link
		data[i]["authorName"] = items[i].Custom["author_name"]
	}
	return data
}

func getStatus(url string, options *Options) (map[string]string, error) {

	if data, found := cache.Get(url); found {
		return data.(map[string]string), nil
	}

	response, err := http.Get(url)
	log.Println("Fetching", url)

	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var status map[string]string
	err = json.NewDecoder(response.Body).Decode(&status)
	if err != nil {
		return nil, err
	}

	cache.Set(url, status, options.DefaultCacheTTL)

	return status, nil
}

func getCommits(options *Options) ([]map[string]interface{}, error) {
	if data, found := cache.Get("commits"); found {
		return data.([]map[string]interface{}), nil
	}

	jsonData := map[string]string{
		"query": options.GithubQuery,
	}

	reqBody, err := json.Marshal(jsonData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", "Bearer "+options.GithubToken)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(req)
	defer response.Body.Close()

	log.Println("Fetching commits")
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	result, err := io.ReadAll(response.Body)
	var body map[string]interface{}
	if err := json.Unmarshal(result, &body); err != nil {
		return nil, err
	}

	//log.Println(body)

	data := body["data"].(map[string]interface{})
	viewer := data["viewer"].(map[string]interface{})
	repo := viewer["repository"].(map[string]interface{})
	defaultBranch := repo["defaultBranchRef"].(map[string]interface{})
	target := defaultBranch["target"].(map[string]interface{})
	history := target["history"].(map[string]interface{})
	nodes := history["nodes"].([]interface{})

	commits := make([]map[string]interface{}, len(nodes))
	for i, node := range nodes {
		commits[i] = node.(map[string]interface{})
	}

	for _, commit := range commits {
		formattedTime, err := formatDateTime(time.RFC3339, commit["committedDate"].(string))
		if err != nil {
			return nil, err
		}

		commit["repositoryName"] = repo["name"]
		commit["repositoryUrl"] = repo["url"]
		commit["formattedCommittedDate"] = formattedTime
	}

	cache.Set("commits", commits, options.DefaultCacheTTL)

	return commits, nil
}

func getTopArtists(options *Options) ([]map[string]interface{}, error) {
	if data, found := cache.Get("top-artists"); found {
		return data.([]map[string]interface{}), nil
	}

	log.Println("Fetching top artists")
	auth := spotifyauth.New(
		spotifyauth.WithClientID(os.Getenv("SPOTIFY_CLIENT_ID")),
		spotifyauth.WithClientSecret(os.Getenv("SPOTIFY_CLIENT_SECRET")),
	)

	token, err := auth.RefreshToken(context.Background(), &oauth2.Token{RefreshToken: os.Getenv("SPOTIFY_REFRESH_TOKEN")})
	if err != nil {
		log.Fatal(err)
	}

	client := auth.Client(context.Background(), token)
	resp, err := client.Get("https://api.spotify.com/v1/me/top/artists?limit=5")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	result, err := io.ReadAll(resp.Body)
	var body map[string]interface{}
	if err := json.Unmarshal(result, &body); err != nil {
		return nil, err
	}

	items := body["items"].([]interface{})

	artists := make([]map[string]interface{}, len(items))
	for i, item := range items {
		artists[i] = item.(map[string]interface{})
		genre := artists[i]["genres"].([]interface{})
		if len(genre) > 0 {
			artists[i]["genre"] = genre[0]
		} else {
			artists[i]["genre"] = ""
		}
	}

	cache.Set("top-artists", artists, options.DefaultCacheTTL)
	return artists, nil
}
