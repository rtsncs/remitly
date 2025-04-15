package handler

import "github.com/rtsncs/remitly-swift-api/database"

type Handler struct {
	db *database.Database
}

type genericResponse struct {
	Message string `json:"message"`
}

func New(db *database.Database) Handler {
	return Handler{db}
}
