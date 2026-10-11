package gosource

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// hoistMethods moves out of a component file the methods it declares on types
// of the runtime (a method cannot be declared in another package): it returns
// the file without them, and them as a runtime file together with the imports
// they use. An empty hoisted means there were none.
func hoistMethods(src string, runtimeTypes map[string]bool) (rest, hoisted string, err error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "c.go", src, parser.ParseComments)
	if err != nil {
		return "", "", err
	}
	var cuts []edit
	var moved []string
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
			continue
		}
		t := fn.Recv.List[0].Type
		if star, ok := t.(*ast.StarExpr); ok {
			t = star.X
		}
		id, ok := t.(*ast.Ident)
		if !ok || !runtimeTypes[id.Name] {
			continue
		}
		from := fset.Position(d.Pos())
		if fn.Doc != nil {
			from = fset.Position(fn.Doc.Pos())
		}
		to := fset.Position(d.End())
		cuts = append(cuts, edit{from.Offset, to.Offset, ""})
		moved = append(moved, src[from.Offset:to.Offset])
	}
	if len(moved) == 0 {
		return src, "", nil
	}
	body := strings.Join(moved, "\n\n")
	var imports []string
	for _, imp := range f.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		name := path[strings.LastIndex(path, "/")+1:]
		line := imp.Path.Value
		if imp.Name != nil {
			name = imp.Name.Name
			line = name + " " + imp.Path.Value
		}
		if strings.Contains(body, name+".") {
			imports = append(imports, "\t"+line)
		}
	}
	head := GeneratedHeader + "\n\npackage cuppa\n\n"
	if len(imports) > 0 {
		head += "import (\n" + strings.Join(imports, "\n") + "\n)\n\n"
	}
	return applyEdits(src, cuts), head + body + "\n", nil
}
