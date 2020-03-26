package main

import (
	"flag"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/unrolled/render"
	"net/http"

	"log"
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
	StringField:    "string",
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
type PickList []Pick

type DBField struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Filter   bool      `yaml:",omitempty" json:"filter"`
	Edit     bool      `yaml:",omitempty" json:"edit"`
	TypeName string    `yaml:"type" json:"type"`
	Relation *Relation `yaml:",omitempty" json:"relation"`

	Type     DBFieldType `yaml:"-" json:"-"`
	PickList PickList    `yaml:",omitempty" json:"-"`
	IsKey    bool        `yaml:",omitempty" json:"-"`
}

type Relation struct {
	To   string `json:"to"`
	Name string `json:"-"`
}

func (d *DBField) GetTypeName() string {
	return fieldTypeNames[d.Type]
}

type DBObject struct {
	ID     string    `json:"id"`
	Name   string    `json:"name"`
	Fields []DBField `json:"-"`
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

func main() {
	//config
	flag.Parse()
	Config.LoadFromFile("./config.yml")

	db, err := sqlx.Connect("mysql", Config.DataSourceName())
	if err != nil {
		log.Fatal(err)
	}
	if *saveScheme != "" {
		pull := readFromDB(db)
		writeToFile(*saveScheme, pull)
		return
	}

	pull := readFromFile(*loadScheme)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// list of tables
	objects := make([]*DBObject, 0, len(pull))
	for _, v := range pull {
		objects = append(objects, v)
	}
	r.Get("/api/objects", func(w http.ResponseWriter, r *http.Request) {
		format.JSON(w, 200, objects)
	})

	// list of fields
	r.Get("/api/fields/{object}", func(w http.ResponseWriter, r *http.Request) {
		table, ok := pull[chi.URLParam(r, "object")]
		if !ok {
			format.Text(w, 404, "Object not found")
		} else {
			format.JSON(w, 200, table.Fields)
		}
	})

	// options
	r.Get("/api/options/{object}/{id}", func(w http.ResponseWriter, r *http.Request) {
		object := chi.URLParam(r, "object")
		table, ok := pull[object]
		if !ok {
			format.Text(w, 404, fmt.Sprintf("Object [%s] not found", object))
			return
		}

		id := chi.URLParam(r, "id")
		var field DBField
		for _, f := range table.Fields {
			if f.ID == id {
				field = f
			}
		}

		if field.ID == "" {
			format.Text(w, 404, fmt.Sprintf("Field [%s@%s] not found", object, id))
			return
		}
		if field.Relation == nil && field.PickList == nil {
			format.Text(w, 404, fmt.Sprintf("Can't use [%s@%s] as picklist", object, id))
			return
		}

		list := PickList{}
		sql := fmt.Sprintf("SELECT `id`,`%s` as value FROM `%s`", field.Relation.Name, field.Relation.To)
		err := db.Select(&list, sql)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, list)
	})

	log.Printf("Start server at %s", Config.Server.Port)
	http.ListenAndServe(Config.Server.Port, r)
}
