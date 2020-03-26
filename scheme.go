package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"strings"
)

func writeToFile(f string, pull map[string]*DBObject) {
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
		}
	}
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
				ID:       f.Field,
				Name:     f.Field,
				Type:     TypeToField(f.Type),
				IsKey:    f.Key == "PRI",
				Relation: nil,
				Filter:   true,
				Edit:     false,
			}
		}
		pull[t] = &DBObject{
			ID:     t,
			Name:   t,
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
					f.Relation = &Relation{To: test, Name: getFirstStringLike(target)}
					f.Type = ReferenceField
				}

				table.Fields[i] = f
			}
		}
	}

	return pull
}

func getFirstStringLike(target *DBObject) string {
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
