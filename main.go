package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"test-blog/database"
	"test-blog/model"
	"time"

	"github.com/joho/godotenv"
)

var db *sql.DB

func main() {

	database.InitDB()

	database.CreateEnumType()

	database.CreateTables()

	var err error

	if os.Getenv("GO_ENV") == "local" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	user := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable", user, password, dbname, host, port)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Successfully connected to the database")

	defer db.Close()

	http.HandleFunc("/api/posts/", handlePost)

	log.Println("Server started on localhost:7878")
	log.Fatal(http.ListenAndServe(":7878", nil))

}

func handlePost(w http.ResponseWriter, r *http.Request) {
	// Extract the ID parameter from the request URL
	idStr := r.URL.Path[len("/api/posts/"):]
	id, err := strconv.Atoi(idStr)

	switch r.Method {
	case http.MethodGet:
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		getContents(w, id)
	case http.MethodPost:
		createContent(w, r)
	case http.MethodPut:
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		updateContent(w, r, id)
	case http.MethodDelete:
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}
		deleteContent(w, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getContents(w http.ResponseWriter, id int) {
	rows, err := db.Query("SELECT content.id, content.title, content.content, content.status, content.publish_date, tag.label as tag FROM content left join content_tag on content.id = content_tag.content_id left join tag on tag.id = content_tag.tag_id where content.id = $1 ", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	contentsMap := make(map[int]*model.Content)
	for rows.Next() {
		var ContentID int
		var Title string
		var Content string
		var Status string
		var Label sql.NullString
		var Publish time.Time

		if err := rows.Scan(&ContentID, &Title, &Content, &Status, &Publish, &Label); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		content, ok := contentsMap[ContentID]
		if !ok {
			content = &model.Content{
				ID:      ContentID,
				Title:   Title,
				Content: Content,
				Tag:     []string{},
				Status:  Status,
			}
			contentsMap[ContentID] = content
		}

		if Label.Valid {
			content.Tag = append(content.Tag, Label.String)
		}
	}

	var contents []model.Content
	for _, content := range contentsMap {
		contents = append(contents, *content)
	}

	Response := model.ResponseSuccessData{
		Message: "success",
		Data:    contents,
	}

	jsonResponse, err := json.Marshal(Response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)

}

func createContent(w http.ResponseWriter, r *http.Request) {
	var content model.Content
	if err := json.NewDecoder(r.Body).Decode(&content); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if strings.ToLower(content.Status) == "draft" || strings.ToLower(content.Status) == "publish" {

	} else if strings.ToLower(content.Content) == "" || strings.ToLower(content.Title) == "" || strings.ToLower(content.Status) == "" {
		Response := model.Response{
			Message: "error input",
		}

		if strings.ToLower(content.Status) == "" || strings.ToLower(content.Status) != "draft" || strings.ToLower(content.Status) != "publish" {
			Response = model.Response{
				Message: "error input Status must (draft / publish)",
			}
		}

		jsonResponse, err := json.Marshal(Response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
		return
	}

	var lastIDresultContent int

	err := db.QueryRow("INSERT INTO content (title, content, status) VALUES ($1, $2, $3) RETURNING id", content.Title, content.Content, strings.ToLower(content.Status)).Scan(&lastIDresultContent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	var listIDTag []int

	if content.Tag != nil {
		for _, s := range content.Tag {
			var IDTag int
			err := db.QueryRow("SELECT id FROM tag where LOWER(label)=$1", strings.ToLower(s)).Scan(&IDTag)
			if err != nil {
				err := db.QueryRow("INSERT INTO tag (label) VALUES ($1) RETURNING id", s).Scan(&IDTag)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				listIDTag = append(listIDTag, int(IDTag))
			} else {
				listIDTag = append(listIDTag, int(IDTag))
			}
		}
	}

	for _, idTag := range listIDTag {
		_, err := db.Exec("INSERT INTO content_tag (content_id, tag_id) VALUES ($1, $2)", lastIDresultContent, idTag)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	Response := model.Response{
		Message: "success create",
	}

	jsonResponse, err := json.Marshal(Response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func updateContent(w http.ResponseWriter, r *http.Request, id int) {
	var content model.Content
	if err := json.NewDecoder(r.Body).Decode(&content); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if strings.ToLower(content.Status) == "draft" || strings.ToLower(content.Status) == "publish" {

	} else if strings.ToLower(content.Content) == "" || strings.ToLower(content.Title) == "" || strings.ToLower(content.Status) == "" {
		Response := model.Response{
			Message: "error input",
		}

		if strings.ToLower(content.Status) == "" || strings.ToLower(content.Status) != "draft" || strings.ToLower(content.Status) != "publish" {
			Response = model.Response{
				Message: "error input Status must (draft / publish)",
			}
		}

		jsonResponse, err := json.Marshal(Response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonResponse)
		return
	}

	_, err := db.Exec("UPDATE content SET title = $1, content = $2, status = $3 WHERE id = $4", content.Title, content.Content, strings.ToLower(content.Status), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("DELETE FROM content_tag WHERE content_id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var listIDTag []int

	if content.Tag != nil {
		for _, s := range content.Tag {
			var IDTag int
			err := db.QueryRow("SELECT id FROM tag where LOWER(label)=$1", strings.ToLower(s)).Scan(&IDTag)
			if err != nil {
				err := db.QueryRow("INSERT INTO tag (label) VALUES ($1) RETURNING id", s).Scan(&IDTag)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				listIDTag = append(listIDTag, int(IDTag))
			} else {
				listIDTag = append(listIDTag, int(IDTag))
			}
		}
	}

	for _, idTag := range listIDTag {
		_, err := db.Exec("INSERT INTO content_tag (content_id, tag_id) VALUES ($1, $2)", id, idTag)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	Response := model.Response{
		Message: "success update",
	}

	jsonResponse, err := json.Marshal(Response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func deleteContent(w http.ResponseWriter, id int) {
	_, err := db.Exec("DELETE FROM content WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("DELETE FROM content_tag WHERE content_id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	Response := model.Response{
		Message: "success delete",
	}

	jsonResponse, err := json.Marshal(Response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
