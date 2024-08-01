package main

import (
	_ "embed"
	"errors"
	"fmt"
	"go/types"
	"os"
	"strings"
	"unicode"

	"golang.org/x/tools/go/packages"
	"lesiw.io/flag"
)

var (
	typename string
	pkgname  string
	flags    = flag.NewSet(os.Stderr, "moxie TYPE")
	printver = flags.Bool("V,version", "print version and exit")
	imports  = map[string]string{
		"runtime": "runtime",
		"sync":    "sync",
		"unsafe":  "unsafe",
	}

	//go:embed version.txt
	versionfile string
	version     = strings.TrimSuffix(versionfile, "\n")
)

func main() {
	var code int
	defer func() { os.Exit(code) }()

	if err := run(); err != nil {
		if err.Error() != "" {
			fmt.Fprintln(os.Stderr, err)
		}
		code = 1
	}
}

func run() error {
	if err := flags.Parse(os.Args[1:]...); err != nil {
		return errors.New("")
	}
	if *printver {
		fmt.Println(version)
		return nil
	}
	if len(flags.Args) < 1 {
		flags.PrintError("bad type: no type provided")
		return errors.New("")
	}
	typename = flags.Args[0]

	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedName | packages.NeedTypesInfo,
		Dir:  ".",
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return fmt.Errorf("failed to load packages: %w", err)
	}
	for _, pkg := range pkgs {
		obj := pkg.Types.Scope().Lookup(typename)
		if obj == nil {
			continue
		}
		pkgname = pkg.Name
		return generate(obj.Type())
	}
	return fmt.Errorf("bad type: %s", typename)
}

func generate(typ types.Type) error {
	ntype, ok := typ.(*types.Named)
	if !ok {
		return fmt.Errorf("could not get name of type '%s'", typ)
	}
	tname := ntype.Obj().Name()
	fname := "mock_" + snakecase(tname) + "_test.go"
	f, err := os.Create(fname)
	if err != nil {
		return fmt.Errorf("failed to open '%s' for writing: %w", fname, err)
	}
	defer f.Close()
	var out strings.Builder

	mset := types.NewMethodSet(typ)

	out.WriteString(fmt.Sprintf(headerstart, pkgname, tname))
	for i := range mset.Len() {
		sel := mset.At(i)
		if !sel.Obj().Exported() {
			continue
		}
		sig := types.SelectionString(sel, qualifier)
		for range 2 {
			_, sig, _ = strings.Cut(sig, " ") // Discard field.
		}
		mname, sig, _ := strings.Cut(sig, "(")
		sig = "(" + sig
		out.WriteString(
			fmt.Sprintf(
				funcinfo,
				tname,
				mname,
				sig,
			),
		)
	}
	out.WriteString(fmt.Sprintf(headerend, tname))

	for i := range mset.Len() {
		sel := mset.At(i)
		if !sel.Obj().Exported() {
			continue
		}
		sig := types.SelectionString(sel, qualifier)
		for range 2 {
			_, sig, _ = strings.Cut(sig, " ") // Discard field.
		}
		mname, _, _ := strings.Cut(sig, "(")
		tsig := sel.Obj().Type().(*types.Signature)
		out.WriteString(
			fmt.Sprintf(
				calltype,
				tname,
				mname,
				paramfields(tsig),
			),
		)
	}

	for i := range mset.Len() {
		sel := mset.At(i)
		if !sel.Obj().Exported() {
			continue
		}
		origin := types.TypeString(
			sel.Obj().(*types.Func).Origin().Type().(*types.Signature).
				Recv().Type(),
			func(p *types.Package) string { return "" },
		)
		origin = strings.TrimLeft(origin, "*")
		sig := types.SelectionString(sel, qualifier)
		for range 2 {
			_, sig, _ = strings.Cut(sig, " ") // Discard field.
		}
		mname, sig, _ := strings.Cut(sig, "(")
		sig = "(" + sig
		tsig := sel.Obj().Type().(*types.Signature)
		out.WriteString(
			fmt.Sprintf(
				fn,
				tname,
				mname,
				sig,
				args(tsig.Params(), tsig.Variadic()),
				origin,
				args(tsig.Params(), false),
				argtypes(tsig.Params(), tsig.Variadic()),
				resultparams(tsig.Results()),
				resulttypes(tsig.Results()),
				resultargs(tsig.Results()),
			),
		)
	}
	_, _ = f.WriteString(strings.Replace(out.String(), "import()",
		importblock(), 1))

	return nil
}

func snakecase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func args(tup *types.Tuple, variadic bool) string {
	var b strings.Builder
	for i := range tup.Len() {
		v := tup.At(i)
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(v.Name())
		if i == tup.Len()-1 && variadic {
			b.WriteString("...")
		}
	}
	return b.String()
}

func argtypes(tup *types.Tuple, variadic bool) string {
	var b strings.Builder
	for i := range tup.Len() {
		v := tup.At(i)
		if i > 0 {
			b.WriteString(", ")
		}
		s := types.TypeString(v.Type(), qualifier)
		if i == tup.Len()-1 && variadic {
			s = strings.TrimPrefix(s, "[]")
			s = "..." + s
		}
		b.WriteString(s)
	}
	return b.String()
}

func resultparams(tup *types.Tuple) string {
	var b strings.Builder
	for i := range tup.Len() {
		v := tup.At(i)
		if i > 0 {
			b.WriteString(", ")
		}
		if v.Name() == "" {
			b.WriteString(fmt.Sprintf("r%d", i))
		} else {
			b.WriteString(v.Name())
		}
		b.WriteString(" " + types.TypeString(v.Type(), qualifier))
	}
	return b.String()
}

func resulttypes(tup *types.Tuple) string {
	var b strings.Builder
	for i := range tup.Len() {
		v := tup.At(i)
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(types.TypeString(v.Type(), qualifier))
	}
	return b.String()
}

func resultargs(tup *types.Tuple) string {
	var b strings.Builder
	for i := range tup.Len() {
		v := tup.At(i)
		if i > 0 {
			b.WriteString(", ")
		}
		if v.Name() == "" {
			b.WriteString(fmt.Sprintf("r%d", i))
		} else {
			b.WriteString(v.Name())
		}
	}
	return b.String()
}

func qualifier(pkg *types.Package) string {
	if name, ok := imports[pkg.Path()]; ok {
		return name
	}
	name := pkg.Name()
	if name == pkgname {
		name = ""
	}
pickname:
	for _, v := range imports {
		if v == name {
			name = name + "_"
			goto pickname
		}
	}
	imports[pkg.Path()] = name
	return imports[pkg.Path()]
}

func importblock() string {
	var b strings.Builder
	b.WriteString("import (\n")
	for path, name := range imports {
		if name == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("\t%s \"%s\"\n", name, path))
	}
	b.WriteString(")")

	return b.String()
}

func paramfields(sig *types.Signature) string {
	var b strings.Builder
	params := sig.Params()
	for i := range params.Len() {
		if i > 0 {
			b.WriteString("\n")
		}
		param := params.At(i)
		b.WriteString(fmt.Sprintf("\t%s %s", param.Name(),
			types.TypeString(param.Type(), qualifier),
		))
	}
	return b.String()
}
