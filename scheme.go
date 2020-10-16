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

func readFromFile(f string) (map[string]*DBObject, map[string]*[]Pick) {
	b, _ := ioutil.ReadFile(f)
	info := DBInfo{}
	yaml.Unmarshal(b, &info)

	pull := make(map[string]*DBObject)
	picks := make(map[string]*[]Pick)

	for i := range info.Models {
		for _, f := range info.Models[i].Fields {
			if f.IsKey {
				info.Models[i].Key = f.ID
			}
			if f.IsLabel {
				info.Models[i].Label = f.ID
			}
		}
		pull[info.Models[i].ID] = &info.Models[i]
	}
	for i := range info.Picklists {
		picks[info.Picklists[i].ID] = &info.Picklists[i].Options
	}

	fixStringToType(pull)
	return pull, picks
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
				if _, ok := pull[test]; ok {
					// it seems we have an ID
					f.Ref = test
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
