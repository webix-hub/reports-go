package main

import (
	"fmt"
	"strings"
)

type Join struct {
	Source      string `json:"sid"`
	Target      string `json:"tid"`
	SourceField string `json:"sf"`
	TargetField string `json:"tf"`
}

func FromSQL(source string, joins []Join) string {
	source = " from " + source

	for _, j := range joins {
		sid := j.SourceField
		if sid == "" {
			sid = pull[j.Source].Key
		}
		tid := j.TargetField
		if tid == "" {
			tid = pull[j.Target].Key
		}

		source += fmt.Sprintf(" inner join %s on %s.%s = %s.%s ", j.Target, j.Source, sid, j.Target, tid)
	}

	return source
}

func GroupSQL(by []string) string {
	out := ""
	for _, c := range by {
		parts := strings.Split(c, ".")
		switch len(parts) {
		case 3:
			opStart, opEnd := sqlOperator(parts[0])
			out += fmt.Sprintf(", %s`%s`.`%s`%s", opStart, parts[1], parts[2], opEnd)
		case 2:
			out += fmt.Sprintf(", `%s`.`%s`", parts[0], parts[1])
		case 1:
			out += fmt.Sprintf(", `%s`", parts[0])
		}
	}

	return out[1:]
}

func SelectSQL(columns []string, table, key string) string {
	out := ""
	for _, c := range columns {
		var parts []string
		if c == "count."{
			parts = []string{ "count", table, key }
		} else {
			parts = strings.Split(c, ".")
		}

		switch len(parts) {
		case 3:
			opStart, opEnd := aggregateOperator(parts[0])
			out += fmt.Sprintf(", %s`%s`.`%s`%s as `%s`", opStart, parts[1], parts[2], opEnd, c)
		case 2:
			out += fmt.Sprintf(", `%s`.`%s` as `%s`", parts[0], parts[1], c)
		case 1:
			out += fmt.Sprintf(", `%s` as `%s`", parts[0], c)
		}
	}
	return fmt.Sprintf("select %s ", out[1:])
}

func aggregateOperator(code string) (string, string) {
	switch code {
	case "sum":
		return "SUM(", ")"
	case "max":
		return "MAX(", ")"
	case "min":
		return "MIN(", ")"
	case "average":
		return "AVERAGE(", ")"
	case "count":
		return "COUNT(", ")"
	default:
		return sqlOperator(code)
	}
}

func sqlOperator(code string) (string, string) {
	switch code {
	case "year":
		return `DATE_FORMAT(`, `, "%Y")`
	case "month":
		return `DATE_FORMAT(`, `, "%M")`
	case "day":
		return `DATE_FORMAT(`, `, "%d")`
	case "yearmonth":
		return `DATE_FORMAT(`, `, "%m/%Y")`
	case "yearmonthday":
		return `DATE_FORMAT(`, `, "%d %M %Y")`
	default:
		return "", ""
	}
}
