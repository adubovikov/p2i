package main

import (
	"fmt"
	"reflect"
	"time"

	"github.com/adubovikov/p2i/transpiler"
	"github.com/influxdata/influxql"
	"github.com/prometheus/prometheus/promql/parser"
)

const LineBreak = "\n"

type OperationType int

const (
	QUERY_OPERATION OperationType = iota + 1
	WRITE_OPERATION               = iota + 1
)

type DialectType string

type CommandType struct {
	OperationType OperationType
	DialectType   DialectType
}

type DataType int

const (
	TABLE_DATA DataType = iota + 1
	GRAPH_DATA
)

const (
	INFLUXQL_DIALECT DialectType = "influxql"
)

type Command struct {
	Cmd     string
	Dialect DialectType

	Database string
	// Start and End attributes are used for PromQL as it doesn't support time range itself
	Start      time.Time
	End        time.Time
	Timezone   *time.Location
	Evaluation time.Time
	Step       time.Duration

	DataType DataType
}

func main() {

	cmd := Command{}

	cmd.Cmd = "sum by (endpoint) (sum(go_gc_duration_seconds_count) by (container))"
	cmd.Start = time.Date(2023, 1, 6, 12, 0, 0, 0, time.Local)
	cmd.End = time.Date(2023, 1, 8, 10, 0, 0, 0, time.Local)
	cmd.Timezone = cmd.Start.Location()
	cmd.Evaluation = time.Date(2023, 1, 6, 15, 0, 0, 0, time.Local)
	//cmd.Step = time.Nanosecond
	cmd.DataType = TABLE_DATA

	mustHave := influxql.MustParseStatement(`SELECT sum(sum) FROM (SELECT sum(last) FROM (SELECT *::tag, last(value) FROM go_gc_duration_seconds_count GROUP BY *) GROUP BY container) GROUP BY endpoint`)

	expr, err := parser.ParseExpr(cmd.Cmd)
	if err != nil {
		fmt.Println("command parse fail:", err)
		return
	}

	t := transpiler.NewTranspiler(&cmd.Start, &cmd.End,
		transpiler.WithTimezone(cmd.Timezone),
		transpiler.WithEvaluation(&cmd.Evaluation),
		transpiler.WithStep(cmd.Step),
		transpiler.WithDataType(transpiler.DataType(cmd.DataType)),
	)
	sql, err := t.Transpile(expr)
	if err != nil {
		fmt.Println("command execute fail:", err)
		return
	}

	if !reflect.DeepEqual(sql, mustHave.String()) {
		fmt.Println("they are not equal!")
		fmt.Println("Expression: ", cmd.Cmd)
		fmt.Println("Generated:- [" + sql + "]")
		fmt.Println("Expected: - [" + mustHave.String() + "]")
	} else {
		fmt.Println("they are equal!")
		fmt.Println("Expression: ", cmd.Cmd)
		fmt.Println("Generated:- [" + sql + "]")
		fmt.Println("Expected: - [" + mustHave.String() + "]")
	}
}
