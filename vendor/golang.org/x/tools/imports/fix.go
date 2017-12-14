// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE ***REMOVED***le.

package imports

import (
	"bu***REMOVED***o"
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/***REMOVED***lepath"
	"sort"
	"strings"
	"sync"

	"golang.org/x/tools/go/ast/astutil"
)

// Debug controls verbose logging.
var Debug = false

var (
	inTests = false      // set true by ***REMOVED***x_test.go; if false, no need to use testMu
	testMu  sync.RWMutex // guards globals reset by tests; used only if inTests
)

// LocalPre***REMOVED***x, if set, instructs Process to sort import paths with the given
// pre***REMOVED***x into another group after 3rd-party packages.
var LocalPre***REMOVED***x string

// importToGroup is a list of functions which map from an import path to
// a group number.
var importToGroup = []func(importPath string) (num int, ok bool){
	func(importPath string) (num int, ok bool) {
		if LocalPre***REMOVED***x != "" && strings.HasPre***REMOVED***x(importPath, LocalPre***REMOVED***x) {
			return 3, true
		}
		return
	},
	func(importPath string) (num int, ok bool) {
		if strings.HasPre***REMOVED***x(importPath, "appengine") {
			return 2, true
		}
		return
	},
	func(importPath string) (num int, ok bool) {
		if strings.Contains(importPath, ".") {
			return 1, true
		}
		return
	},
}

func importGroup(importPath string) int {
	for _, fn := range importToGroup {
		if n, ok := fn(importPath); ok {
			return n
		}
	}
	return 0
}

// importInfo is a summary of information about one import.
type importInfo struct {
	Path  string // full import path (e.g. "crypto/rand")
	Alias string // import alias, if present (e.g. "crand")
}

// packageInfo is a summary of features found in a package.
type packageInfo struct {
	Globals map[string]bool       // symbol => true
	Imports map[string]importInfo // pkg base name or alias => info
}

// dirPackageInfo exposes the dirPackageInfoFile function so that it can be overridden.
var dirPackageInfo = dirPackageInfoFile

// dirPackageInfoFile gets information from other ***REMOVED***les in the package.
func dirPackageInfoFile(pkgName, srcDir, ***REMOVED***lename string) (*packageInfo, error) {
	considerTests := strings.HasSuf***REMOVED***x(***REMOVED***lename, "_test.go")

	***REMOVED***leBase := ***REMOVED***lepath.Base(***REMOVED***lename)
	packageFileInfos, err := ioutil.ReadDir(srcDir)
	if err != nil {
		return nil, err
	}

	info := &packageInfo{Globals: make(map[string]bool), Imports: make(map[string]importInfo)}
	for _, ***REMOVED*** := range packageFileInfos {
		if ***REMOVED***.Name() == ***REMOVED***leBase || !strings.HasSuf***REMOVED***x(***REMOVED***.Name(), ".go") {
			continue
		}
		if !considerTests && strings.HasSuf***REMOVED***x(***REMOVED***.Name(), "_test.go") {
			continue
		}

		***REMOVED***leSet := token.NewFileSet()
		root, err := parser.ParseFile(***REMOVED***leSet, ***REMOVED***lepath.Join(srcDir, ***REMOVED***.Name()), nil, 0)
		if err != nil {
			continue
		}

		for _, decl := range root.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			for _, spec := range genDecl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				info.Globals[valueSpec.Names[0].Name] = true
			}
		}

		for _, imp := range root.Imports {
			impInfo := importInfo{Path: strings.Trim(imp.Path.Value, `"`)}
			name := path.Base(impInfo.Path)
			if imp.Name != nil {
				name = strings.Trim(imp.Name.Name, `"`)
				impInfo.Alias = name
			}
			info.Imports[name] = impInfo
		}
	}
	return info, nil
}

func ***REMOVED***xImports(fset *token.FileSet, f *ast.File, ***REMOVED***lename string) (added []string, err error) {
	// refs are a set of possible package references currently unsatis***REMOVED***ed by imports.
	// ***REMOVED***rst key: either base package (e.g. "fmt") or renamed package
	// second key: referenced package symbol (e.g. "Println")
	refs := make(map[string]map[string]bool)

	// decls are the current package imports. key is base package or renamed package.
	decls := make(map[string]*ast.ImportSpec)

	abs, err := ***REMOVED***lepath.Abs(***REMOVED***lename)
	if err != nil {
		return nil, err
	}
	srcDir := ***REMOVED***lepath.Dir(abs)
	if Debug {
		log.Printf("***REMOVED***xImports(***REMOVED***lename=%q), abs=%q, srcDir=%q ...", ***REMOVED***lename, abs, srcDir)
	}

	var packageInfo *packageInfo
	var loadedPackageInfo bool

	// collect potential uses of packages.
	var visitor visitFn
	visitor = visitFn(func(node ast.Node) ast.Visitor {
		if node == nil {
			return visitor
		}
		switch v := node.(type) {
		case *ast.ImportSpec:
			if v.Name != nil {
				decls[v.Name.Name] = v
				break
			}
			ipath := strings.Trim(v.Path.Value, `"`)
			if ipath == "C" {
				break
			}
			local := importPathToName(ipath, srcDir)
			decls[local] = v
		case *ast.SelectorExpr:
			xident, ok := v.X.(*ast.Ident)
			if !ok {
				break
			}
			if xident.Obj != nil {
				// if the parser can resolve it, it's not a package ref
				break
			}
			pkgName := xident.Name
			if refs[pkgName] == nil {
				refs[pkgName] = make(map[string]bool)
			}
			if !loadedPackageInfo {
				loadedPackageInfo = true
				packageInfo, _ = dirPackageInfo(f.Name.Name, srcDir, ***REMOVED***lename)
			}
			if decls[pkgName] == nil && (packageInfo == nil || !packageInfo.Globals[pkgName]) {
				refs[pkgName][v.Sel.Name] = true
			}
		}
		return visitor
	})
	ast.Walk(visitor, f)

	// Nil out any unused ImportSpecs, to be removed in following passes
	unusedImport := map[string]string{}
	for pkg, is := range decls {
		if refs[pkg] == nil && pkg != "_" && pkg != "." {
			name := ""
			if is.Name != nil {
				name = is.Name.Name
			}
			unusedImport[strings.Trim(is.Path.Value, `"`)] = name
		}
	}
	for ipath, name := range unusedImport {
		if ipath == "C" {
			// Don't remove cgo stuff.
			continue
		}
		astutil.DeleteNamedImport(fset, f, name, ipath)
	}

	for pkgName, symbols := range refs {
		if len(symbols) == 0 {
			// skip over packages already imported
			delete(refs, pkgName)
		}
	}

	// Fast path, all references already imported.
	if len(refs) == 0 {
		return nil, nil
	}

	// Can assume this will be necessary in all cases now.
	if !loadedPackageInfo {
		packageInfo, _ = dirPackageInfo(f.Name.Name, srcDir, ***REMOVED***lename)
	}

	// Search for imports matching potential package references.
	searches := 0
	type result struct {
		ipath string // import path (if err == nil)
		name  string // optional name to rename import as
		err   error
	}
	results := make(chan result)
	for pkgName, symbols := range refs {
		go func(pkgName string, symbols map[string]bool) {
			if packageInfo != nil {
				sibling := packageInfo.Imports[pkgName]
				if sibling.Path != "" {
					results <- result{ipath: sibling.Path, name: sibling.Alias}
					return
				}
			}
			ipath, rename, err := ***REMOVED***ndImport(pkgName, symbols, ***REMOVED***lename)
			r := result{ipath: ipath, err: err}
			if rename {
				r.name = pkgName
			}
			results <- r
		}(pkgName, symbols)
		searches++
	}
	for i := 0; i < searches; i++ {
		result := <-results
		if result.err != nil {
			return nil, result.err
		}
		if result.ipath != "" {
			if result.name != "" {
				astutil.AddNamedImport(fset, f, result.name, result.ipath)
			} ***REMOVED*** {
				astutil.AddImport(fset, f, result.ipath)
			}
			added = append(added, result.ipath)
		}
	}

	return added, nil
}

// importPathToName returns the package name for the given import path.
var importPathToName func(importPath, srcDir string) (packageName string) = importPathToNameGoPath

// importPathToNameBasic assumes the package name is the base of import path.
func importPathToNameBasic(importPath, srcDir string) (packageName string) {
	return path.Base(importPath)
}

// importPathToNameGoPath ***REMOVED***nds out the actual package name, as declared in its .go ***REMOVED***les.
// If there's a problem, it falls back to using importPathToNameBasic.
func importPathToNameGoPath(importPath, srcDir string) (packageName string) {
	// Fast path for standard library without going to disk.
	if pkg, ok := stdImportPackage[importPath]; ok {
		return pkg
	}

	pkgName, err := importPathToNameGoPathParse(importPath, srcDir)
	if Debug {
		log.Printf("importPathToNameGoPathParse(%q, srcDir=%q) = %q, %v", importPath, srcDir, pkgName, err)
	}
	if err == nil {
		return pkgName
	}
	return importPathToNameBasic(importPath, srcDir)
}

// importPathToNameGoPathParse is a faster version of build.Import if
// the only thing desired is the package name. It uses build.FindOnly
// to ***REMOVED***nd the directory and then only parses one ***REMOVED***le in the package,
// trusting that the ***REMOVED***les in the directory are consistent.
func importPathToNameGoPathParse(importPath, srcDir string) (packageName string, err error) {
	buildPkg, err := build.Import(importPath, srcDir, build.FindOnly)
	if err != nil {
		return "", err
	}
	d, err := os.Open(buildPkg.Dir)
	if err != nil {
		return "", err
	}
	names, err := d.Readdirnames(-1)
	d.Close()
	if err != nil {
		return "", err
	}
	sort.Strings(names) // to have predictable behavior
	var lastErr error
	var n***REMOVED***le int
	for _, name := range names {
		if !strings.HasSuf***REMOVED***x(name, ".go") {
			continue
		}
		if strings.HasSuf***REMOVED***x(name, "_test.go") {
			continue
		}
		n***REMOVED***le++
		fullFile := ***REMOVED***lepath.Join(buildPkg.Dir, name)

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, fullFile, nil, parser.PackageClauseOnly)
		if err != nil {
			lastErr = err
			continue
		}
		pkgName := f.Name.Name
		if pkgName == "documentation" {
			// Special case from go/build.ImportDir, not
			// handled by ctx.MatchFile.
			continue
		}
		if pkgName == "main" {
			// Also skip package main, assuming it's a +build ignore generator or example.
			// Since you can't import a package main anyway, there's no harm here.
			continue
		}
		return pkgName, nil
	}
	if lastErr != nil {
		return "", lastErr
	}
	return "", fmt.Errorf("no importable package found in %d Go ***REMOVED***les", n***REMOVED***le)
}

var stdImportPackage = map[string]string{} // "net/http" => "http"

func init() {
	// Nothing in the standard library has a package name not
	// matching its import base name.
	for _, pkg := range stdlib {
		if _, ok := stdImportPackage[pkg]; !ok {
			stdImportPackage[pkg] = path.Base(pkg)
		}
	}
}

// Directory-scanning state.
var (
	// scanGoRootOnce guards calling scanGoRoot (for $GOROOT)
	scanGoRootOnce sync.Once
	// scanGoPathOnce guards calling scanGoPath (for $GOPATH)
	scanGoPathOnce sync.Once

	// populateIgnoreOnce guards calling populateIgnore
	populateIgnoreOnce sync.Once
	ignoredDirs        []os.FileInfo

	dirScanMu sync.RWMutex
	dirScan   map[string]*pkg // abs dir path => *pkg
)

type pkg struct {
	dir             string // absolute ***REMOVED***le path to pkg directory ("/usr/lib/go/src/net/http")
	importPath      string // full pkg import path ("net/http", "foo/bar/vendor/a/b")
	importPathShort string // vendorless import path ("net/http", "a/b")
	distance        int    // relative distance to target
}

// byDistanceOrImportPathShortLength sorts by relative distance breaking ties
// on the short import path length and then the import string itself.
type byDistanceOrImportPathShortLength []*pkg

func (s byDistanceOrImportPathShortLength) Len() int { return len(s) }
func (s byDistanceOrImportPathShortLength) Less(i, j int) bool {
	di, dj := s[i].distance, s[j].distance
	if di == -1 {
		return false
	}
	if dj == -1 {
		return true
	}
	if di != dj {
		return di < dj
	}

	vi, vj := s[i].importPathShort, s[j].importPathShort
	if len(vi) != len(vj) {
		return len(vi) < len(vj)
	}
	return vi < vj
}
func (s byDistanceOrImportPathShortLength) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func distance(basepath, targetpath string) int {
	p, err := ***REMOVED***lepath.Rel(basepath, targetpath)
	if err != nil {
		return -1
	}
	if p == "." {
		return 0
	}
	return strings.Count(p, string(***REMOVED***lepath.Separator)) + 1
}

// guarded by populateIgnoreOnce; populates ignoredDirs.
func populateIgnore() {
	for _, srcDir := range build.Default.SrcDirs() {
		if srcDir == ***REMOVED***lepath.Join(build.Default.GOROOT, "src") {
			continue
		}
		populateIgnoredDirs(srcDir)
	}
}

// populateIgnoredDirs reads an optional con***REMOVED***g ***REMOVED***le at <path>/.goimportsignore
// of relative directories to ignore when scanning for go ***REMOVED***les.
// The provided path is one of the $GOPATH entries with "src" appended.
func populateIgnoredDirs(path string) {
	***REMOVED***le := ***REMOVED***lepath.Join(path, ".goimportsignore")
	slurp, err := ioutil.ReadFile(***REMOVED***le)
	if Debug {
		if err != nil {
			log.Print(err)
		} ***REMOVED*** {
			log.Printf("Read %s", ***REMOVED***le)
		}
	}
	if err != nil {
		return
	}
	bs := bu***REMOVED***o.NewScanner(bytes.NewReader(slurp))
	for bs.Scan() {
		line := strings.TrimSpace(bs.Text())
		if line == "" || strings.HasPre***REMOVED***x(line, "#") {
			continue
		}
		full := ***REMOVED***lepath.Join(path, line)
		if ***REMOVED***, err := os.Stat(full); err == nil {
			ignoredDirs = append(ignoredDirs, ***REMOVED***)
			if Debug {
				log.Printf("Directory added to ignore list: %s", full)
			}
		} ***REMOVED*** if Debug {
			log.Printf("Error statting entry in .goimportsignore: %v", err)
		}
	}
}

func skipDir(***REMOVED*** os.FileInfo) bool {
	for _, ignoredDir := range ignoredDirs {
		if os.SameFile(***REMOVED***, ignoredDir) {
			return true
		}
	}
	return false
}

// shouldTraverse reports whether the symlink ***REMOVED*** should, found in dir,
// should be followed.  It makes sure symlinks were never visited
// before to avoid symlink loops.
func shouldTraverse(dir string, ***REMOVED*** os.FileInfo) bool {
	path := ***REMOVED***lepath.Join(dir, ***REMOVED***.Name())
	target, err := ***REMOVED***lepath.EvalSymlinks(path)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, err)
		}
		return false
	}
	ts, err := os.Stat(target)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}
	if !ts.IsDir() {
		return false
	}
	if skipDir(ts) {
		return false
	}
	// Check for symlink loops by statting each directory component
	// and seeing if any are the same ***REMOVED***le as ts.
	for {
		parent := ***REMOVED***lepath.Dir(path)
		if parent == path {
			// Made it to the root without seeing a cycle.
			// Use this symlink.
			return true
		}
		parentInfo, err := os.Stat(parent)
		if err != nil {
			return false
		}
		if os.SameFile(ts, parentInfo) {
			// Cycle. Don't traverse.
			return false
		}
		path = parent
	}

}

var testHookScanDir = func(dir string) {}

var scanGoRootDone = make(chan struct{}) // closed when scanGoRoot is done

func scanGoRoot() {
	go func() {
		scanGoDirs(true)
		close(scanGoRootDone)
	}()
}

func scanGoPath() { scanGoDirs(false) }

func scanGoDirs(goRoot bool) {
	if Debug {
		which := "$GOROOT"
		if !goRoot {
			which = "$GOPATH"
		}
		log.Printf("scanning " + which)
		defer log.Printf("scanned " + which)
	}
	dirScanMu.Lock()
	if dirScan == nil {
		dirScan = make(map[string]*pkg)
	}
	dirScanMu.Unlock()

	for _, srcDir := range build.Default.SrcDirs() {
		isGoroot := srcDir == ***REMOVED***lepath.Join(build.Default.GOROOT, "src")
		if isGoroot != goRoot {
			continue
		}
		testHookScanDir(srcDir)
		walkFn := func(path string, typ os.FileMode) error {
			dir := ***REMOVED***lepath.Dir(path)
			if typ.IsRegular() {
				if dir == srcDir {
					// Doesn't make sense to have regular ***REMOVED***les
					// directly in your $GOPATH/src or $GOROOT/src.
					return nil
				}
				if !strings.HasSuf***REMOVED***x(path, ".go") {
					return nil
				}
				dirScanMu.Lock()
				if _, dup := dirScan[dir]; !dup {
					importpath := ***REMOVED***lepath.ToSlash(dir[len(srcDir)+len("/"):])
					dirScan[dir] = &pkg{
						importPath:      importpath,
						importPathShort: vendorlessImportPath(importpath),
						dir:             dir,
					}
				}
				dirScanMu.Unlock()
				return nil
			}
			if typ == os.ModeDir {
				base := ***REMOVED***lepath.Base(path)
				if base == "" || base[0] == '.' || base[0] == '_' ||
					base == "testdata" || base == "node_modules" {
					return ***REMOVED***lepath.SkipDir
				}
				***REMOVED***, err := os.Lstat(path)
				if err == nil && skipDir(***REMOVED***) {
					if Debug {
						log.Printf("skipping directory %q under %s", ***REMOVED***.Name(), dir)
					}
					return ***REMOVED***lepath.SkipDir
				}
				return nil
			}
			if typ == os.ModeSymlink {
				base := ***REMOVED***lepath.Base(path)
				if strings.HasPre***REMOVED***x(base, ".#") {
					// Emacs noise.
					return nil
				}
				***REMOVED***, err := os.Lstat(path)
				if err != nil {
					// Just ignore it.
					return nil
				}
				if shouldTraverse(dir, ***REMOVED***) {
					return traverseLink
				}
			}
			return nil
		}
		if err := fastWalk(srcDir, walkFn); err != nil {
			log.Printf("goimports: scanning directory %v: %v", srcDir, err)
		}
	}
}

// vendorlessImportPath returns the devendorized version of the provided import path.
// e.g. "foo/bar/vendor/a/b" => "a/b"
func vendorlessImportPath(ipath string) string {
	// Devendorize for use in import statement.
	if i := strings.LastIndex(ipath, "/vendor/"); i >= 0 {
		return ipath[i+len("/vendor/"):]
	}
	if strings.HasPre***REMOVED***x(ipath, "vendor/") {
		return ipath[len("vendor/"):]
	}
	return ipath
}

// loadExports returns the set of exported symbols in the package at dir.
// It returns nil on error or if the package name in dir does not match expectPackage.
var loadExports func(expectPackage, dir string) map[string]bool = loadExportsGoPath

func loadExportsGoPath(expectPackage, dir string) map[string]bool {
	if Debug {
		log.Printf("loading exports in dir %s (seeking package %s)", dir, expectPackage)
	}
	exports := make(map[string]bool)

	ctx := build.Default

	// ReadDir is like ioutil.ReadDir, but only returns *.go ***REMOVED***les
	// and ***REMOVED***lters out _test.go ***REMOVED***les since they're not relevant
	// and only slow things down.
	ctx.ReadDir = func(dir string) (notTests []os.FileInfo, err error) {
		all, err := ioutil.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		notTests = all[:0]
		for _, ***REMOVED*** := range all {
			name := ***REMOVED***.Name()
			if strings.HasSuf***REMOVED***x(name, ".go") && !strings.HasSuf***REMOVED***x(name, "_test.go") {
				notTests = append(notTests, ***REMOVED***)
			}
		}
		return notTests, nil
	}

	***REMOVED***les, err := ctx.ReadDir(dir)
	if err != nil {
		log.Print(err)
		return nil
	}

	fset := token.NewFileSet()

	for _, ***REMOVED*** := range ***REMOVED***les {
		match, err := ctx.MatchFile(dir, ***REMOVED***.Name())
		if err != nil || !match {
			continue
		}
		fullFile := ***REMOVED***lepath.Join(dir, ***REMOVED***.Name())
		f, err := parser.ParseFile(fset, fullFile, nil, 0)
		if err != nil {
			if Debug {
				log.Printf("Parsing %s: %v", fullFile, err)
			}
			return nil
		}
		pkgName := f.Name.Name
		if pkgName == "documentation" {
			// Special case from go/build.ImportDir, not
			// handled by ctx.MatchFile.
			continue
		}
		if pkgName != expectPackage {
			if Debug {
				log.Printf("scan of dir %v is not expected package %v (actually %v)", dir, expectPackage, pkgName)
			}
			return nil
		}
		for name := range f.Scope.Objects {
			if ast.IsExported(name) {
				exports[name] = true
			}
		}
	}

	if Debug {
		exportList := make([]string, 0, len(exports))
		for k := range exports {
			exportList = append(exportList, k)
		}
		sort.Strings(exportList)
		log.Printf("loaded exports in dir %v (package %v): %v", dir, expectPackage, strings.Join(exportList, ", "))
	}
	return exports
}

// ***REMOVED***ndImport searches for a package with the given symbols.
// If no package is found, ***REMOVED***ndImport returns ("", false, nil)
//
// This is declared as a variable rather than a function so goimports
// can be easily extended by adding a ***REMOVED***le with an init function.
//
// The rename value tells goimports whether to use the package name as
// a local quali***REMOVED***er in an import. For example, if ***REMOVED***ndImports("pkg",
// "X") returns ("foo/bar", rename=true), then goimports adds the
// import line:
// 	import pkg "foo/bar"
// to satisfy uses of pkg.X in the ***REMOVED***le.
var ***REMOVED***ndImport func(pkgName string, symbols map[string]bool, ***REMOVED***lename string) (foundPkg string, rename bool, err error) = ***REMOVED***ndImportGoPath

// ***REMOVED***ndImportGoPath is the normal implementation of ***REMOVED***ndImport.
// (Some companies have their own internally.)
func ***REMOVED***ndImportGoPath(pkgName string, symbols map[string]bool, ***REMOVED***lename string) (foundPkg string, rename bool, err error) {
	if inTests {
		testMu.RLock()
		defer testMu.RUnlock()
	}

	pkgDir, err := ***REMOVED***lepath.Abs(***REMOVED***lename)
	if err != nil {
		return "", false, err
	}
	pkgDir = ***REMOVED***lepath.Dir(pkgDir)

	// Fast path for the standard library.
	// In the common case we hopefully never have to scan the GOPATH, which can
	// be slow with moving disks.
	if pkg, rename, ok := ***REMOVED***ndImportStdlib(pkgName, symbols); ok {
		return pkg, rename, nil
	}
	if pkgName == "rand" && symbols["Read"] {
		// Special-case rand.Read.
		//
		// If ***REMOVED***ndImportStdlib didn't ***REMOVED***nd it above, don't go
		// searching for it, lest it ***REMOVED***nd and pick math/rand
		// in GOROOT (new as of Go 1.6)
		//
		// crypto/rand is the safer choice.
		return "", false, nil
	}

	// TODO(sameer): look at the import lines for other Go ***REMOVED***les in the
	// local directory, since the user is likely to import the same packages
	// in the current Go ***REMOVED***le.  Return rename=true when the other Go ***REMOVED***les
	// use a renamed package that's also used in the current ***REMOVED***le.

	// Read all the $GOPATH/src/.goimportsignore ***REMOVED***les before scanning directories.
	populateIgnoreOnce.Do(populateIgnore)

	// Start scanning the $GOROOT asynchronously, then run the
	// GOPATH scan synchronously if needed, and then wait for the
	// $GOROOT to ***REMOVED***nish.
	//
	// TODO(brad***REMOVED***tz): run each $GOPATH entry async. But nobody
	// really has more than one anyway, so low priority.
	scanGoRootOnce.Do(scanGoRoot) // async
	if !***REMOVED***leInDir(***REMOVED***lename, build.Default.GOROOT) {
		scanGoPathOnce.Do(scanGoPath) // blocking
	}
	<-scanGoRootDone

	// Find candidate packages, looking only at their directory names ***REMOVED***rst.
	var candidates []*pkg
	for _, pkg := range dirScan {
		if pkgIsCandidate(***REMOVED***lename, pkgName, pkg) {
			pkg.distance = distance(pkgDir, pkg.dir)
			candidates = append(candidates, pkg)
		}
	}

	// Sort the candidates by their import package length,
	// assuming that shorter package names are better than long
	// ones.  Note that this sorts by the de-vendored name, so
	// there's no "penalty" for vendoring.
	sort.Sort(byDistanceOrImportPathShortLength(candidates))
	if Debug {
		for i, pkg := range candidates {
			log.Printf("%s candidate %d/%d: %v in %v", pkgName, i+1, len(candidates), pkg.importPathShort, pkg.dir)
		}
	}

	// Collect exports for packages with matching names.

	done := make(chan struct{}) // closed when we ***REMOVED***nd the answer
	defer close(done)

	rescv := make([]chan *pkg, len(candidates))
	for i := range candidates {
		rescv[i] = make(chan *pkg)
	}
	const maxConcurrentPackageImport = 4
	loadExportsSem := make(chan struct{}, maxConcurrentPackageImport)

	go func() {
		for i, pkg := range candidates {
			select {
			case loadExportsSem <- struct{}{}:
				select {
				case <-done:
					return
				default:
				}
			case <-done:
				return
			}
			pkg := pkg
			resc := rescv[i]
			go func() {
				if inTests {
					testMu.RLock()
					defer testMu.RUnlock()
				}
				defer func() { <-loadExportsSem }()
				exports := loadExports(pkgName, pkg.dir)

				// If it doesn't have the right
				// symbols, send nil to mean no match.
				for symbol := range symbols {
					if !exports[symbol] {
						pkg = nil
						break
					}
				}
				select {
				case resc <- pkg:
				case <-done:
				}
			}()
		}
	}()
	for _, resc := range rescv {
		pkg := <-resc
		if pkg == nil {
			continue
		}
		// If the package name in the source doesn't match the import path's base,
		// return true so the rewriter adds a name (import foo "github.com/bar/go-foo")
		needsRename := path.Base(pkg.importPath) != pkgName
		return pkg.importPathShort, needsRename, nil
	}
	return "", false, nil
}

// pkgIsCandidate reports whether pkg is a candidate for satisfying the
// ***REMOVED***nding which package pkgIdent in the ***REMOVED***le named by ***REMOVED***lename is trying
// to refer to.
//
// This check is purely lexical and is meant to be as fast as possible
// because it's run over all $GOPATH directories to ***REMOVED***lter out poor
// candidates in order to limit the CPU and I/O later parsing the
// exports in candidate packages.
//
// ***REMOVED***lename is the ***REMOVED***le being formatted.
// pkgIdent is the package being searched for, like "client" (if
// searching for "client.New")
func pkgIsCandidate(***REMOVED***lename, pkgIdent string, pkg *pkg) bool {
	// Check "internal" and "vendor" visibility:
	if !canUse(***REMOVED***lename, pkg.dir) {
		return false
	}

	// Speed optimization to minimize disk I/O:
	// the last two components on disk must contain the
	// package name somewhere.
	//
	// This permits mismatch naming like directory
	// "go-foo" being package "foo", or "pkg.v3" being "pkg",
	// or directory "google.golang.org/api/cloudbilling/v1"
	// being package "cloudbilling", but doesn't
	// permit a directory "foo" to be package
	// "bar", which is strongly discouraged
	// anyway. There's no reason goimports needs
	// to be slow just to accomodate that.
	lastTwo := lastTwoComponents(pkg.importPathShort)
	if strings.Contains(lastTwo, pkgIdent) {
		return true
	}
	if hasHyphenOrUpperASCII(lastTwo) && !hasHyphenOrUpperASCII(pkgIdent) {
		lastTwo = lowerASCIIAndRemoveHyphen(lastTwo)
		if strings.Contains(lastTwo, pkgIdent) {
			return true
		}
	}

	return false
}

func hasHyphenOrUpperASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b == '-' || ('A' <= b && b <= 'Z') {
			return true
		}
	}
	return false
}

func lowerASCIIAndRemoveHyphen(s string) (ret string) {
	buf := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		b := s[i]
		switch {
		case b == '-':
			continue
		case 'A' <= b && b <= 'Z':
			buf = append(buf, b+('a'-'A'))
		default:
			buf = append(buf, b)
		}
	}
	return string(buf)
}

// canUse reports whether the package in dir is usable from ***REMOVED***lename,
// respecting the Go "internal" and "vendor" visibility rules.
func canUse(***REMOVED***lename, dir string) bool {
	// Fast path check, before any allocations. If it doesn't contain vendor
	// or internal, it's not tricky:
	// Note that this can false-negative on directories like "notinternal",
	// but we check it correctly below. This is just a fast path.
	if !strings.Contains(dir, "vendor") && !strings.Contains(dir, "internal") {
		return true
	}

	dirSlash := ***REMOVED***lepath.ToSlash(dir)
	if !strings.Contains(dirSlash, "/vendor/") && !strings.Contains(dirSlash, "/internal/") && !strings.HasSuf***REMOVED***x(dirSlash, "/internal") {
		return true
	}
	// Vendor or internal directory only visible from children of parent.
	// That means the path from the current directory to the target directory
	// can contain ../vendor or ../internal but not ../foo/vendor or ../foo/internal
	// or bar/vendor or bar/internal.
	// After stripping all the leading ../, the only okay place to see vendor or internal
	// is at the very beginning of the path.
	abs***REMOVED***le, err := ***REMOVED***lepath.Abs(***REMOVED***lename)
	if err != nil {
		return false
	}
	absdir, err := ***REMOVED***lepath.Abs(dir)
	if err != nil {
		return false
	}
	rel, err := ***REMOVED***lepath.Rel(abs***REMOVED***le, absdir)
	if err != nil {
		return false
	}
	relSlash := ***REMOVED***lepath.ToSlash(rel)
	if i := strings.LastIndex(relSlash, "../"); i >= 0 {
		relSlash = relSlash[i+len("../"):]
	}
	return !strings.Contains(relSlash, "/vendor/") && !strings.Contains(relSlash, "/internal/") && !strings.HasSuf***REMOVED***x(relSlash, "/internal")
}

// lastTwoComponents returns at most the last two path components
// of v, using either / or \ as the path separator.
func lastTwoComponents(v string) string {
	nslash := 0
	for i := len(v) - 1; i >= 0; i-- {
		if v[i] == '/' || v[i] == '\\' {
			nslash++
			if nslash == 2 {
				return v[i:]
			}
		}
	}
	return v
}

type visitFn func(node ast.Node) ast.Visitor

func (fn visitFn) Visit(node ast.Node) ast.Visitor {
	return fn(node)
}

func ***REMOVED***ndImportStdlib(shortPkg string, symbols map[string]bool) (importPath string, rename, ok bool) {
	for symbol := range symbols {
		key := shortPkg + "." + symbol
		path := stdlib[key]
		if path == "" {
			if key == "rand.Read" {
				continue
			}
			return "", false, false
		}
		if importPath != "" && importPath != path {
			// Ambiguous. Symbols pointed to different things.
			return "", false, false
		}
		importPath = path
	}
	if importPath == "" && shortPkg == "rand" && symbols["Read"] {
		return "crypto/rand", false, true
	}
	return importPath, false, importPath != ""
}

// ***REMOVED***leInDir reports whether the provided ***REMOVED***le path looks like
// it's in dir. (without hitting the ***REMOVED***lesystem)
func ***REMOVED***leInDir(***REMOVED***le, dir string) bool {
	rest := strings.TrimPre***REMOVED***x(***REMOVED***le, dir)
	if len(rest) == len(***REMOVED***le) {
		// dir is not a pre***REMOVED***x of ***REMOVED***le.
		return false
	}
	// Check for boundary: either nothing (***REMOVED***le == dir), or a slash.
	return len(rest) == 0 || rest[0] == '/' || rest[0] == '\\'
}
