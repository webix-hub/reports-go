package main

import (
	"github.com/jmoiron/sqlx"
)

func InitDB(db *sqlx.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS modules (
  		id int NOT NULL AUTO_INCREMENT,
  		name varchar(255) NOT NULL,
  		text text NOT NULL,
  		updated datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  		PRIMARY KEY (id)
  	)`);
	if err != nil {
		return err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS queries (
  		id int NOT NULL AUTO_INCREMENT,
  		text text,
  		name varchar(255) DEFAULT NULL,
  		PRIMARY KEY (id)
  	)`);
	if err != nil {
		return err
	}

	return nil
}
