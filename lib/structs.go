package lib

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strings"

	"github.com/iancoleman/strcase"
)

func FindEntities(path string) []Entity {
	set := token.NewFileSet()
	pack, err := parser.ParseFile(set, path, nil, 0)
	if err != nil {
		log.Fatal("Failed to parse package:", err)
	}
	entities := make([]Entity, 0)
	for _, d := range pack.Decls {
		if gen, isGen := d.(*ast.GenDecl); isGen {
			if gen.Tok == token.TYPE {
				for _, s := range gen.Specs {
					if t, isType := s.(*ast.TypeSpec); isType {
						e := populateEntity(t)
						if e != nil {
							entities = append(entities, *e)
						}
					}
				}
			}
		}

	}
	return entities
}

func populateField(t ast.Expr, name string) Field {
	f := Field{
		Name: name,
		Type: entityTypeString(t),
	}
	switch t.(type) {
	case *ast.ArrayType:
		f.Slice = true
	case *ast.StarExpr:
		f.Star = true
	}
	return f
}

func populateEntity(s ast.Spec) *Entity {
	t, isType := s.(*ast.TypeSpec)
	if !isType {
		return nil
	}
	st, isStruct := t.Type.(*ast.StructType)
	if !isStruct {
		return nil
	}
	e := Entity{
		Name: t.Name.String(),
	}
	for _, p := range st.Fields.List {
		for _, n := range p.Names {
			if !token.IsExported(n.Name) {
				continue
			}
			e.Fields = append(e.Fields, populateField(p.Type, n.Name))
		}
	}
	return &e
}

func entityTypeString(t ast.Expr) string {
	switch t.(type) {
	case *ast.Ident:
		i := t.(*ast.Ident)
		if i.Name == "timestamppb" {
			return "time.Time"
		}
		return i.Name
	case *ast.ArrayType:
		i := t.(*ast.ArrayType)
		return entityTypeString(i.Elt)
	case *ast.StarExpr:
		i := t.(*ast.StarExpr)
		return entityTypeString(i.X)
	case *ast.SelectorExpr:
		i := t.(*ast.SelectorExpr)
		return fmt.Sprintf("%s.%s", entityTypeString(i.X), i.Sel.Name)
	default:
		return "unknown"
	}
}

type Field struct {
	Name  string
	Type  string
	Star  bool
	Slice bool
}

func (f Field) JSONField() string {
	return strcase.ToSnake(f.Name)
}

func (f Field) String() string {
	sb := strings.Builder{}
	sb.WriteString(f.Name)
	sb.WriteRune(' ')
	if f.Slice {
		sb.WriteString("[]")
	}
	if f.Star {
		sb.WriteRune('*')
	}
	sb.WriteString(f.Type)
	return sb.String()
}

type Entity struct {
	Name   string
	Fields []Field
}

func (e Entity) String() string {
	sb := strings.Builder{}
	sb.WriteString("type ")
	sb.WriteString(e.Name)
	sb.WriteString(" struct")
	sb.WriteRune('{')
	for _, f := range e.Fields {
		sb.WriteRune('\n')
		sb.WriteRune('\t')
		sb.WriteString(f.String())
	}
	sb.WriteRune('\n')
	sb.WriteRune('}')
	return sb.String()
}
