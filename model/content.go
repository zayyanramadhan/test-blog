package model

import "time"

type Content struct {
	ID      int       `json:"id"`
	Title   string    `json:"title"`
	Content string    `json:"content"`
	Tag     []string  `json:"tags"`
	Publish time.Time `json:"publish"`
	Status  string    `json:"status"`
}
