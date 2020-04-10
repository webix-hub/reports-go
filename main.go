package main

import (
	"flag"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/go-chi/chi/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/unrolled/render"
	"io/ioutil"
	"net/http"
	"time"

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

var pull map[string]*DBObject

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

	pull = readFromFile(*loadScheme)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.Use(cors.Handler)

	// list of tables
	objects := make([]*DBObject, 0, len(pull))
	for _, v := range pull {
		objects = append(objects, v)
	}


	r.Post("/api/data/{table}", func(w http.ResponseWriter, r *http.Request) {
		table := chi.URLParam(r, "table")

		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		data, err := getDataFromDB(db, table, body)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		format.JSON(w, 200, data)
	})

	r.Get("/api/data/{table}/{field}/suggest", func(w http.ResponseWriter, r *http.Request) {
		table := chi.URLParam(r, "table")
		field := chi.URLParam(r, "field")


		f, err := getFieldInfo(table, field)
		if err != nil {
			format.Text(w, 500, err.Error())
			return
		}

		sql := fmt.Sprintf("select distinct %s from %s ORDER BY %s ASC", field, table, field)
		if f.Type == StringField {
			out := make([]string, 0)
			err := db.Select(&out, sql)
			if err !=nil {
				fmt.Printf("%+v", err)
			}
			format.JSON(w, 200, out)
			return
		}

		if f.Type == NumberField {
			out := make([]float64, 0)
			err := db.Select(&out, sql)
			if err !=nil {
				log.Printf("%+v", err)
			}
			format.JSON(w, 200, out)
			return
		}

		if f.Type == DateField {
			out := make([]time.Time, 0)
			err := db.Select(&out, sql)
			if err !=nil {
				log.Printf("%+v", err)
			}
			format.JSON(w, 200, out)
			return
		}

		format.JSON(w, 200, []string{})
	})

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


func getFieldInfo(table, field string) (*DBField, error){
	t,ok := pull[table]
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