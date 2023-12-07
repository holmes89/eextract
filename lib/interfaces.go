package lib

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strings"
)

func FindServices(path string) []Service {
	set := token.NewFileSet()
	pack, err := parser.ParseFile(set, path, nil, 0)
	if err != nil {
		log.Fatal("Failed to parse package:", err)
	}
	entities := make([]Service, 0)
	for _, d := range pack.Decls {
		if gen, isGen := d.(*ast.GenDecl); isGen {
			if gen.Tok == token.TYPE {
				for _, s := range gen.Specs {
					if t, isType := s.(*ast.TypeSpec); isType {
						e := populateService(t)
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

func populateMethod(m *ast.Field) *Method {

	for _, fun := range m.Names {
		fmt.Printf("%+v\n", fun)
	}
	t, isType := m.Type.(*ast.FuncType)
	if !isType {
		return nil
	}
	if t.Params != nil {
		for _, p := range t.Params.List {
			narray := []string{}
			for _, n := range p.Names {
				narray = append(narray, n.String())
			}
			name := strings.Join(narray, ".")
			fmt.Printf("%+v\n", populateField(p.Type, name))
		}
	}

	return nil
}

func populateService(s ast.Spec) *Service {
	t, isType := s.(*ast.TypeSpec)
	if !isType {
		return nil
	}
	st, isIFace := t.Type.(*ast.InterfaceType)
	if !isIFace {
		return nil
	}
	e := Service{
		Name: t.Name.String(),
	}
	if !strings.Contains(e.Name, "Server") || strings.Contains(e.Name, "Unsafe") {
		return nil
	}
	for _, p := range st.Methods.List {
		for _, n := range p.Names {
			if !token.IsExported(n.Name) {
				continue
			}
			meth := populateMethod(p)
			if meth == nil {
				continue
			}
			e.Methods = append(e.Methods, *meth)
		}
	}
	return &e
}

func processTypeExpr(e ast.Expr) string {
	switch tyExpr := e.(type) {
	case *ast.StarExpr:
		switch ex := tyExpr.X.(type) {
		case *ast.Ident:
			return fmt.Sprintf("%v", ex.Name)
		case *ast.SelectorExpr:
			return ex.Sel.Name
		default:
			return fmt.Sprintf("%v", ex)
		}
	case *ast.Ellipsis:
		return fmt.Sprintf("%v", tyExpr.Elt)
	case *ast.ArrayType:
		return fmt.Sprintf("%v", tyExpr.Elt)
	case *ast.SelectorExpr:
		return fmt.Sprintf("%v.%v", tyExpr.X.(*ast.Ident).Name, tyExpr.Sel.Name)
	default:
		return fmt.Sprintf("%v", tyExpr)
	}
}

func serviceTypeString(t ast.Expr) string {
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
		return entityTypeString(i.X)
	default:
		return "unknown"
	}
}

type Method struct {
	Name   string
	Params []struct {
		Name string
		Type string
	}
	ReturnVals  []string
	ReturnCount int
	ReturnError bool
}

func (m Method) String() string {
	sb := strings.Builder{}
	sb.WriteString(m.Name)
	sb.WriteRune('(')
	for i, p := range m.Params {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%v", p))
	}
	sb.WriteString(") ")
	if len(m.ReturnVals) > 1 {
		sb.WriteRune('(')
	}
	for i, r := range m.ReturnVals {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%v", r))
	}
	if len(m.ReturnVals) > 1 {
		sb.WriteRune(')')
	}
	return sb.String()
}

type Service struct {
	Name    string
	Methods []Method
}

func (s Service) String() string {
	sb := strings.Builder{}
	sb.WriteString("type ")
	sb.WriteString(s.Name)
	sb.WriteString(" interface")
	sb.WriteRune('{')
	for _, m := range s.Methods {
		sb.WriteRune('\n')
		sb.WriteRune('\t')
		sb.WriteString(m.String())
	}
	sb.WriteRune('\n')
	sb.WriteRune('}')
	return sb.String()
}
