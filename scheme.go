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

	info := DBInfo{}
	for _, s := range pull {
		info.Models = append(info.Models, *s)
	}

	scheme, _ := yaml.Marshal(info)
	ioutil.WriteFile(f, scheme, 0777)
}

func readFromFile(f string) (map[string]*DBObject, map[string]*[]Pick) {
	b, _ := ioutil.ReadFile(f)
	info := DBInfo{}
	yaml.Unmarshal(b, &info)

	pull := make(map[string]*DBObject)
	picks := make(map[string]*[]Pick)
	refs := make(map[string][]DBReference)

	idCounter := 1
	for i := range info.Models {
		t := &info.Models[i]
		for j := range info.Models[i].Fields {
			f := &info.Models[i].Fields[j]

			f.Type = backFieldTypeNames[t.Fields[j].TypeName]
			if f.IsKey {
				t.Key = f.ID
			}
			if f.IsLabel {
				t.Label = f.ID
			}

			if f.Type == ReferenceField {
				link := DBReference{ID: idCounter, Source: t.ID, Target: f.Ref, Field: &info.Models[i].Fields[j]}
				idCounter += 1

				if _, ok := refs[t.ID]; ok {
					refs[t.ID] = append(refs[t.ID], link)
				} else {
					refs[t.ID] = []DBReference{link}
				}

				if _, ok := refs[f.Ref]; ok {
					refs[f.Ref] = append(refs[f.Ref], link)
				} else {
					refs[f.Ref] = []DBReference{link}
				}
			}
		}

		pull[info.Models[i].ID] = &info.Models[i]
	}

	// store references on objects
	for _, s := range pull {
		s.References = refs[s.ID]
		if s.References != nil {
			for i, r := range s.References {
				s.References[i].Name = r.Field.Name
			}
		}
	}

	for i := range info.Picklists {
		picks[info.Picklists[i].ID] = &info.Picklists[i].Options
	}

	return pull, picks
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
				ID:     f.Field,
				Name:   strings.TrimRight(f.Field, "_id"),
				Type:   TypeToField(f.Type),
				IsKey:  f.Key == "PRI",
				Filter: true,
				Edit:   false,
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
		hasLabel := false
		for i, f := range table.Fields {
			if strings.HasSuffix(f.ID, "_id") {
				test := strings.TrimSuffix(f.Name, "_id")
				if _, ok := pull[test]; ok {
					// it seems we have an ID
					table.Fields[i].Ref = test
					table.Fields[i].Type = ReferenceField
				} else if _, ok := pull[test+"s"]; ok {
					// it seems we have an ID
					table.Fields[i].Ref = test
					table.Fields[i].Type = ReferenceField
				} else if _, ok := pull[test+"es"]; ok {
					// it seems we have an ID
					table.Fields[i].Ref = test
					table.Fields[i].Type = ReferenceField
				}
			}

			if f.Type == StringField && !hasLabel {
				table.Fields[i].IsLabel = true
				hasLabel = true
			}
		}

		if !hasLabel {
			table.Fields[0].IsLabel = true
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
	if strings.HasPrefix(t, "decimal") {
		return NumberField
	}
	if strings.HasPrefix(t, "date") {
		return DateField
	}

	return StringField
}
