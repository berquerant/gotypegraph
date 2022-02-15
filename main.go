package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/berquerant/gotypegraph/def"
	"github.com/berquerant/gotypegraph/display"
	"github.com/berquerant/gotypegraph/load"
	"github.com/berquerant/gotypegraph/profile"
	"github.com/berquerant/gotypegraph/ref"
	"github.com/berquerant/gotypegraph/use"
	"golang.org/x/tools/go/packages"
)

var (
	outputType      = flag.String("type", "dot", "Output format. string, json or dot.")
	searchForeign   = flag.Bool("foreign", false, "Search definitions in foreign packages.")
	searchUniverse  = flag.Bool("universe", false, "Search definitions in builtin packages.")
	searchPrivate   = flag.Bool("private", false, "Search private definitions.")
	stat            = flag.Bool("stat", false, "Generate stat graph when type is dot.")
	acceptNameRegex = flag.String("accept.name", "", "Accept objects whose name matches this.")
	denyNameRegex   = flag.String("deny.name", "", "Deny objects whose name matches this.")
	acceptPkgRegex  = flag.String("accept.pkg", "", "Accept packages whose name matches this.")
	denyPkgRegex    = flag.String("deny.pkg", "", "Deny packages whose name matches this.")
)

// TODO: more filter options

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

func loadPackages() []*packages.Package {
	pkgs, err := load.New().Load(flag.Args()...)
	fail(err)
	return pkgs
}

func compileRegex(v string) *regexp.Regexp {
	if v == "" {
		return nil
	}
	return regexp.MustCompile(v)
}

func newSearcher(pkgs []*packages.Package) use.Searcher {
	var (
		setExtractor = def.NewSetExtractor(def.NewExtactor())
		sets         = make([]*def.Set, len(pkgs))
	)
	for i, pkg := range pkgs {
		sets[i] = setExtractor.Extract(pkg)
	}
	return use.NewSearcher(
		pkgs,
		ref.NewSearcher(ref.NewLocalSearcherSet(sets)),
		use.DefFilter(sets),
		use.WithSearchForeign(*searchForeign),
		use.WithSearchUniverse(*searchUniverse),
		use.WithSearchPrivate(*searchPrivate),
		use.WithAcceptNameRegex(compileRegex(*acceptNameRegex)),
		use.WithDenyNameRegex(compileRegex(*denyNameRegex)),
		use.WithAcceptPkgRegex(compileRegex(*acceptPkgRegex)),
		use.WithDenyPkgRegex(compileRegex(*denyPkgRegex)),
	)
}

func newWriter() display.Writer {
	switch *outputType {
	case "string":
		return display.NewStringWriter(os.Stdout)
	case "json":
		return display.NewJSONWriter(os.Stdout)
	case "dot":
		if *stat {
			return display.NewStatDotWriter(os.Stdout)
		}
		return display.NewDotWriter(os.Stdout)
	default:
		panic(fmt.Sprintf("unknown output type: %s", *outputType))
	}
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
		searcher = newSearcher(pkgs)
		writer   = newWriter()
	)
	for result := range searcher.Search() {
		fail(writer.Write(result))
		profiler.Add(result)
	}
	profiler.Searched()
	fail(writer.Flush())
	profiler.Flushed()
}
