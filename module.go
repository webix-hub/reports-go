package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
)

type RawData map[string]interface{}

type ModuleData struct {
	ID      int       `json:"id"`
	Text    string    `json:"text"`
	Name    string    `json:"name"`
	Updated time.Time `json:"updated"`
}

type ModuleDataResponse struct {
	ID int `json:"id"`
}

func bytesToString(m map[string]interface{}) {
	for k, v := range m {
		b, ok := v.([]byte)
		if ok {
			m[k] = string(b)
		}
	}
}

func moduleAPI(r *chi.Mux, db *sqlx.DB, dataDB *sqlx.DB) {
	r.Get("/api/modules", func(w http.ResponseWriter, r *http.Request) {
		temp := make([]ModuleData, 0, 0)
		err := db.Select(&temp, "SELECT * FROM modules")
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}
		format.JSON(w, 200, temp)
	})

	r.Post("/api/modules", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		name := r.Form.Get("name")
		text := r.Form.Get("text")

		res, err := db.Exec("INSERT INTO modules(name, text) VALUES(?, ?)", name, text)

		var nid int64
		if err == nil {
			nid, err = res.LastInsertId()
		}
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, ModuleDataResponse{int(nid)})
	})

	r.Put("/api/modules/{id}", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		id, _ := strconv.Atoi(chi.URLParam(r, "id"))
		name := r.Form.Get("name")
		text := r.Form.Get("text")

		_, err := db.Exec("UPDATE modules SET name = ?, text = ? WHERE id = ?", name, text, id)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, ModuleDataResponse{id})
	})

	r.Delete("/api/modules/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(chi.URLParam(r, "id"))

		_, err := db.Exec("DELETE FROM modules WHERE id = ?", id)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, ModuleDataResponse{id})
	})

}
