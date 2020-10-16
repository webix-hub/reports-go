package main

import (
	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
	"net/http"
)

func metaAPI(r *chi.Mux, db *sqlx.DB) {

	r.Get("/api/objects", func(w http.ResponseWriter, r *http.Request) {
		format.JSON(w, 200, objects)
	})

	// list of fields
	r.Get("/api/objects/{object}/fields", func(w http.ResponseWriter, r *http.Request) {
		table, ok := pull[chi.URLParam(r, "object")]
		if !ok {
			format.Text(w, 404, "Object not found")
		} else {
			format.JSON(w, 200, table.Fields)
		}
	})
}
