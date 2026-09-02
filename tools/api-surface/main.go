// Command api-surface emits a deterministic inventory of exported package symbols.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type entry struct {
	kind string
	name string
	sig  string
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <package-dir>\n", filepath.Base(os.Args[0]))
		os.Exit(2)
	}

	dir, err := filepath.Abs(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "abs path: %v\n", err)
		os.Exit(1)
	}

	entries, err := inventory(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "inventory: %v\n", err)
		os.Exit(1)
	}

	for _, e := range entries {
		fmt.Printf("%s %s %s\n", e.kind, e.name, e.sig)
	}
}

func inventory(dir string) ([]entry, error) {
	fset := token.NewFileSet()
	var files []*ast.File

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		files = append(files, file)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no Go source files in %s", dir)
	}

	var out []entry
	for _, file := range files {
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				out = append(out, genEntries(fset, d)...)
			case *ast.FuncDecl:
				if d.Name == nil || !ast.IsExported(d.Name.Name) {
					continue
				}
				if d.Recv != nil {
					recv := receiverString(fset, d.Recv)
					if recv == "" || !ast.IsExported(strings.TrimPrefix(recv, "*")) {
						continue
					}
					out = append(out, entry{
						kind: "method",
						name: fmt.Sprintf("(%s) %s", recv, d.Name.Name),
						sig:  signatureString(fset, d.Type),
					})
					continue
				}
				out = append(out, entry{
					kind: "func",
					name: d.Name.Name,
					sig:  signatureString(fset, d.Type),
				})
			}
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].kind != out[j].kind {
			return out[i].kind < out[j].kind
		}
		if out[i].name != out[j].name {
			return out[i].name < out[j].name
		}
		return out[i].sig < out[j].sig
	})

	return out, nil
}

func genEntries(fset *token.FileSet, decl *ast.GenDecl) []entry {
	var out []entry
	for _, spec := range decl.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			if s.Name == nil || !ast.IsExported(s.Name.Name) {
				continue
			}
			out = append(out, entry{
				kind: "type",
				name: s.Name.Name,
				sig:  typeString(fset, s.Type),
			})
		case *ast.ValueSpec:
			if s.Names == nil {
				continue
			}
			typeSig := ""
			if s.Type != nil {
				typeSig = typeString(fset, s.Type)
			}
			for _, name := range s.Names {
				if !ast.IsExported(name.Name) {
					continue
				}
				kind := decl.Tok.String()
				out = append(out, entry{
					kind: kind,
					name: name.Name,
					sig:  typeSig,
				})
			}
		}
	}
	return out
}

func typeString(fset *token.FileSet, expr ast.Expr) string {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, expr); err != nil {
		return "<format error>"
	}
	return strings.TrimSpace(buf.String())
}

func signatureString(fset *token.FileSet, ft *ast.FuncType) string {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, ft); err != nil {
		return "<format error>"
	}
	return strings.TrimSpace(buf.String())
}

func receiverString(fset *token.FileSet, recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}
	field := recv.List[0]
	var name string
	switch t := field.Type.(type) {
	case *ast.Ident:
		name = t.Name
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			name = "*" + ident.Name
		}
	default:
		name = typeString(fset, field.Type)
	}
	return name
}
