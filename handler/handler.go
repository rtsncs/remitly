package handler

import "github.com/rtsncs/remitly-swift-api/database"

type Handler struct {
	db *database.Database
}

func New(db *database.Database) Handler {
	return Handler{db}
}
