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
		Dir:   ".",
		Mode:  packages.NeedTypes | packages.NeedName | packages.NeedTypesInfo,
		Tests: true,
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
		return generate(pkg.Types, obj.Type())
	}
	return fmt.Errorf("bad type: %s", typename)
}

func generate(pkg *types.Package, typ types.Type) error {
	ntype, ok := typ.(*types.Named)
	if !ok {
		return fmt.Errorf("could not get name of type '%s'", typ)
	}
	st, ok := ntype.Underlying().(*types.Struct)
	if !ok {
		return errors.New("type is not a struct")
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
		tsig := sel.Obj().Type().(*types.Signature)
		mname := sel.Obj().Name()
		obj, idx, _ := types.LookupFieldOrMethod(typ, true, pkg, mname)
		var callorig string
		if obj != nil {
			callorig = fmt.Sprintf(
				" else {\n\t\t_fn = _recv.%s.%s\n\t}",
				st.Field(idx[0]).Name(),
				mname,
			)
		}
		out.WriteString(
			fmt.Sprintf(
				fn,
				tname,
				mname,
				sig(sel),
				args(tsig.Params(), tsig.Variadic()),
				callorig,
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
	names := paramnames(sig.Params(), false)
	for i := range sig.Params().Len() {
		p := sig.Params().At(i)
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(names[i])
		b.WriteString(" ")
		if i == sig.Params().Len()-1 && sig.Variadic() {
			b.WriteString("...")
			b.WriteString(types.TypeString(p.Type(), qualifier)[2:])
		} else {
			b.WriteString(types.TypeString(p.Type(), qualifier))
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
			b.WriteString(types.TypeString(r.Type(), qualifier))
		}
		if sig.Results().Len() > 1 {
			b.WriteString(")")
		}
	}
	return b.String()
}

func args(tup *types.Tuple, variadic bool) string {
	var b strings.Builder
	names := paramnames(tup, false)
	for i := range tup.Len() {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(names[i])
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
	names := resultnames(tup)
	for i := range tup.Len() {
		v := tup.At(i)
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(names[i])
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
	names := resultnames(tup)
	for i := range tup.Len() {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(names[i])
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
	names := paramnames(params, true)
	for i := range params.Len() {
		if i > 0 {
			b.WriteString("\n")
		}
		param := params.At(i)
		b.WriteString(fmt.Sprintf("\t%s %s",
			names[i],
			types.TypeString(param.Type(), qualifier),
		))
	}
	b.WriteString("\n")
	return b.String()
}

func paramnames(tup *types.Tuple, export bool) []string {
	names := make([]string, 0, tup.Len())
	for i := range tup.Len() {
		v := tup.At(i)
		var name string
		if v.Name() == "" {
			name = fmt.Sprintf("P%d", i)
		} else {
			name = v.Name()
			if export {
				name = capitalize(name)
			}
		}
		for slices.Contains(names, name) {
			name = name + "_"
		}
		names = append(names, name)
	}
	return names
}

func resultnames(tup *types.Tuple) []string {
	names := make([]string, 0, tup.Len())
	for i := range tup.Len() {
		v := tup.At(i)
		var name string
		if v.Name() == "" {
			name = fmt.Sprintf("r%d", i)
		} else if v.Name() == "error" {
			// Fix a common type shadowing error.
			name = "err"
		} else {
			name = v.Name()
		}
		for slices.Contains(names, name) {
			name = name + "_"
		}
		names = append(names, name)
	}
	return names
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

func capitalize(s string) string {
	runes := []rune(s)
	return string(append([]rune{unicode.ToUpper(runes[0])}, runes[1:]...))
}
