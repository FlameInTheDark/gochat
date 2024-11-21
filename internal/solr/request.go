package solr

import (
	"fmt"
	"strconv"
	"strings"
)

type Querier interface {
	toQuery() string
}

type QueryBuilder struct {
	Q  *string  `json:"query,omitempty"`
	Fq []string `json:"filter,omitempty"`
	Fl []string `json:"fields,omitempty"`
	L  *int     `json:"limit,omitempty"`
	O  *int     `json:"offset,omitempty"`
}

var Builder = QueryBuilder{}

// M map of values 'test:value test2:value2'
type M map[string]interface{}

func (m M) toQuery() string {
	var val []string
	for k, v := range m {
		val = append(val, fmt.Sprintf("%s:%s", k, toString(v)))
	}
	return strings.Join(val, " ")
}

// F the same as M but for filters
type F map[string]interface{}

// Q plain text query
type Q string

func (q Q) toQuery() string {
	return string(q)
}

// And is a query for AND statement 'test1:value1 AND test2:value2'
type And []Querier

func (q And) toQuery() string {
	var val []string
	for i := range q {
		val = append(val, q[i].toQuery())
	}
	return strings.Join(val, " AND ")
}

// Or is a query for OR statement 'test1:value1 OR test2:value2'
type Or []Querier

func (q Or) toQuery() string {
	var val []string
	for i := range q {
		val = append(val, q[i].toQuery())
	}
	return strings.Join(val, " OR ")
}

func toString(v interface{}) string {
	switch v.(type) {
	case string:
		return v.(string)
	case int:
		return strconv.Itoa(v.(int))
	case int32:
		return fmt.Sprintf("%d", v.(int32))
	case int64:
		return fmt.Sprintf("%d", v.(int64))
	case float32:
		return fmt.Sprintf("%f", v.(float32))
	case float64:
		return fmt.Sprintf("%f", v.(float64))
	case bool:
		return strconv.FormatBool(v.(bool))
	}
	return ""
}

func (q QueryBuilder) Query(query Querier) QueryBuilder {
	qstr := query.toQuery()
	return QueryBuilder{
		Q: &qstr,
	}
}

func (q QueryBuilder) Filter(filters F) QueryBuilder {
	for k, v := range filters {
		q.Fq = append(q.Fq, k+":"+toString(v))
	}
	return q
}

func (q QueryBuilder) Limit(limit int) QueryBuilder {
	q.L = &limit
	return q
}

func (q QueryBuilder) Offset(offset int) QueryBuilder {
	q.O = &offset
	return q
}

func (q QueryBuilder) Field(field ...string) QueryBuilder {
	q.Fl = append(q.Fl, field...)
	return q
}
