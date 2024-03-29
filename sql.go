package main

import (
	"fmt"
	"strings"
)

type Sort struct {
	Field     string `json:"id"`
	Direction string `json:"mod"`
}

type Join struct {
	Source      string `json:"sid"`
	Target      string `json:"tid"`
	SourceField string `json:"sf"`
	TargetField string `json:"tf"`
}

func FromSQL(source string, joins []Join, allowed map[string]bool) string {
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

		if !allowed[j.Target+"."+tid] || !allowed[j.Source+"."+sid] {
			continue
		}

		source += fmt.Sprintf(" inner join %s on %s.%s = %s.%s ", j.Target, j.Source, sid, j.Target, tid)
	}

	return source
}

func SortSQL(by []Sort, allowed map[string]bool) string {
	out := ""
	for _, c := range by {
		dbName := getDBName(c.Field)
		if !allowed[dbName] {
			continue
		}

		var order string
		if c.Direction == "asc" {
			order = "ASC"
		} else {
			order = "DESC"
		}

		out += fmt.Sprintf(", `%s` %s", c.Field, order)
	}

	if len(out) == 0 {
		return ""
	}

	return out[1:]
}

func GroupSQL(by []string, buckets []Bucket, allowed map[string]bool) string {
	out := ""
	for _, c := range by {
		parts := strings.Split(c, ".")
		if len(parts) < 2 {
			continue
		}

		switch len(parts) {
		case 2:
			if !allowed[parts[0]+"."+parts[1]] {
				continue
			}
			if hasColumn(c, buckets) {
				out += fmt.Sprintf(", `%s.%s`", parts[0], parts[1])
			} else {
				out += fmt.Sprintf(", `%s`.`%s`", parts[0], parts[1])
			}
		default:
			if !allowed[parts[1]+"."+parts[2]] {
				continue
			}
			opStart, opEnd := sqlOperator(parts[0])
			if hasColumn(c, buckets) {
				out += fmt.Sprintf(", %s`%s.%s`%s", opStart, parts[1], parts[2], opEnd)
			} else {
				out += fmt.Sprintf(", %s`%s`.`%s`%s", opStart, parts[1], parts[2], opEnd)
			}
		}
	}

	if len(out) == 0 {
		return ""
	}

	return out[1:]
}

func SelectSQL(columns []string, buckets []Bucket, table, key string, allowed map[string]bool) string {
	out := ""

	for _, c := range columns {
		if hasColumn(c, buckets) {
			continue
		}

		var parts []string
		if c == "count." {
			parts = []string{"count", table, key}
		} else {
			parts = strings.Split(c, ".")
		}

		if len(parts) < 2 {
			continue
		}

		switch len(parts) {
		case 2:
			if !allowed[parts[0]+"."+parts[1]] {
				continue
			}
			out += fmt.Sprintf(", `%s`.`%s` as `%s`", parts[0], parts[1], c)
		default:
			if len(parts) < 2 || !allowed[parts[1]+"."+parts[2]] {
				continue
			}
			opStart, opEnd := aggregateOperator(parts[0])
			out += fmt.Sprintf(", %s`%s`.`%s`%s as `%s`", opStart, parts[1], parts[2], opEnd, c)
		}
	}

	if len(buckets) > 0 {
		for _, b := range buckets {
			if !hasBucket(b, columns) {
				continue
			}
			out += ", case "
			for _, o := range b.Options {
				if o.Values == nil {
					out += fmt.Sprintf("else '%s' ", o.Id)
					break
				} else {
					out += fmt.Sprintf("when %s in ('%s') then '%s' ", b.BucketColumn, strings.Join(o.Values, "','"), o.Id)
				}
			}
			out += fmt.Sprintf(" end as `%s` ", b.BucketColumn)
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
	case "avg":
		return "AVG(", ")"
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

func getDBName(n string) string {
	if strings.Count(n, "`") > 0 {
		return ""
	}

	if strings.Count(n, ".") == 2 {
		return n[strings.Index(n, ".")+1:]
	}

	return n
}

func hasColumn(column string, buckets []Bucket) bool {
	for _, b := range buckets {
		if b.BucketColumn == column {
			return true
		}
	}
	return false
}

func hasBucket(bucket Bucket, columns []string) bool {
	for _, c := range columns {
		if c == bucket.BucketColumn {
			return true
		}
	}
	return false
}
