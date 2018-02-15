// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE ***REMOVED***le.

//go:generate go run mkstdlib.go

// Package imports implements a Go pretty-printer (like package "go/format")
// that also adds or removes import statements as necessary.
package imports // import "golang.org/x/tools/imports"

import (
	"bu***REMOVED***o"
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

// Options speci***REMOVED***es options for processing ***REMOVED***les.
type Options struct {
	Fragment  bool // Accept fragment of a source ***REMOVED***le (no package statement)
	AllErrors bool // Report all errors (not just the ***REMOVED***rst 10 on different lines)

	Comments  bool // Print comments (true if nil *Options provided)
	TabIndent bool // Use tabs for indent (true if nil *Options provided)
	TabWidth  int  // Tab width (8 if nil *Options provided)

	FormatOnly bool // Disable the insertion and deletion of imports
}

// Process formats and adjusts imports for the provided ***REMOVED***le.
// If opt is nil the defaults are used.
//
// Note that ***REMOVED***lename's directory influences which imports can be chosen,
// so it is important that ***REMOVED***lename be accurate.
// To process data ``as if'' it were in ***REMOVED***lename, pass the data as a non-nil src.
func Process(***REMOVED***lename string, src []byte, opt *Options) ([]byte, error) {
	if opt == nil {
		opt = &Options{Comments: true, TabIndent: true, TabWidth: 8}
	}

	***REMOVED***leSet := token.NewFileSet()
	***REMOVED***le, adjust, err := parse(***REMOVED***leSet, ***REMOVED***lename, src, opt)
	if err != nil {
		return nil, err
	}

	if !opt.FormatOnly {
		_, err = ***REMOVED***xImports(***REMOVED***leSet, ***REMOVED***le, ***REMOVED***lename)
		if err != nil {
			return nil, err
		}
	}

	sortImports(***REMOVED***leSet, ***REMOVED***le)
	imps := astutil.Imports(***REMOVED***leSet, ***REMOVED***le)
	var spacesBefore []string // import paths we need spaces before
	for _, impSection := range imps {
		// Within each block of contiguous imports, see if any
		// import lines are in different group numbers. If so,
		// we'll need to put a space between them so it's
		// compatible with gofmt.
		lastGroup := -1
		for _, importSpec := range impSection {
			importPath, _ := strconv.Unquote(importSpec.Path.Value)
			groupNum := importGroup(importPath)
			if groupNum != lastGroup && lastGroup != -1 {
				spacesBefore = append(spacesBefore, importPath)
			}
			lastGroup = groupNum
		}

	}

	printerMode := printer.UseSpaces
	if opt.TabIndent {
		printerMode |= printer.TabIndent
	}
	printCon***REMOVED***g := &printer.Con***REMOVED***g{Mode: printerMode, Tabwidth: opt.TabWidth}

	var buf bytes.Buffer
	err = printCon***REMOVED***g.Fprint(&buf, ***REMOVED***leSet, ***REMOVED***le)
	if err != nil {
		return nil, err
	}
	out := buf.Bytes()
	if adjust != nil {
		out = adjust(src, out)
	}
	if len(spacesBefore) > 0 {
		out, err = addImportSpaces(bytes.NewReader(out), spacesBefore)
		if err != nil {
			return nil, err
		}
	}

	out, err = format.Source(out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// parse parses src, which was read from ***REMOVED***lename,
// as a Go source ***REMOVED***le or statement list.
func parse(fset *token.FileSet, ***REMOVED***lename string, src []byte, opt *Options) (*ast.File, func(orig, src []byte) []byte, error) {
	parserMode := parser.Mode(0)
	if opt.Comments {
		parserMode |= parser.ParseComments
	}
	if opt.AllErrors {
		parserMode |= parser.AllErrors
	}

	// Try as whole source ***REMOVED***le.
	***REMOVED***le, err := parser.ParseFile(fset, ***REMOVED***lename, src, parserMode)
	if err == nil {
		return ***REMOVED***le, nil, nil
	}
	// If the error is that the source ***REMOVED***le didn't begin with a
	// package line and we accept fragmented input, fall through to
	// try as a source fragment.  Stop and return on any other error.
	if !opt.Fragment || !strings.Contains(err.Error(), "expected 'package'") {
		return nil, nil, err
	}

	// If this is a declaration list, make it a source ***REMOVED***le
	// by inserting a package clause.
	// Insert using a ;, not a newline, so that parse errors are on
	// the correct line.
	const pre***REMOVED***x = "package main;"
	psrc := append([]byte(pre***REMOVED***x), src...)
	***REMOVED***le, err = parser.ParseFile(fset, ***REMOVED***lename, psrc, parserMode)
	if err == nil {
		// Gofmt will turn the ; into a \n.
		// Do that ourselves now and update the ***REMOVED***le contents,
		// so that positions and line numbers are correct going forward.
		psrc[len(pre***REMOVED***x)-1] = '\n'
		fset.File(***REMOVED***le.Package).SetLinesForContent(psrc)

		// If a main function exists, we will assume this is a main
		// package and leave the ***REMOVED***le.
		if containsMainFunc(***REMOVED***le) {
			return ***REMOVED***le, nil, nil
		}

		adjust := func(orig, src []byte) []byte {
			// Remove the package clause.
			src = src[len(pre***REMOVED***x):]
			return matchSpace(orig, src)
		}
		return ***REMOVED***le, adjust, nil
	}
	// If the error is that the source ***REMOVED***le didn't begin with a
	// declaration, fall through to try as a statement list.
	// Stop and return on any other error.
	if !strings.Contains(err.Error(), "expected declaration") {
		return nil, nil, err
	}

	// If this is a statement list, make it a source ***REMOVED***le
	// by inserting a package clause and turning the list
	// into a function body.  This handles expressions too.
	// Insert using a ;, not a newline, so that the line numbers
	// in fsrc match the ones in src.
	fsrc := append(append([]byte("package p; func _() {"), src...), '}')
	***REMOVED***le, err = parser.ParseFile(fset, ***REMOVED***lename, fsrc, parserMode)
	if err == nil {
		adjust := func(orig, src []byte) []byte {
			// Remove the wrapping.
			// Gofmt has turned the ; into a \n\n.
			src = src[len("package p\n\nfunc _() {"):]
			src = src[:len(src)-len("}\n")]
			// Gofmt has also indented the function body one level.
			// Remove that indent.
			src = bytes.Replace(src, []byte("\n\t"), []byte("\n"), -1)
			return matchSpace(orig, src)
		}
		return ***REMOVED***le, adjust, nil
	}

	// Failed, and out of options.
	return nil, nil, err
}

// containsMainFunc checks if a ***REMOVED***le contains a function declaration with the
// function signature 'func main()'
func containsMainFunc(***REMOVED***le *ast.File) bool {
	for _, decl := range ***REMOVED***le.Decls {
		if f, ok := decl.(*ast.FuncDecl); ok {
			if f.Name.Name != "main" {
				continue
			}

			if len(f.Type.Params.List) != 0 {
				continue
			}

			if f.Type.Results != nil && len(f.Type.Results.List) != 0 {
				continue
			}

			return true
		}
	}

	return false
}

func cutSpace(b []byte) (before, middle, after []byte) {
	i := 0
	for i < len(b) && (b[i] == ' ' || b[i] == '\t' || b[i] == '\n') {
		i++
	}
	j := len(b)
	for j > 0 && (b[j-1] == ' ' || b[j-1] == '\t' || b[j-1] == '\n') {
		j--
	}
	if i <= j {
		return b[:i], b[i:j], b[j:]
	}
	return nil, nil, b[j:]
}

// matchSpace reformats src to use the same space context as orig.
// 1) If orig begins with blank lines, matchSpace inserts them at the beginning of src.
// 2) matchSpace copies the indentation of the ***REMOVED***rst non-blank line in orig
//    to every non-blank line in src.
// 3) matchSpace copies the trailing space from orig and uses it in place
//   of src's trailing space.
func matchSpace(orig []byte, src []byte) []byte {
	before, _, after := cutSpace(orig)
	i := bytes.LastIndex(before, []byte{'\n'})
	before, indent := before[:i+1], before[i+1:]

	_, src, _ = cutSpace(src)

	var b bytes.Buffer
	b.Write(before)
	for len(src) > 0 {
		line := src
		if i := bytes.IndexByte(line, '\n'); i >= 0 {
			line, src = line[:i+1], line[i+1:]
		} ***REMOVED*** {
			src = nil
		}
		if len(line) > 0 && line[0] != '\n' { // not blank
			b.Write(indent)
		}
		b.Write(line)
	}
	b.Write(after)
	return b.Bytes()
}

var impLine = regexp.MustCompile(`^\s+(?:[\w\.]+\s+)?"(.+)"`)

func addImportSpaces(r io.Reader, breaks []string) ([]byte, error) {
	var out bytes.Buffer
	in := bu***REMOVED***o.NewReader(r)
	inImports := false
	done := false
	for {
		s, err := in.ReadString('\n')
		if err == io.EOF {
			break
		} ***REMOVED*** if err != nil {
			return nil, err
		}

		if !inImports && !done && strings.HasPre***REMOVED***x(s, "import") {
			inImports = true
		}
		if inImports && (strings.HasPre***REMOVED***x(s, "var") ||
			strings.HasPre***REMOVED***x(s, "func") ||
			strings.HasPre***REMOVED***x(s, "const") ||
			strings.HasPre***REMOVED***x(s, "type")) {
			done = true
			inImports = false
		}
		if inImports && len(breaks) > 0 {
			if m := impLine.FindStringSubmatch(s); m != nil {
				if m[1] == breaks[0] {
					out.WriteByte('\n')
					breaks = breaks[1:]
				}
			}
		}

		fmt.Fprint(&out, s)
	}
	return out.Bytes(), nil
}
