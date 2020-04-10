package main

import (
	"github.com/jmoiron/sqlx"
	"log"
)
import "github.com/xbsoftware/querysql"

type PersonData struct {
	A0 int `json:"id" db:"id"`
	A1 string `json:"last_name" db:"last_name"`
	A2 string `json:"first_name" db:"first_name"`
	A3 string `json:"birthdate" db:"birthdate"`
	A4 string `json:"country" db:"country"`
	A5 string `json:"city" db:"city"`
	A6 string `json:"email" db:"email"`
	A7 string `json:"phone" db:"phone"`
	A8 string `json:"job" db:"job"`
	A9 string `json:"address" db:"address"`
	A10 int `json:"company_id" db:"company_id"`
	A11 int `json:"notify" db:"notify"`
	A12 int `json:"age" db:"age"`
}

func getDataFromDB(db *sqlx.DB, name string, body []byte) ([]interface{}, error) {
	var filter = querysql.Filter{}
	var err error

	if len(body) > 0 {
		filter, err = querysql.FromJSON(body)
		if err != nil {
			return nil, err
		}
	}

	query, data, err := querysql.GetSQL(filter, nil)
	if err != nil {
		return nil, err
	}

	t := make([]PersonData, 0)
	sql := "select * from " + name
	if query != "" {
		sql += " where "+query
	}

	log.Println(sql)

	err = db.Select(&t, sql, data...)
	if err != nil {
		return nil, err
	}

	 ret := make([]interface{}, len(t))
	 for i := range t {
	 	ret[i] = t[i]
	 }

	 return ret, nil
}