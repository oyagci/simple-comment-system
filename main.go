package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/lib/pq"
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

type CommentDB struct {
	id          string
	textfr      string
	texten      string
	publishedat string
	authorid    string
	targetid    string
	replies     []string
}

type OnCommentMessage struct {
	Message string `json:"message"`
	Author  string `json:"author"`
}

func query_comments(tx *sql.Tx, id string) ([]CommentDB, error) {

	comments := []CommentDB{}

	query_replies :=
		`WITH RECURSIVE x AS (
			SELECT id, textFr, textEn, publishedAt, authorId, targetId, replies
			FROM comments
			WHERE id = $1

			UNION ALL

			SELECT t.id, t.textFr, t.textEn, t.publishedAt, t.authorId, t.targetId, t.replies
			FROM x
			INNER JOIN comments AS t
			ON t.targetId = x.id
		) SELECT * FROM x;`

	if rows, err := tx.Query(query_replies, id); err == nil {
		for rows.Next() {
			var comment_db CommentDB
			rows.Scan(&comment_db.id, &comment_db.textfr, &comment_db.texten, &comment_db.publishedat, &comment_db.authorid, &comment_db.targetid, pq.Array(&comment_db.replies))
			comments = append(comments, comment_db)
		}
		rows.Close()
	}

	return comments, nil
}

func sort_comments_db(comment_db []CommentDB) []Comment {

	comments_db_map := map[string]CommentDB{}
	comments := []Comment{}

	for c := range comment_db {
		comments_db_map[comment_db[c].id] = comment_db[c]
	}

	for c := range comment_db {

		current := comment_db[c]

		comment := Comment{
			Id:          current.id,
			TextFr:      current.textfr,
			TextEn:      current.texten,
			PublishedAt: current.publishedat,
			AuthorId:    current.authorid,
			TargetId:    current.targetid,
		}

		for r := range current.replies {
			current_reply := current.replies[r]
			current_reply_db, present := comments_db_map[current_reply]

			if present {
				comment.Replies = append(comment.Replies, Comment{
					Id:          current_reply_db.id,
					TextFr:      current_reply_db.textfr,
					TextEn:      current_reply_db.texten,
					PublishedAt: current_reply_db.publishedat,
					AuthorId:    current_reply_db.authorid,
					TargetId:    current_reply_db.targetid,
					Replies:     []Comment{},
				})

				delete(comments_db_map, current_reply)
			}
		}

		if _, present := comments_db_map[comment.Id]; present {
			comments = append(comments, comment)
		}
	}

	return comments
}

func get_comments(w http.ResponseWriter, r *http.Request) {

	connStr := "user=postgres password=root dbname=postgres sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Println(err)
		return
	}

	defer func() {
		db.Close()
	}()

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	tx, err := db.Begin()

	if err != nil {
		log.Println(err)
		return
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	commentsDb, err := query_comments(tx, vars["targetId"])

	if err != nil {
		log.Println(err)
	}

	comments := sort_comments_db(commentsDb)

	json.NewEncoder(w).Encode(comments)
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
		log.Println(err)
		return
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare("INSERT INTO comments (id, textFr, textEn, publishedAt, authorId, targetId) VALUES($1, $2, $3, TO_TIMESTAMP($4), $5, $6);")

	if err != nil {
		log.Println(err)
		return
	}

	if _, err = stmt.Exec(c.Id, c.TextFr, c.TextEn, c.PublishedAt, c.AuthorId, c.TargetId); err != nil {
		log.Println(err)
		return
	}

	updateRepliesStmt, err := tx.Prepare("UPDATE comments SET replies = ARRAY_APPEND(replies, $1) WHERE id = $2")

	if _, err = updateRepliesStmt.Exec(c.Id, c.TargetId); err != nil {
		log.Println(err)
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
		db.Close()
	}()

	insert_comment_db(db, c)

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
			Replies:     []Comment{},
		}

		err := insert_comment(comment)

		if err != nil {
			w.WriteHeader(400)
			return
		}

		commentAsStr, err := json.Marshal(newComment)

		log.Println("Sending payload:", string(commentAsStr))

		onCommentMessage := OnCommentMessage{
			Message: string(commentAsStr),
			Author:  newComment.AuthorId,
		}

		jsonData, _ := json.Marshal(onCommentMessage)
		jsonReader := bytes.NewReader(jsonData)

		response, err := http.Post("http://tech-test-back.owlint.fr:8080/on_comment", "application/json", jsonReader)

		w.WriteHeader(response.StatusCode)
	} else {
		w.WriteHeader(400)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/target/{targetId}/comments", get_comments).Methods("GET")
	r.HandleFunc("/target/{targetId}/comments", post_new_comment).Methods("POST")

	log.Println("Listening on localhost:8080")
	log.Fatal(http.ListenAndServe("localhost:8080", r))
}
