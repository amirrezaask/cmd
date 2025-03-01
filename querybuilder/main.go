package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
)

func main() {
	var file string
	var dialect string
	flag.StringVar(&file, "file", "", "path to the file to generate the query builder for")
	flag.StringVar(&dialect, "dialect", "mysql", "dialect to generate the query builder for")
	flag.Parse()

	if file == "" {
		flag.Usage()
		return
	}

	generateForFile(dialect, file)
}

const ModelAnnotation = "@querybuilder"

type structField struct {
	Name         string
	Type         string
	IsComparable bool
	IsNullable   bool
	Tag          string
}

func (s structField) String() string {
	return s.Name
}

func isComparable(typeExpr ast.Expr) bool {
	switch t := typeExpr.(type) {
	case *ast.Ident:
		if t.Obj == nil {
			// it's a primitive go type
			if t.Name == "int" || t.Name == "int8" || t.Name == "int16" || t.Name == "int32" || t.Name == "int64" ||
				t.Name == "uint" || t.Name == "uint8" || t.Name == "uint16" || t.Name == "uint32" || t.Name == "uint64" ||
				t.Name == "float32" || t.Name == "float64" {
				return true
			}
			return false
		}
	}
	return false
}

func resolveTypes(structDecl *ast.StructType) []structField {
	var fields []structField
	for _, field := range structDecl.Fields.List {
		for _, name := range field.Names {
			sf := structField{
				Name:         name.Name,
				Type:         fmt.Sprint(field.Type),
				IsComparable: isComparable(field.Type),
				IsNullable:   false, // TODO: fix this
			}
			if field.Tag != nil {
				sf.Tag = field.Tag.Value
			}
			fields = append(fields, sf)
		}
	}
	return fields
}

func generateForStruct(dialect string, pkg string, name string, structDecl *ast.StructType) string {
	fields := resolveTypes(structDecl)
	var buff bytes.Buffer
	// if strings.Contains(strings.ToLower(name), "model") {
	// 	name = strings.Replace(strings.ToLower(name), "model", "", -1)
	// 	name = strcase.ToCamel(name)
	// }
	td := templateData{
		ModelName:                 name,
		QueryBuilderStructName:    fmt.Sprintf("_dont_use_%s_query_builder", strings.ToLower(name)),
		QueryBuilderInterfaceName: name + "QueryBuilder",
		Fields:                    fields,
		Pkg:                       pkg,
		Dialect:                   dialect,
		TableName:                 strcase.ToSnake(pluralize.NewClient().Plural(name)),
	}

	err := tmpl.Execute(&buff, td)
	if err != nil {
		panic(err)
	}

	out, err := format.Source(buff.Bytes())
	if err != nil {
		return buff.String()
	}

	return string(out)
}

func generateForFile(dialect string, filePath string) {
	inputFilePath, err := filepath.Abs(filePath)
	if err != nil {
		panic(err)
	}

	pathList := filepath.SplitList(inputFilePath)
	pathList = pathList[:len(pathList)-1]
	fileDir := filepath.Join(pathList...)
	fileSet := token.NewFileSet()
	fileAst, err := parser.ParseFile(fileSet, filePath, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	actualName := strings.TrimSuffix(filePath, filepath.Ext(filePath))
	outputFilePath := filepath.Join(fileDir, fmt.Sprintf("%s_model_gen.go", actualName))
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		panic(err)
	}
	defer func(outputFile *os.File) {
		err := outputFile.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(outputFile)

	var notEmpty bool

	for _, decl := range fileAst.Decls {
		if _, ok := decl.(*ast.GenDecl); ok {

			declComment := decl.(*ast.GenDecl).Doc.Text()
			// Ensure the GenDecl contains a TypeSpec
			if len(decl.(*ast.GenDecl).Specs) == 0 {
				continue
			}

			typeSpec, ok := decl.(*ast.GenDecl).Specs[0].(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			var codes []string

			if strings.Contains(typeSpec.Name.Name, "Model") ||
				(len(declComment) > 0 && declComment[:len(ModelAnnotation)] == ModelAnnotation) {

				output := generateForStruct(dialect, fileAst.Name.String(), typeSpec.Name.Name, structType)
				if output == "" {
					continue
				}
				codes = append(codes, output)
				notEmpty = true
			}

			if len(codes) > 0 {
				err = fileTemplate.Execute(outputFile, struct {
					Pkg  string
					Code string
				}{Code: strings.Join(codes, "\n\n")})
				if err != nil {
					panic(err)
				}
			}

		}
	}
	if !notEmpty {
		os.Remove(outputFilePath)
	}
}

func generate(dialect string, packagePath string) {
	err := filepath.Walk(packagePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") && !strings.Contains(info.Name(), "_gen") {
			generateForFile(dialect, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

var funcMap = template.FuncMap{
	"toSnakeCase": func(name string) string {
		return strcase.ToSnake(name)
	},
	"ToLowerCamelCase": func(name string) string {
		return strcase.ToLowerCamel(name)
	},
}

type templateData struct {
	Pkg                       string
	ModelName                 string
	QueryBuilderInterfaceName string
	QueryBuilderStructName    string
	TableName                 string
	Fields                    []structField
	Dialect                   string
}

var fileTemplate = template.Must(template.New("modelgenfile").Funcs(funcMap).Parse(`// Code generated by modelgen. DO NOT EDIT

package {{ .Pkg }}
import (
    "fmt"
    "strings"
    "database/sql"
)

{{ .Code }}
	`))

var tmpl = template.Must(template.New("modelgen").Funcs(funcMap).Parse(
	`
type {{.QueryBuilderInterfaceName}} interface{
	{{ range .Fields }}
	Where{{.Name}}Is({{.Type}}) {{$.QueryBuilderInterfaceName}}
	Where{{.Name}}(operator string, rhs {{.Type}}) {{$.QueryBuilderInterfaceName}}
	{{ if .IsComparable  }}
	// Where{{.Name}}GT({{ .Type }}) {{$.QueryBuilderInterfaceName}}
	// Where{{.Name}}GE({{ .Type }}) {{$.QueryBuilderInterfaceName}}
	// Where{{.Name}}LT({{ .Type }}) {{$.QueryBuilderInterfaceName}}
	// Where{{.Name}}LE({{ .Type }}) {{$.QueryBuilderInterfaceName}}
	{{ end }}
	{{ end }}

	OrderByAsc(column {{$.ModelName}}Column) {{$.QueryBuilderInterfaceName}}
	OrderByDesc(column {{$.ModelName}}Column) {{$.QueryBuilderInterfaceName}}

	Limit(int) {{$.QueryBuilderInterfaceName}}
	Offset(int) {{$.QueryBuilderInterfaceName}}

    getPlaceholder() string

	First(db *sql.DB) ({{ $.ModelName }}, error)
	Last(db *sql.DB) ({{ $.ModelName }}, error)

	{{ range .Fields }}
	Set{{.Name}}({{.Type}}) {{$.QueryBuilderInterfaceName}}
	{{end}}

	Update(db *sql.DB) (sql.Result, error)

	Delete(db *sql.DB) (sql.Result, error)

	Fetch(db *sql.DB) ([]{{ $.ModelName }}, error)
	FindAll(db *sql.DB) ([]{{ $.ModelName }}, error)

	SQL() (string, error)
}


type {{ .QueryBuilderStructName }} struct {
	mode string

    where struct {
	{{ range .Fields }}
		{{.Name}} struct {
      	  argument interface{}
          operator string
    	}
	{{ end }}
	}

	set struct {
	{{ range .Fields }}
		{{.Name }} string
    {{ end }}
	}

	orderBy []string
	groupBy string

	projected []string

	limit int
	offset int

	whereArgs []interface{}
    setArgs []interface{}
}


func {{.ModelName}}s() {{ .QueryBuilderInterfaceName }} {
	return &{{ .QueryBuilderStructName }}{}
}

func (q *{{.QueryBuilderStructName}}) SQL() (string, error) {
	if q.mode == "" { q.mode = "select" }
	if q.mode == "select" {
		return q.sqlSelect()
	} else if q.mode == "update" {
		return q.sqlUpdate()
	} else if q.mode == "delete" {
		return q.sqlDelete()
	} else {
		return "", fmt.Errorf("unsupported query mode '%s'", q.mode)
	}
}


type {{.ModelName}}Column string

var {{.ModelName}}Columns = struct {
	{{ range .Fields }} {{.Name}} {{$.ModelName}}Column
	{{ end }}
}{
	{{ range .Fields }} {{.Name}}: {{$.ModelName}}Column("{{ toSnakeCase .Name }}"),
	{{ end }}
}


{{ if eq .Dialect "mysql" }}func (q *{{.QueryBuilderStructName}}) getPlaceholder() string {
	return "?"
}
{{ end }}
{{ if eq .Dialect "sqlite" }}func (q *{{.QueryBuilderStructName}}) getPlaceholder() string {
	return "?"
}
{{ end }}
{{ if eq .Dialect "postgres" }}func (q *{{.QueryBuilderStructName}}) getPlaceholder() string {
	return fmt.Sprintf("$", len(q.whereArgs) + len(q.setArgs) + 1)
}
{{ end }}


func (q *{{.QueryBuilderStructName}}) Limit(l int) {{ .QueryBuilderInterfaceName }} {
	q.mode = "select"
	q.limit = l
	return q
}

func (q *{{.QueryBuilderStructName}}) Offset(l int) {{ .QueryBuilderInterfaceName }} {
	q.mode = "select"
	q.offset = l
	return q
}


func (q {{ .ModelName }}) Values() []interface{} {
    var values []interface{}
	{{ range .Fields }}values = append(values, &q.{{ .Name }})
	{{ end }}
    return values
}


func {{ .ModelName }}sFromRows(rows *sql.Rows) ([]{{.ModelName}}, error) {
    var {{.ModelName}}s []{{.ModelName}}
    for rows.Next() {
        var m {{ .ModelName }}
        err := rows.Scan(
            {{ range .Fields }}
            &m.{{ .Name }},
            {{ end }}
        )
        if err != nil {
            return nil, err
        }
        {{.ModelName}}s = append({{.ModelName}}s, m)
    }
    return {{.ModelName}}s, nil
}

func {{ .ModelName }}FromRow(row *sql.Row) ({{.ModelName}}, error) {
    if row.Err() != nil {
        return {{.ModelName}}{}, row.Err()
    }
    var q {{ .ModelName }}
    err := row.Scan(
        {{ range .Fields }}&q.{{ .Name }},
        {{ end }}
    )
    if err != nil {
        return {{.ModelName}}{}, err
    }

    return q, nil
}

func (q *{{.QueryBuilderStructName}}) Update(db *sql.DB) (sql.Result, error) {
	q.mode = "update"
	args := append(q.setArgs, q.whereArgs...)
	query, err := q.SQL()
	if err != nil {
		return nil, err
	}
	return db.Exec(query, args...)
}

func (q *{{.QueryBuilderStructName}}) Delete(db *sql.DB) (sql.Result, error) {
	q.mode = "delete"
	query, err := q.SQL()
	if err != nil {
		return nil, err
	}
	return db.Exec(query, q.whereArgs...)
}

func (q *{{.QueryBuilderStructName}}) Fetch(db *sql.DB) ([]{{ .ModelName }}, error) {
	q.mode = "select"
	query, err := q.SQL()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(query, q.whereArgs...)
	if err != nil {
		return nil, err
	}
	return {{ .ModelName }}sFromRows(rows)
}

func (q *{{.QueryBuilderStructName}}) FindAll(db *sql.DB) ([]{{ .ModelName }}, error) {
	return q.Fetch(db)
}

func (q *{{.QueryBuilderStructName}}) First(db *sql.DB) ({{ .ModelName }}, error) {
	q.mode = "select"
	q.orderBy = []string{"ORDER BY id ASC"}
	q.Limit(1)
	query, err := q.SQL()
	if err != nil {
		return {{ .ModelName }}{}, err
	}
	row := db.QueryRow(query, q.whereArgs...)
	if row.Err() != nil {
		return {{ .ModelName }}{}, row.Err()
	}
	return {{ .ModelName}}FromRow(row)
}


func (q *{{.QueryBuilderStructName}}) Last(db *sql.DB) ({{ .ModelName }}, error) {
	q.mode = "select"
	q.orderBy = []string{"ORDER BY id DESC"}
	q.Limit(1)
	query, err := q.SQL()
	if err != nil {
		return {{ .ModelName }}{}, err
	}
	row := db.QueryRow(query, q.whereArgs...)
	if row.Err() != nil {
		return {{ .ModelName}}{}, row.Err()
	}
	return {{ .ModelName }}FromRow(row)
}

func (q *{{ $.QueryBuilderStructName }}) OrderByAsc(column {{.ModelName}}Column) {{ $.QueryBuilderInterfaceName }} {
    q.mode = "select"
	q.orderBy = append(q.orderBy, fmt.Sprintf("%s ASC", string(column)))
	return q
}

func (q *{{ $.QueryBuilderStructName }}) OrderByDesc(column {{.ModelName}}Column) {{ $.QueryBuilderInterfaceName }} {
    q.mode = "select"
	q.orderBy = append(q.orderBy, fmt.Sprintf("%s DESC", string(column)))
	return q
}

func (q *{{ .QueryBuilderStructName }}) sqlSelect() (string, error) {
	if q.projected == nil {
		q.projected = append(q.projected, "*")
	}
	base := fmt.Sprintf("SELECT %s FROM {{ .TableName }}", strings.Join(q.projected, ", "))

	var wheres []string
	{{ range .Fields }}
	if q.where.{{.Name}}.operator != "" {
		wheres = append(wheres, fmt.Sprintf("%s %s %s", "{{ toSnakeCase .Name }}", q.where.{{ .Name }}.operator, fmt.Sprint(q.where.{{ .Name }}.argument)))
	}
	{{ end }}
	if len(wheres) > 0 {
		base += " WHERE " + strings.Join(wheres, " AND ")
	}

	if len(q.orderBy) > 0 {
		base += fmt.Sprintf(" ORDER BY %s", strings.Join(q.orderBy, ", "))
	}

	if q.limit != 0 {
		base += " LIMIT " + fmt.Sprint(q.limit)
	}
	if q.offset != 0 {
		base += " OFFSET " + fmt.Sprint(q.offset)
	}
	return base, nil
}


func (q *{{ .QueryBuilderStructName }}) sqlUpdate() (string, error) {
	base := fmt.Sprintf("UPDATE {{.TableName}} ")

	var wheres []string
    var sets []string

    {{ range .Fields }}
	if q.where.{{.Name}}.operator != "" {
		wheres = append(wheres, fmt.Sprintf("%s %s %s", "{{ toSnakeCase .Name }}", q.where.{{ .Name }}.operator, fmt.Sprint(q.where.{{ .Name }}.argument)))
	}
	if q.set.{{ .Name }} != "" {
		sets = append(sets, fmt.Sprintf("%s = %s", "{{ toSnakeCase .Name }}", fmt.Sprint(q.set.{{ .Name }})))
	}
    {{ end }}

	if len(sets) > 0 {
		base += "SET " + strings.Join(sets, " , ")
	}

	if len(wheres) > 0 {
		base += " WHERE " + strings.Join(wheres, " AND ")
	}



	return base, nil
}

func (q *{{ .QueryBuilderStructName }}) sqlDelete() (string, error) {
    base := fmt.Sprintf("DELETE FROM {{ .TableName }}")

	var wheres []string
	{{ range .Fields }}
	if q.where.{{.Name}}.operator != "" {
		wheres = append(wheres, fmt.Sprintf("%s %s %s", "{{ toSnakeCase .Name  }}", q.where.{{.Name }}.operator, fmt.Sprint(q.where.{{.Name }}.argument)))
	}
	{{ end }}
	if len(wheres) > 0 {
		base += " WHERE " + strings.Join(wheres, " AND ")
	}

	return base, nil
}

{{ range .Fields }}
{{ if .IsComparable  }}
func (q *{{ $.QueryBuilderStructName}}) Where{{.Name}}GE({{.Name }} {{.Type}}) {{$.QueryBuilderInterfaceName}} {
    q.whereArgs = append(q.whereArgs, {{.Name }})
    q.where.{{.Name }}.argument = q.getPlaceholder()
    q.where.{{.Name }}.operator = ">="
	return q
}

func (q *{{ $.QueryBuilderStructName}}) Where{{.Name }}GT({{.Name }} {{.Type}}) {{$.QueryBuilderInterfaceName}} {
    q.whereArgs = append(q.whereArgs, {{.Name }})
    q.where.{{.Name }}.argument = q.getPlaceholder()
    q.where.{{.Name }}.operator = ">"
	return q
}

func (q *{{ $.QueryBuilderStructName}}) Where{{.Name }}LE({{.Name }} {{.Type}}) {{$.QueryBuilderInterfaceName}} {
    q.whereArgs = append(q.whereArgs, {{.Name }})
    q.where.{{.Name }}.argument = q.getPlaceholder()
    q.where.{{.Name }}.operator = "<="
	return q
}

func (q *{{ $.QueryBuilderStructName}}) Where{{.Name }}LT({{.Name }} {{.Type}}) {{$.QueryBuilderInterfaceName}} {
    q.whereArgs = append(q.whereArgs, {{.Name }})
    q.where.{{.Name }}.argument = q.getPlaceholder()
    q.where.{{.Name }}.operator = "<"
	return q
}


{{ end }}
{{ end }}


{{ range .Fields }}
func (q *{{ $.QueryBuilderStructName}}) Where{{.Name }}(operator string, {{.Name }} {{.Type}}) {{$.QueryBuilderInterfaceName}} {
    q.whereArgs = append(q.whereArgs, {{.Name }})
    q.where.{{.Name }}.argument = q.getPlaceholder()
    q.where.{{.Name }}.operator = operator
	return q
}

func (q *{{ $.QueryBuilderStructName }}) Where{{.Name}}Is({{ .Name }} {{ .Type }}) {{ $.QueryBuilderInterfaceName }} {
    q.whereArgs = append(q.whereArgs, {{.Name}})
    q.where.{{.Name}}.argument = q.getPlaceholder()
    q.where.{{.Name}}.operator = "="
	return q
}
{{ end }}

{{ range .Fields }}
func (q *{{ $.QueryBuilderStructName }}) Set{{ .Name }}({{ .Name }} {{ .Type }}) {{ $.QueryBuilderInterfaceName }} {
	q.mode = "update"
    q.setArgs = append(q.setArgs, {{ .Name }})
	q.set.{{.Name}} = q.getPlaceholder()
	return q
}
{{ end }}
`,
))
