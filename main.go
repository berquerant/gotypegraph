package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/berquerant/gotypegraph/display"
	"github.com/berquerant/gotypegraph/load"
	"github.com/berquerant/gotypegraph/profile"
	"github.com/berquerant/gotypegraph/search"
	"github.com/berquerant/gotypegraph/util"
	"golang.org/x/tools/go/packages"
)

var (
	searchForeign    = flag.Bool("foreign", false, "Search definitions in foreign packages.")
	searchUniverse   = flag.Bool("universe", false, "Search definitions in builtin packages.")
	searchPrivate    = flag.Bool("private", false, "Search private definitions.")
	acceptNameRegex  = flag.String("accept.name", "", "Accept objects whose name matches this.")
	denyNameRegex    = flag.String("deny.name", "", "Deny objects whose name matches this.")
	acceptPkgRegex   = flag.String("accept.pkg", "", "Accept packages whose name matches this.")
	denyPkgRegex     = flag.String("deny.pkg", "", "Deny packages whose name matches this.")
	searchWorkerNum  = flag.Int("worker", 4, "Number of search workers.")
	searchBufferSize = flag.Int("buffer", 1000, "Size of search buffers.")
)

const usage = `Usage of gotypegraph:
  gotypegraph [flags] -type TYPE patterns...
Flags:`

func Usage() {
	fmt.Fprintln(os.Stderr, usage)
	flag.PrintDefaults()
}

func fail(err error) {
	if err != nil {
		panic(err)
	}
}

func compileRegex(v string) *regexp.Regexp {
	if v == "" {
		return nil
	}
	return regexp.MustCompile(v)
}

func loadPackages() []*packages.Package {
	pkgs, err := load.New().Load(flag.Args()...)
	fail(err)
	return pkgs
}

func newWriter() display.Writer {
	return display.NewJSONWriter(os.Stdout)
}

func newSearcher(pkgs []*packages.Package, opt ...search.UseSearcherOption) search.UseSearcher {
	var (
		defSetExtractor = search.NewDefSetExtractor(search.NewDefExtractor())
		defSetList      = make([]search.DefSet, len(pkgs))
	)
	for i, pkg := range pkgs {
		defSetList[i] = defSetExtractor.Extract(pkg)
	}
	return search.NewUseSearcher(
		pkgs,
		search.NewRefPkgSearcher(search.NewRefSearcher(), defSetList),
		search.NewObjExtractor(),
		search.NewTargetExtractor(),
		search.DefSetFilter(defSetList),
		opt...,
	)
}

func main() {
	flag.Usage = Usage
	flag.Parse()

	profiler := profile.NewProfiler(profile.NewStopwatch())
	profiler.Init()
	defer func() {
		fmt.Fprint(os.Stderr, profiler.Result().String())
	}()
	pkgs := loadPackages()
	profiler.PkgLoaded(pkgs)
	var (
		searcher = newSearcher(
			pkgs,
			search.WithUseSearcherSearchForeign(*searchForeign),
			search.WithUseSearcherSearchUniverse(*searchUniverse),
			search.WithUseSearcherSearchPrivate(*searchPrivate),
			search.WithUseSearcherPkgNameRegexp(util.NewRegexpPair(
				compileRegex(*acceptPkgRegex),
				compileRegex(*denyPkgRegex),
			)),
			search.WithUseSearcherObjNameRegexp(util.NewRegexpPair(
				compileRegex(*acceptNameRegex),
				compileRegex(*denyNameRegex),
			)),
			search.WithUseSearcherWorkerNum(*searchWorkerNum),
			search.WithUseSearcherResultBufferSize(*searchBufferSize),
		)
		writer = newWriter()
	)
	for result := range searcher.Search() {
		fail(writer.Write(result))
		profiler.Add(result)
	}
	profiler.Searched()
	fail(writer.Flush())
	profiler.Flushed()
}
