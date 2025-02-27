package modelgen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
)

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
	var buff strings.Builder
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

	return buff.String()
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

			if strings.Contains(typeSpec.Name.Name, "Model") || (len(declComment) > 0 && declComment[:len(ModelAnnotation)] == ModelAnnotation) {
				output := generateForStruct(dialect, fileAst.Name.String(), typeSpec.Name.Name, structType)
				if output == "" {
					continue
				}

				notEmpty = true
				_, err := fmt.Fprint(outputFile, output)
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

func Generate(dialect string, packagePath string) {
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
