package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "github.com/lib/pq"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Comment struct {
	Id          string    `json:"id"`
	TextFr      string    `json:"textFr"`
	TextEn      string    `json:"textEn"`
	PublishedAt string    `json:"publishedAt"`
	AuthorId    string    `json:"authorId"`
	TargetId    string    `json:"targetId"`
	Replies     []Comment `json:"replies"`
}

type NewComment struct {
	TextFr      string `json:"textFr"`
	TextEn      string `json:"textEn"`
	PublishedAt string `json:"publishedAt"`
	AuthorId    string `json:"authorId"`
	TargetId    string `json:"targetId"`
}

var allComments []Comment = []Comment{
	{
		Id:          "Comment-kjh784fgevdhhdwhh7563",
		TextFr:      "Bonjour ! je suis un commentaire.",
		TextEn:      "Hi ! Im a comment.",
		PublishedAt: "1639477064",
		AuthorId:    "User-kjh784fgevdhhdwhh7563",
		TargetId:    "Photo-bdgetr657434hfggrt8374",
		Replies: []Comment{
			{
				Id:          "Comment-1234abcd",
				TextFr:      "Je suis une réponse au commentaire",
				TextEn:      "Im a reply!",
				PublishedAt: "1639477064",
				AuthorId:    "User-5647565dhfbdshs",
				TargetId:    "Comment-kjh784fgevdhhdwhh7563",
			},
			{
				Id:          "Comment-5678efgh",
				TextFr:      "Je suis une autre reponse !",
				TextEn:      "Im another reply!",
				PublishedAt: "1639477064",
				AuthorId:    "User-5342hdfgetrfiw789",
				TargetId:    "Comment-kjh784fgevdhhdwhh7563",
			},
		},
	},
}

func get_matching_comments(comments *[]Comment, targetId string) []Comment {

	var matches []Comment

	for i := range *comments {
		if (*comments)[i].TargetId == targetId {
			matches = append(matches, (*comments)[i])
		}
	}

	return matches
}

func get_comments(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	matchingComments := get_matching_comments(&allComments, vars["targetId"])

	json.NewEncoder(w).Encode(matchingComments)
}

func is_valid_new_comment(newComment *NewComment) bool {
	return newComment.TextFr != "" &&
		newComment.TextEn != "" &&
		newComment.AuthorId != "" &&
		newComment.PublishedAt != "" &&
		newComment.TargetId != ""
}

func generate_comment_uuid() string {
	uuidHyphen := uuid.New()
	return strings.Replace(uuidHyphen.String(), "-", "", -1)
}

func insert_comment_db(db *sql.DB, c Comment) {
	tx, err := db.Begin()

	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		switch err {
		case nil:
			log.Println("Committing")
			err = tx.Commit()
		default:
			log.Println("Rolling back")
			tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare("INSERT INTO comments (id, textFr, textEn, publishedAt, authorId, targetId, replies) VALUES($1, $2, $3, TO_TIMESTAMP($4), $5, $6, $7);")

	if err != nil {
		fmt.Println(err)
		return
	}

	if _, err = stmt.Exec(c.Id, c.TextFr, c.TextEn, c.PublishedAt, c.AuthorId, c.TargetId, nil); err != nil {
		fmt.Println(err)
		return
	}

	log.Println("Comment inserted")

	return
}

func insert_comment(c Comment) error {

	connStr := "user=postgres password=root dbname=postgres sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Println(err)
		return err
	}

	defer func() {
		log.Println("Closing connection")
		db.Close()
	}()

	insert_comment_db(db, c)

	allComments = append(allComments, c)

	return nil
}

func post_new_comment(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var newComment NewComment

	_ = json.NewDecoder(r.Body).Decode(&newComment)

	uuid := generate_comment_uuid()

	if is_valid_new_comment(&newComment) {
		comment := Comment{
			Id:          "Comment-" + uuid,
			TextFr:      newComment.TextFr,
			TextEn:      newComment.TextEn,
			PublishedAt: newComment.PublishedAt,
			AuthorId:    newComment.AuthorId,
			TargetId:    newComment.TargetId,
		}

		err := insert_comment(comment)

		if err == nil {
			w.WriteHeader(200)
		}
	} else {
		w.WriteHeader(400)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/target/{targetId}/comments", get_comments).Methods("GET")
	r.HandleFunc("/target/{targetId}/comments", post_new_comment).Methods("POST")

	log.Println("Listening on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
