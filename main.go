package main

import (
	"cmp"
	_ "embed"
	"errors"
	"fmt"
	"go/types"
	"os"
	"slices"
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

	if err := run(os.Args[1:]...); err != nil {
		if err.Error() != "" {
			fmt.Fprintln(os.Stderr, err)
		}
		code = 1
	}
}

func run(args ...string) error {
	if err := flags.Parse(args...); err != nil {
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
		tsig := sel.Obj().Type().(*types.Signature)
		mname := sel.Obj().Name()
		out.WriteString(
			fmt.Sprintf(
				fn,
				tname,
				mname,
				sig(sel),
				args(tsig.Params(), tsig.Variadic()),
				origin,
				args(tsig.Params(), false),
				argtypes(tsig.Params(), tsig.Variadic()),
				resultparams(tsig.Results()),
				resulttypes(tsig.Results()),
				resultargs(tsig.Results()),
				ternary(tsig.Results().Len() > 0, "return ", ""),
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

func sig(sel *types.Selection) string {
	var b strings.Builder
	b.WriteString(sel.Obj().Name())
	sig := sel.Obj().Type().(*types.Signature)
	b.WriteString("(")
	for i := range sig.Params().Len() {
		p := sig.Params().At(i)
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(cmp.Or(p.Name(), fmt.Sprintf("p%d", i)))
		b.WriteString(" ")
		if i == sig.Params().Len()-1 && sig.Variadic() {
			b.WriteString("...")
			b.WriteString(p.Type().String()[2:])
		} else {
			b.WriteString(p.Type().String())
		}
	}
	b.WriteString(")")
	if sig.Results().Len() > 0 {
		b.WriteString(" ")
		if sig.Results().Len() > 1 {
			b.WriteString("(")
		}
		for i := range sig.Results().Len() {
			r := sig.Results().At(i)
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(r.Type().String())
		}
		if sig.Results().Len() > 1 {
			b.WriteString(")")
		}
	}
	return b.String()
}

func args(tup *types.Tuple, variadic bool) string {
	var b strings.Builder
	for i := range tup.Len() {
		v := tup.At(i)
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(cmp.Or(v.Name(), fmt.Sprintf("p%d", i)))
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
		} else if v.Name() == "error" {
			// Fix a common type shadowing error.
			b.WriteString("err")
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
		} else if v.Name() == "error" {
			// Fix a common type shadowing error.
			b.WriteString("err")
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
	for _, path := range keys(imports) {
		name := imports[path]
		if name == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("\t%s \"%s\"\n", name, path))
	}
	b.WriteString(")")

	return b.String()
}

func paramfields(sig *types.Signature) string {
	params := sig.Params()
	if params.Len() == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n")
	for i := range params.Len() {
		if i > 0 {
			b.WriteString("\n")
		}
		param := params.At(i)
		b.WriteString(fmt.Sprintf("\t%s %s",
			cmp.Or(param.Name(), fmt.Sprintf("p%d", i)),
			types.TypeString(param.Type(), qualifier),
		))
	}
	b.WriteString("\n")
	return b.String()
}

func ternary[T any](cond bool, t, f T) T {
	if cond {
		return t
	} else {
		return f
	}
}

func keys[M ~map[K]V, K cmp.Ordered, V any](m M) []K {
	r := make([]K, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	slices.Sort(r)
	return r
}
