package main

import (
	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
	"net/http"
	"strconv"
)

type QueryData struct {
	ID int `json:"id"`
	ObjID string `db:"obj_id" json:"obj_id"`
	Text string `json:"text"`
	Name string `json:"name"`
}

type QueryDataResponse struct {
	ID int `json:"id"`
}

func queryAPI(r *chi.Mux, db *sqlx.DB){
	r.Get("/api/objects/{oid}/queries", func(w http.ResponseWriter, r *http.Request) {
		oid := chi.URLParam(r, "oid")
		temp := make([]QueryData, 0, 0)
		err := db.Select(&temp, "SELECT * FROM queries WHERE obj_id = ?", oid)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}
		format.JSON(w, 200, temp)
	})

	r.Post("/api/objects/{oid}/queries", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		oid := chi.URLParam(r, "oid")
		name := r.Form.Get("name")
		text := r.Form.Get("text")

		res, err := db.Exec("INSERT INTO queries(obj_id, name, text) VALUES(?, ?, ?)", oid, name, text)

		var nid int64
		if err == nil {
			nid, err = res.LastInsertId()
		}
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}


		format.JSON(w, 200, QueryDataResponse{ int(nid)})
	})

	r.Put("/api/objects/{oid}/queries/{id}", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		oid := chi.URLParam(r, "oid")
		id, _ := strconv.Atoi(chi.URLParam(r, "id"))
		name := r.Form.Get("name")
		text := r.Form.Get("text")

		_, err := db.Exec("UPDATE queries SET name = ?, text = ? WHERE obj_id = ? AND id = ?", name, text, oid, id)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, QueryDataResponse{id})
	})

	r.Delete("/api/objects/{oid}/queries/{id}", func(w http.ResponseWriter, r *http.Request) {
		oid := chi.URLParam(r, "oid")
		id, _ := strconv.Atoi(chi.URLParam(r, "id"))

		_, err := db.Exec("DELETE FROM queries WHERE obj_id = ? AND id = ?", oid, id)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, QueryDataResponse{id})
	})
}