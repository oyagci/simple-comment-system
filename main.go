package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

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

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/target/{targetId}/comments", get_comments).Methods("GET")
	r.HandleFunc("/target/{targetId}/comments", post_new_comment).Methods("POST")

	log.Println("Listening on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
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

var commentIdx = 0

func post_new_comment(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var newComment NewComment

	_ = json.NewDecoder(r.Body).Decode(&newComment)

	commentIdxStr := fmt.Sprint(commentIdx)

	comment := Comment{
		Id:          "Comment-" + commentIdxStr,
		TextFr:      newComment.TextFr,
		TextEn:      newComment.TextEn,
		PublishedAt: "0",
		AuthorId:    newComment.AuthorId,
		TargetId:    newComment.TargetId,
	}

	commentIdx += 1

	if newComment.TargetId != "" {
		allComments = append(allComments, comment)
	}

	json.NewEncoder(w).Encode(allComments)
}
