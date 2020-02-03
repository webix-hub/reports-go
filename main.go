package main

import (
	"flag"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/unrolled/render"
	"gopkg.in/yaml.v2"
	"net/http"

	"io/ioutil"
	"log"
	"strings"
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

var fieldTypeNames = map[DBFieldType]string {
	NumberField : "number",
	StringField : "string",
	DateField: "date",
	BoolField: "bool",
	ReferenceField: "reference",
	PickListField: "picklist",
}

var backFieldTypeNames = map[string]DBFieldType{}
func init(){
	for key, value := range fieldTypeNames {
		backFieldTypeNames[value] = key
	}
}


type PickList struct {
	ID string `json:"id"`
	Value string `json:"value"`
}
type DBField struct {
	ID string `json:"id"`
	Name string	`json:"name"`
	Type DBFieldType `yaml:"-",json:"-"`
	TypeName string `yaml:"type",json:"type"`
	Relation *Relation `yaml:",omitempty",json:"-"`
	RelationName string `yaml:"-",json:"relation"`
	IsKey bool `yaml:",omitempty",json:"-"`
	Filter bool  `yaml:",omitempty",json:"filter"`
	Edit bool  `yaml:",omitempty",json:"edit"`
}

type Relation struct {
	To string
	Name string
}

func fixStringToType(pull map[string]*DBObject) {
	for _, table := range pull {
		for i := range table.Fields {
			table.Fields[i].Type = backFieldTypeNames[table.Fields[i].TypeName]
		}
	}
}

func fixTypeToString(pull map[string]*DBObject) {
	for _, table := range pull {
		for i := range table.Fields {
			table.Fields[i].TypeName = fieldTypeNames[table.Fields[i].Type]
			if table.Fields[i].Relation != nil {
				table.Fields[i].RelationName = table.Fields[i].Relation.Name
			}
		}
	}
}


func (d *DBField) GetTypeName() string {
	return fieldTypeNames[d.Type]
}

type DBObject struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Fields []DBField `json:"-"`
}

type RelationDetectionStrategy interface {
	IsRelation(name string, pull map[string]*DBObject) bool
}

type MatchByName struct {}
func (m MatchByName) IsRelation(name string, pull map[string]*DBObject) bool {
	return false
}

type MySQLField struct {
	Field string `db:"Field"`
	Type string `db:"Type"`
	Default *string `db:"Default"`
	Key string `db:"Key"`
	Null string `db:"Null"`
	Extra string `db:"Extra"`
}

var saveScheme = flag.String("save", "", "import scheme from DB and save to the file")
var loadScheme = flag.String("scheme", "scheme.yml", "path to file with scheme config")

func main(){

	//config
	Config.LoadFromFile("./config.yml")

	db, err := sqlx.Connect("mysql",Config.DataSourceName())
	if err != nil {
		log.Fatal(err)
	}

	if *saveScheme != "" {
		pull := readFromDB(db);
		writeScheme(*saveScheme, pull)
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

		id := chi.URLParam(r,"id")
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
		if field.Relation == nil || field.Type != PickListField {
			format.Text(w, 404, fmt.Sprintf("Can't use [%s@%s] as picklist", object, id))
			return
		}


		list := []PickList{}
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

func writeScheme(f string, pull map[string]*DBObject){
	fixTypeToString(pull)
	scheme, _ := yaml.Marshal(pull)
	ioutil.WriteFile(f, scheme, 0777)
}

func readFromFile(f string) map[string]*DBObject {
	b, _ := ioutil.ReadFile(f)
	pull := make(map[string]*DBObject)
	yaml.Unmarshal(b, &pull)

	fixStringToType(pull)

	return pull
}

func readFromDB(db *sqlx.DB) map[string]*DBObject {
	tables := []string{}
	_ = db.Select(&tables, "SHOW tables")

	pull := make(map[string]*DBObject)
	temp := make([]MySQLField, 0, 50)

	for _, t := range tables {
		err := db.Select(&temp, fmt.Sprintf("DESCRIBE `%s`", t))
		if err != nil {
			log.Fatal(err)
		}
		fields := make([]DBField, len(temp))
		for i, f := range temp {
			fields[i] = DBField{
				ID: f.Field,
				Name: f.Field,
				Type: TypeToField(f.Type),
				IsKey: f.Key == "PRI",
				Relation: nil,
				Filter:true,
				Edit: false,
			}
		}
		pull[t] = &DBObject{
			ID:t,
			Name: t,
			Fields: fields,
		}
		temp = temp[:0]
	}

	// fix relations
	for _, table := range pull {
		for i, f := range table.Fields {
			if strings.HasSuffix(f.Name, "_id") {
				test := strings.TrimSuffix(f.Name, "_id")
				if target, ok := pull[test]; ok {
					// it seems we have an ID
					f.Relation = &Relation{To:test, Name: getFirstStringLike(target) }
					f.Type = ReferenceField
				}

				table.Fields[i] = f
			}
		}
	}

	return pull
}

func getFirstStringLike(target *DBObject) string{
	for _, f := range target.Fields {
		if f.Type == StringField {
			return f.ID
		}
	}
	return target.Fields[0].ID
}

func TypeToField(t string) DBFieldType {
	if strings.HasPrefix(t, "int") {
		return NumberField
	}

	if strings.HasPrefix(t, "date") {
		return DateField
	}

	return StringField
}