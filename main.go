package main

import (
	"github.com/jmoiron/sqlx"
)

type DBFieldType int
const (
	_ DBFieldType = iota
	UnknownField
	NumberField
	StringField
	DateField
	BoolField
	Reference
)

var fieldTypeNames = map[DBFieldType]string {
	NumberField : "number",
	StringField : "string",
	DateField: "date",
	BoolField: "bool",
	Reference: "reference",
}

type DBField struct {
	Name string
	Type DBFieldType
	Relation *DBRelation
}

type DBRelationType int
const (
	_ DBRelationType = iota
	OneToMany
	ManyToOne
	ManyToMany
)

type DBRelation struct {
	LeftObject DBObject
	LeftField DBField
	RightObject DBObject
	RightField DBField
}

func (d *DBField) GetTypeName() string {
	return fieldTypeNames[d.Type]
}

type DBObject struct {
	Fields []DBField
}

type RelationDetectionStrategy interface {
	IsRelation(name string, pull map[string]*DBObject) bool
}

type MatchByName struct {}
func (m MatchByName) IsRelation(name string, pull map[string]*DBObject) bool {
	return false
}

func main(){

	db, _ := sqlx.Connect("postgres", "user=foo dbname=bar sslmode=disable")

	tables := []string{}
	db.Select(&tables, "SHOW tables")

	pull := make(map[string]*DBObject)
	for t := range tables {
		db.Select("DESCRIBE `t`")
		fields := make([]DBFields, 0)
		pull["a"] = &DBObject{
			Fields: fields,
		}
	}
}
