package main

import "fmt"

type Join struct {
	Source string `json:"sid"`
	Target string `json:"tid"`
	SourceField string `json:"sf"`
	TargetField string `json:"tf"`
}

func FromSQL(source string, joins []Join) string {
	source = " from "+source

	for _, j := range joins {
		sid := j.SourceField
		if sid == "" {
			sid = pull[j.Source].Key
		}
		tid := j.SourceField
		if tid == "" {
			tid = pull[j.Source].Key
		}

		source += fmt.Sprintf(" inner join %s on %s.%s = %s.%s ", j.Target, j.Source, sid, j.Target, tid)
	}

	return source
}

func SelectSQL(columns []string) string {
	fields := ""
	for _, c := range columns {
		fields += fmt.Sprintf(", %s as `%s` ", c, c)
	}
	return fmt.Sprintf("select %s ", fields[1:])
}