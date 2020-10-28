package main

import (
	"flag"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/unrolled/render"
	"log"
	"net/http"
)

var format = render.New()

type DBFieldType int

const (
	_ DBFieldType = iota
	UnknownField
	NumberField
	StringField
	DateField
	BoolField
	ReferenceField
	PickListField
)

var fieldTypeNames = map[DBFieldType]string{
	NumberField:    "number",
	StringField:    "text",
	DateField:      "date",
	BoolField:      "bool",
	ReferenceField: "reference",
	PickListField:  "picklist",
}

var backFieldTypeNames = map[string]DBFieldType{}

func init() {
	for key, value := range fieldTypeNames {
		backFieldTypeNames[value] = key
	}
}

type Pick struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

type PickList struct {
	ID      string
	Options []Pick
}

type DBField struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Filter   bool   `yaml:",omitempty" json:"filter"`
	Edit     bool   `yaml:",omitempty" json:"edit"`
	TypeName string `yaml:"type" json:"type"`
	Ref      string `json:"ref"`

	Type    DBFieldType `yaml:"-" json:"-"`
	IsKey   bool        `yaml:"key,omitempty" json:"key"`
	IsLabel bool        `yaml:"label,omitempty" json:"show"`
}

type Relation struct {
	To string `json:"to"`
}

func (d *DBField) GetTypeName() string {
	return fieldTypeNames[d.Type]
}

type DBReference struct {
	ID     int      `json:"id"`
	Target string   `json:"target"`
	Source string   `json:"source"`
	Name   string   `json:"name"`
	Field  *DBField `json:"-"`
}

type DBObject struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Fields     []DBField     `json:"data"`
	Key        string        `yaml:"-" json:"-"`
	Label      string        `yaml:"-" json:"-"`
	References []DBReference `yaml:"-" json:"refs"`
}

type DBInfo struct {
	Models    []DBObject
	Picklists []PickList
}

type RelationDetectionStrategy interface {
	IsRelation(name string, pull map[string]*DBObject) bool
}

type MatchByName struct{}

func (m MatchByName) IsRelation(name string, pull map[string]*DBObject) bool {
	return false
}

type MySQLField struct {
	Field   string  `db:"Field"`
	Type    string  `db:"Type"`
	Default *string `db:"Default"`
	Key     string  `db:"Key"`
	Null    string  `db:"Null"`
	Extra   string  `db:"Extra"`
}

var saveScheme = flag.String("save", "", "import scheme from DB and save to the file")
var loadScheme = flag.String("scheme", "scheme.yml", "path to file with scheme config")

var pull map[string]*DBObject
var picks map[string]*[]Pick

func main() {
	//config
	flag.Parse()
	Config.LoadFromFile("./config.yml")

	db, err := sqlx.Connect("mysql", Config.DataDBSourceName())
	appDB, err := sqlx.Connect("mysql", Config.AppDBSourceName())
	if err != nil {
		log.Fatal(err)
	}
	if *saveScheme != "" {
		pull := readFromDB(db)
		writeToFile(*saveScheme, pull)
		return
	}

	pull, picks = readFromFile(*loadScheme)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.Use(cors.Handler)

	metaAPI(r, appDB)
	dataAPI(r, db)
	queryAPI(r, appDB)
	moduleAPI(r, appDB, db)

	log.Printf("Start server at %s", Config.Server.Port)
	http.ListenAndServe(Config.Server.Port, r)
}

func getFieldInfo(table, field string) (*DBField, error) {
	t, ok := pull[table]
	if !ok {
		return nil, fmt.Errorf("table %s is unknown", table)
	}

	for i := range t.Fields {
		if t.Fields[i].ID == field {
			return &t.Fields[i], nil
		}
	}

	return nil, fmt.Errorf("table %s doesn't have %s field", table, field)
}
