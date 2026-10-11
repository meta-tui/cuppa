package gosource

import (
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// The slice layout (ADR 0013) puts the runtime in a package of its own and
// each page in another. The page code the generator writes for the flat
// layout refers to the runtime by unexported names in the same package, so the
// slice layout is made from that code: the names the pages use become
// exported, the pages qualify them with the runtime's package name, and the
// names of a page lose the page's prefix, as the names inside a package do.
// The transform works on the syntax tree, not on the text, so a field of a
// user's input called like a runtime type is not touched.

// surface is what a runtime package declares: its package level names and the
// unexported methods and fields of its types.
type surface struct {
	tops    map[string]bool
	members map[string]bool
}

// surfaceOf reads the declarations of the runtime files.
func surfaceOf(files map[string]string) (surface, error) {
	s := surface{tops: map[string]bool{}, members: map[string]bool{}}
	fset := token.NewFileSet()
	for name, src := range files {
		f, err := parser.ParseFile(fset, name, src, parser.SkipObjectResolution)
		if err != nil {
			return s, err
		}
		for _, d := range f.Decls {
			switch d := d.(type) {
			case *ast.FuncDecl:
				if d.Recv == nil {
					s.tops[d.Name.Name] = true
				} else {
					s.members[d.Name.Name] = true
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch spec := spec.(type) {
					case *ast.TypeSpec:
						s.tops[spec.Name.Name] = true
					case *ast.ValueSpec:
						for _, n := range spec.Names {
							s.tops[n.Name] = true
						}
					}
				}
			}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.StructType:
				for _, field := range n.Fields.List {
					for _, name := range field.Names {
						s.members[name.Name] = true
					}
					if len(field.Names) == 0 {
						// An embedded type is a field of the type's name.
						if id, ok := field.Type.(*ast.Ident); ok {
							s.members[id.Name] = true
						}
					}
				}
			case *ast.InterfaceType:
				for _, m := range n.Methods.List {
					for _, name := range m.Names {
						s.members[name.Name] = true
					}
				}
			}
			return true
		})
	}
	return s, nil
}

// edit replaces a byte range of a source.
type edit struct {
	from, to int
	text     string
}

func applyEdits(src string, edits []edit) string {
	sort.Slice(edits, func(i, j int) bool { return edits[i].from > edits[j].from })
	b := []byte(src)
	for _, e := range edits {
		b = append(b[:e.from], append([]byte(e.text), b[e.to:]...)...)
	}
	return string(b)
}

func exportedName(name string) string {
	r, n := utf8.DecodeRuneInString(name)
	return string(unicode.ToUpper(r)) + name[n:]
}

func isUnexported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsLower(r)
}

// pageNamer strips a page's own prefix from the names it declares.
type pageNamer struct {
	ident, lower string
	// events are the names (without the page's prefix) of the events the page
	// raises; their types are called <Name>Event, so an event named Default or
	// View does not take the name of a function of the package.
	events map[string]bool
	// off keeps every name as it is (a component package declares its own).
	off bool
	// own are the fields and methods the file declares itself: a selector of
	// one of those is its own even when the runtime has a member of that name.
	own map[string]bool
}

func upperAt(s string, i int) bool {
	if i >= len(s) {
		return false
	}
	r, _ := utf8.DecodeRuneInString(s[i:])
	return unicode.IsUpper(r)
}

// local says what a name the page declares is called inside its package, and
// whether it is one of the page's names at all.
func (n pageNamer) local(name string) (string, bool) {
	if n.off {
		return "", false
	}
	switch {
	case name == n.ident:
		return "View", true
	case name == "Default"+n.ident+"Props":
		return "Default", true
	case name == n.ident+"Props":
		return "Props", true
	case strings.HasPrefix(name, "Default"+n.ident) && upperAt(name, len("Default"+n.ident)):
		return "Default" + name[len("Default"+n.ident):], true
	case strings.HasPrefix(name, "is"+n.ident) && upperAt(name, len("is"+n.ident)):
		return "is" + name[len("is"+n.ident):], true
	case strings.HasPrefix(name, n.ident) && n.events[name[len(n.ident):]]:
		return name[len(n.ident):] + "Event", true
	case strings.HasPrefix(name, n.ident) && upperAt(name, len(n.ident)):
		return name[len(n.ident):], true
	case n.lower != "" && strings.HasPrefix(name, n.lower) && upperAt(name, len(n.lower)):
		rest := name[len(n.lower):]
		r, size := utf8.DecodeRuneInString(rest)
		return string(unicode.ToLower(r)) + rest[size:], true
	}
	return "", false
}

// pageUse is what a page's code takes from the runtime.
type pageUse struct {
	tops, members map[string]bool
	// pkgs are the component packages whose exported names the code uses.
	pkgs map[string]bool
	// qualify maps an exported name a component package declares to that
	// package's name.
	qualify map[string]string
}

// rewritePage turns the source of a page of the flat layout into the source of
// its package: the runtime's names are qualified with runtime (the package
// name it is imported as), the page's own names lose the page's prefix, and
// pkg is the package clause. It reports the unexported runtime names it used,
// which the runtime must export.
func rewritePage(src, pkg, runtime string, sf surface, namer pageNamer, use *pageUse) (string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "page.go", src, 0)
	if err != nil {
		return "", err
	}
	imported := map[string]bool{}
	for _, imp := range f.Imports {
		if imp.Name != nil {
			imported[imp.Name.Name] = true
		} else {
			path := strings.Trim(imp.Path.Value, `"`)
			imported[path[strings.LastIndex(path, "/")+1:]] = true
		}
	}
	var edits []edit
	at := func(id *ast.Ident) (int, int) {
		from := fset.Position(id.Pos()).Offset
		return from, from + len(id.Name)
	}
	replace := func(id *ast.Ident, text string) {
		from, to := at(id)
		edits = append(edits, edit{from, to, text})
	}
	// declared are the names of the page's own top level declarations.
	declared := map[string]bool{}
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			if d.Recv == nil {
				declared[d.Name.Name] = true
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch spec := spec.(type) {
				case *ast.TypeSpec:
					declared[spec.Name.Name] = true
				case *ast.ValueSpec:
					for _, n := range spec.Names {
						declared[n.Name] = true
					}
				}
			}
		}
	}
	// roles records, for identifiers that are not plain uses, what they are.
	roles := map[*ast.Ident]string{}
	ast.Inspect(f, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.SelectorExpr:
			roles[n.Sel] = "selector"
		case *ast.KeyValueExpr:
			if id, ok := n.Key.(*ast.Ident); ok {
				roles[id] = "key"
			}
		case *ast.Field:
			for _, id := range n.Names {
				if roles[id] == "" {
					roles[id] = "field"
				}
			}
		case *ast.FuncDecl:
			if n.Recv != nil {
				roles[n.Name] = "method"
			}
		case *ast.InterfaceType:
			// The methods of an interface the page declares are named like the
			// methods that implement it.
			for _, m := range n.Methods.List {
				for _, id := range m.Names {
					roles[id] = "method"
				}
			}
		}
		return true
	})
	ast.Inspect(f, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok || id.Name == "_" {
			return true
		}
		name := id.Name
		switch roles[id] {
		case "selector":
			if sf.members[name] && isUnexported(name) && !namer.own[name] {
				use.members[name] = true
				replace(id, exportedName(name))
			}
			return true
		case "key":
			if sf.tops[name] && isUnexported(name) && !namer.own[name] {
				use.tops[name] = true
				replace(id, exportedName(name))
				return true
			}
			if sf.members[name] && isUnexported(name) && !namer.own[name] {
				use.members[name] = true
				replace(id, exportedName(name))
				return true
			}
			if local, ok := namer.local(name); ok {
				replace(id, local)
			}
			return true
		case "field":
			return true
		case "method":
			if local, ok := namer.local(name); ok {
				replace(id, local)
			}
			return true
		}
		// A name the page declares, or refers to as its own.
		if declared[name] {
			if local, ok := namer.local(name); ok {
				replace(id, local)
			}
			return true
		}
		if id.Obj != nil {
			// A local variable, parameter or label.
			return true
		}
		if imported[name] || token.Lookup(name).IsKeyword() || isPredeclared(name) {
			return true
		}
		if pkg, ok := use.qualify[name]; ok {
			use.pkgs[pkg] = true
			replace(id, pkg+"."+name)
			return true
		}
		if sf.tops[name] {
			if isUnexported(name) {
				use.tops[name] = true
				name = exportedName(name)
			}
			replace(id, runtime+"."+name)
		}
		return true
	})
	out := applyEdits(src, edits)
	// The package clause.
	out = strings.Replace(out, "package "+f.Name.Name, "package "+pkg, 1)
	return out, nil
}

// exportRuntime renames, in the runtime files, the names the pages use.
func exportRuntime(files map[string]string, names map[string]bool) map[string]string {
	out := make(map[string]string, len(files))
	for name, src := range files {
		var edits []edit
		var s scanner.Scanner
		fset := token.NewFileSet()
		file := fset.AddFile(name, fset.Base(), len(src))
		s.Init(file, []byte(src), nil, scanner.ScanComments)
		for {
			pos, tok, lit := s.Scan()
			if tok == token.EOF {
				break
			}
			if tok == token.IDENT && names[lit] {
				from := file.Offset(pos)
				edits = append(edits, edit{from, from + len(lit), exportedName(lit)})
			}
		}
		out[name] = applyEdits(src, edits)
	}
	return out
}

var predeclared = map[string]bool{
	"bool": true, "byte": true, "complex64": true, "complex128": true, "error": true, "float32": true, "float64": true,
	"int": true, "int8": true, "int16": true, "int32": true, "int64": true, "rune": true, "string": true,
	"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true, "uintptr": true, "any": true,
	"true": true, "false": true, "iota": true, "nil": true, "append": true, "cap": true, "clear": true, "close": true,
	"complex": true, "copy": true, "delete": true, "imag": true, "len": true, "make": true, "max": true, "min": true,
	"new": true, "panic": true, "print": true, "println": true, "real": true, "recover": true, "comparable": true,
}

func isPredeclared(name string) bool { return predeclared[name] }

// ownMembers are the fields and methods a source declares.
func ownMembers(src string) (map[string]bool, error) {
	sf, err := surfaceOf(map[string]string{"x.go": src})
	return sf.members, err
}
