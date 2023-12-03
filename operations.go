package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func checkDuplicateBookmark(link string, db *pgxpool.Pool) bool {
	duplicateId := 0
	db_err := db.QueryRow(context.Background(), "SELECT id FROM bookmarks WHERE link=$1 LIMIT $2", link, dbLimit).Scan(&duplicateId)
	return db_err != nil && duplicateId == 0
}

func queryDbWithId(id int, db *pgxpool.Pool) (Bookmark, error) {
	var responseData Bookmark
	db_err := db.QueryRow(context.Background(), "SELECT id, title, link, timestamp::TEXT, tag FROM bookmarks WHERE id=$1 LIMIT $2", id, dbLimit).Scan(&responseData.Id, &responseData.Title, &responseData.Link, &responseData.Timestamp, &responseData.Tag)
	if db_err != nil {
		return responseData, errors.New(db_err.Error())
	}
	return responseData, nil
}

func getBookmark(db *pgxpool.Pool) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var response ResponseMessage
		bookmarkId, idErr := strconv.Atoi(chi.URLParam(r, "id"))
		if idErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = badRequest
			w.Write(toJson(response))
			log.Print(idErr.Error(), "Invalid ID")
			return
		}
		bookmark, dbErr := queryDbWithId(bookmarkId, db)
		if dbErr != nil || bookmark.Id != bookmarkId {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(dbError))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(bookmark))
	}
	return http.HandlerFunc(fn)
}

func createBookmark(db *pgxpool.Pool) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var requestData BookmarkData
		var response ResponseMessage
		err := json.NewDecoder(r.Body).Decode(&requestData)
		isLinkValid := validateLinkURL(requestData.Link)
		w.Header().Set("Content-Type", "application/json")
		if !isLinkValid || err != nil || len(requestData.Link) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = badRequest
			w.Write(toJson(response))
			log.Print(err.Error(), "Invalid Link")
			return
		}
		if !checkDuplicateBookmark(requestData.Link, db) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(duplicateBookmark))
			return
		}
		if len(requestData.Title) == 0 {
			title, err := getTitle(requestData.Link)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				response.Message = badRequest
				w.Write(toJson(response))
				log.Print(err.Error(), "Invalid Title")
				return
			}
			requestData.Title = title
		}
		var responseData Bookmark
		db_err := db.QueryRow(context.Background(), "INSERT INTO bookmarks (title, link, timestamp, tag) VALUES ($1,$2,now(),$3) RETURNING id, title, link, timestamp::TEXT, tag", formatText(requestData.Title), requestData.Link, formatText(requestData.Tag)).Scan(&responseData.Id, &responseData.Title, &responseData.Link, &responseData.Timestamp, &responseData.Tag)
		if db_err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(dbError))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(responseData))
	}
	return http.HandlerFunc(fn)
}

func updateBookmark(db *pgxpool.Pool) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var requestData BookmarkData
		var response ResponseMessage
		bookmarkId, idErr := strconv.Atoi(chi.URLParam(r, "id"))
		err := json.NewDecoder(r.Body).Decode(&requestData)
		isLinkValid := validateLinkURL(requestData.Link)
		w.Header().Set("Content-Type", "application/json")
		if !isLinkValid || err != nil || idErr != nil || len(requestData.Link) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = badRequest
			w.Write(toJson(response))
			log.Print(err.Error(), "Invalid Link")
			return
		}
		if !checkDuplicateBookmark(requestData.Link, db) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(duplicateBookmark))
			return
		}
		if len(requestData.Title) == 0 {
			title, err := getTitle(requestData.Link)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				response.Message = badRequest
				w.Write(toJson(response))
				log.Print(err.Error(), "Invalid Title")
				return
			}
			requestData.Title = title
		}
		bookmark, dbErr := queryDbWithId(bookmarkId, db)
		if dbErr != nil || bookmark.Id != bookmarkId {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = notExistMessage
			w.Write(toJson(response))
			return
		}
		var responseData Bookmark
		db_err := db.QueryRow(context.Background(), "UPDATE bookmarks SET title=$1, link=$2, timestamp=now(), tag=$3 WHERE id=$4 RETURNING id, title, link, timestamp::TEXT, tag", formatText(requestData.Title), requestData.Link, formatText(requestData.Tag), bookmarkId).Scan(&responseData.Id, &responseData.Title, &responseData.Link, &responseData.Timestamp, &responseData.Tag)
		if db_err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(dbError))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(responseData))
	}
	return http.HandlerFunc(fn)
}

func deleteBookmark(db *pgxpool.Pool) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var response ResponseMessage
		bookmarkId, idErr := strconv.Atoi(chi.URLParam(r, "id"))
		if idErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = badRequest
			w.Write(toJson(response))
			log.Print(idErr.Error(), "Invalid ID")
			return
		}
		bookmark, dbErr := queryDbWithId(bookmarkId, db)
		if dbErr != nil || bookmark.Id != bookmarkId {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = notExistMessage
			w.Write(toJson(response))
			return
		}
		deleteErr := db.QueryRow(context.Background(), "DELETE FROM bookmarks WHERE id=$1 RETURNING id", bookmark.Id).Scan(&bookmarkId)
		if deleteErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response.Message = internalError
			w.Write(toJson(response))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(deleteMessage))
	}
	return http.HandlerFunc(fn)
}

func listBookmarks(db *pgxpool.Pool) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var responseData []Bookmark
		var response ResponseMessage
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		min, max := page, page+pageLimit
		rows, db_err := db.Query(context.Background(), "SELECT id, title, link, timestamp::TEXT, tag FROM bookmarks where id>$1 AND id<=$2 LIMIT $3", min, max, pageLimit)
		if db_err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response.Message = dbError
			w.Write(toJson(response))
			log.Print(db_err.Error())
			return
		}
		for rows.Next() {
			var temp Bookmark
			err := rows.Scan(&temp.Id, &temp.Title, &temp.Link, &temp.Timestamp, &temp.Tag)
			if err != nil {
				log.Print(temp.Id, err.Error())
			}
			responseData = append(responseData, temp)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(responseData))
	}
	return http.HandlerFunc(fn)
}

func searchBookmark(db *pgxpool.Pool) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var requestData SearchQuery
		var response ResponseMessage
		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			response.Message = badRequest
			w.Write(toJson(response))
			log.Print(err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		var responseData []Bookmark
		searchPattern := buildSearchPattern(requestData.Data)
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		rows, db_err := db.Query(context.Background(), "SELECT id, title, link, timestamp::TEXT, tag FROM bookmarks WHERE ts @@ to_tsquery('english', $1) OR ts @@ to_tsquery('simple', $1) LIMIT $2 OFFSET $3", searchPattern, pageLimit, page)
		if db_err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response.Message = dbError
			w.Write(toJson(response))
			log.Print(db_err.Error())
			return
		}
		for rows.Next() {
			var temp Bookmark
			err := rows.Scan(&temp.Id, &temp.Title, &temp.Link, &temp.Timestamp, &temp.Tag)
			if err != nil {
				log.Print(temp.Id, err.Error())
			}
			responseData = append(responseData, temp)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(toJson(responseData))
	}
	return http.HandlerFunc(fn)
}
