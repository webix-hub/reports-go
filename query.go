package main

import (
	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
	"net/http"
	"strconv"
)

type QueryData struct {
	ID    int    `json:"id"`
	Text  string `json:"text"`
	Name  string `json:"name"`
}

type QueryDataResponse struct {
	ID int `json:"id"`
}

func queryAPI(r *chi.Mux, db *sqlx.DB) {
	r.Get("/api/queries", func(w http.ResponseWriter, r *http.Request) {
		temp := make([]QueryData, 0, 0)
		err := db.Select(&temp, "SELECT * FROM queries")
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}
		format.JSON(w, 200, temp)
	})

	r.Post("/api/queries", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		name := r.Form.Get("name")
		text := r.Form.Get("text")

		res, err := db.Exec("INSERT INTO queries(name, text) VALUES(?, ?)", name, text)

		var nid int64
		if err == nil {
			nid, err = res.LastInsertId()
		}
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, QueryDataResponse{int(nid)})
	})

	r.Put("/api/queries/{id}", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		id, _ := strconv.Atoi(chi.URLParam(r, "id"))
		name := r.Form.Get("name")
		text := r.Form.Get("text")

		_, err := db.Exec("UPDATE queries SET name = ?, text = ? WHERE id = ?", name, text, id)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, QueryDataResponse{id})
	})

	r.Delete("/api/queries/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(chi.URLParam(r, "id"))

		_, err := db.Exec("DELETE FROM queries WHERE id = ?", id)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, QueryDataResponse{id})
	})
}
