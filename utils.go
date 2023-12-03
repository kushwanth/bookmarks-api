package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

var errorBytes []byte

const errorMessage = "Error"
const deleteMessage = "Delete Successful"
const notExistMessage = "URL Redirect for Path doesn't Exists"
const internalError = "Internal Error"
const badRequest = "Bad Request"
const dbError = "DataBase Error"
const duplicateBookmark = "Duplicate Bookmark"
const dbLimit = 1
const pageLimit = 25
const errorStatusCodeLimit = 400

type ResponseMessage struct {
	Message string
}

type Bookmark struct {
	Id        int    `json:"id,omitempty"`
	Title     string `json:"title,omitempty"`
	Link      string `json:"link,omitempty"`
	Timestamp string `json:"lastUpdated,omitempty"`
	Tag       string `json:"tag,omitempty"`
}

type BookmarkData struct {
	Title string `json:"title,omitempty"`
	Link  string `json:"link,omitempty"`
	Tag   string `json:"tag,omitempty"`
}

type SearchQuery struct {
	Data string `json:"data,omitempty"`
}

func validateLinkURL(url string) bool {
	validUrl, err := http.Head(url)
	return err == nil && validUrl.StatusCode < errorStatusCodeLimit
}

func formatText(str string) string {
	formattedStr := strings.ToUpper(str)
	return formattedStr
}

func buildSearchPattern(query string) string {
	queryTokens := strings.Split(query, " ")
	searchPattern := ""
	for i := 0; i < len(queryTokens); i++ {
		searchPattern += formatText(queryTokens[i])
		searchPattern += ":*"
		if i < len(queryTokens)-1 {
			searchPattern += " | "
		}
	}
	return searchPattern
}

func toJson(struc interface{}) []byte {
	responseMessageJson, err := json.Marshal(struc)
	if err != nil {
		return errorBytes
	} else {
		return responseMessageJson
	}
}

func getTitle(link string) (string, error) {
	resp, err := http.Get(link)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	page, err := html.Parse(resp.Body)
	if err != nil {
		return "", err
	}
	return getTitlefromHTML(page), nil
}

func getTitlefromHTML(n *html.Node) string {
	var title string
	if n.Type == html.ElementNode && n.Data == "title" {
		return n.FirstChild.Data
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		title = getTitlefromHTML(c)
		if title != "" {
			break
		}
	}
	return title
}
