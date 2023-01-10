package main

/*
   lib flux should be with commit f6a9675
   https://www.sqlpac.com/en/documents/influxdb-moving-from-influxql-language-to-flux-language.html
   https://www.sqlpac.com/en/documents/influxdb-v2-flux-language-quick-reference-guide-cheat-sheet.html
*/

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	promqlTranspiler "github.com/adubovikov/p2i/promql/transpiler"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/influxdb/v2"
	influxQ "github.com/influxdata/influxdb/v2/query/influxql"
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
	FLUX_DIALECT     DialectType = "flux"
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
	//cmd.Dialect = INFLUXQL_DIALECT
	cmd.Dialect = FLUX_DIALECT
	finalSql := "SELECT sum(sum) FROM (SELECT sum(last) FROM (SELECT *::tag, last(value) FROM go_gc_duration_seconds_count GROUP BY *) GROUP BY container) GROUP BY endpoint"
	mustHave := influxql.MustParseStatement(finalSql)

	if cmd.Dialect == INFLUXQL_DIALECT {

		expr, err := parser.ParseExpr(cmd.Cmd)
		if err != nil {
			fmt.Println("command parse fail:", err)
			return
		}

		t := promqlTranspiler.NewTranspiler(&cmd.Start, &cmd.End,
			promqlTranspiler.WithTimezone(cmd.Timezone),
			promqlTranspiler.WithEvaluation(&cmd.Evaluation),
			promqlTranspiler.WithStep(cmd.Step),
			promqlTranspiler.WithDataType(promqlTranspiler.DataType(cmd.DataType)),
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
	} else {

		//q, err := influxql.ParseQuery(finalSql)
		/*
			if err != nil {
				fmt.Println("error during parse")
				fmt.Println("Expression: ", err)
				return
			}
		*/

		mesSql := "SELECT pcpu FROM netdatatsdb.autogen.vpsmetrics WHERE (time > now() -1d and host =~ /^vps/ and host !~ /(de|uk|us)/)"

		now := time.Now()

		tFlux := influxQ.NewTranspilerWithConfig(dbrpMapper{}, influxQ.Config{
			Now:            now,
			FallbackToDBRP: true,
		})

		pkg, err := tFlux.Transpile(context.Background(), mesSql)
		if err != nil {
			fmt.Println("Expression error: ", err)
			return
		}

		fmt.Println("WAS: ", mesSql)
		fmt.Println("Query : ", ast.Format(pkg))

	}
}

type dbrpMapper struct{}

// FindBy returns the dbrp mapping for the specified ID.
func (d dbrpMapper) FindByID(ctx context.Context, orgID influxdb.ID, id influxdb.ID) (*influxdb.DBRPMappingV2, error) {
	return nil, errors.New("mapping not found")
}

// FindMany returns a list of dbrp mappings that match filter and the total count of matching dbrp mappings.
func (d dbrpMapper) FindMany(ctx context.Context, dbrp influxdb.DBRPMappingFilterV2, opts ...influxdb.FindOptions) ([]*influxdb.DBRPMappingV2, int, error) {
	return nil, 0, errors.New("mapping not found")
}

// Create creates a new dbrp mapping, if a different mapping exists an error is returned.
func (d dbrpMapper) Create(ctx context.Context, dbrp *influxdb.DBRPMappingV2) error {
	return errors.New("dbrpMapper does not support creating new mappings")
}

// Update a new dbrp mapping
func (d dbrpMapper) Update(ctx context.Context, dbrp *influxdb.DBRPMappingV2) error {
	return errors.New("dbrpMapper does not support updating mappings")
}

// Delete removes a dbrp mapping.
func (d dbrpMapper) Delete(ctx context.Context, orgID influxdb.ID, id influxdb.ID) error {
	return errors.New("dbrpMapper does not support deleting mappings")
}
