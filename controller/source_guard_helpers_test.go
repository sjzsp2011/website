package controller

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"testing"
)

func readFuncSource(t *testing.T, file, funcName string) string {
	t.Helper()

	src, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("读取 %s 失败：%v", file, err)
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, src, parser.ParseComments)
	if err != nil {
		t.Fatalf("解析 %s 失败：%v", file, err)
	}
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != funcName {
			continue
		}
		return string(src[fset.Position(fn.Pos()).Offset:fset.Position(fn.End()).Offset])
	}
	t.Fatalf("%s 中找不到函数 %s", file, funcName)

	return ""
}

func mustContain(t *testing.T, src, pattern, why string) {
	t.Helper()
	if !regexp.MustCompile(pattern).MatchString(src) {
		t.Errorf("%s（期望命中 /%s/）", why, pattern)
	}
}

func mustNotContain(t *testing.T, src, pattern, why string) {
	t.Helper()
	if regexp.MustCompile(pattern).MatchString(src) {
		t.Errorf("%s（不应命中 /%s/）", why, pattern)
	}
}
