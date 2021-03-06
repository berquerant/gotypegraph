package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/berquerant/gotypegraph/display"
	"github.com/berquerant/gotypegraph/load"
	"github.com/berquerant/gotypegraph/logger"
	"github.com/berquerant/gotypegraph/profile"
	"github.com/berquerant/gotypegraph/search"
	"github.com/berquerant/gotypegraph/util"
	"golang.org/x/tools/go/packages"
)

var (
	outputType       = flag.String("type", "dot", "Output format. json or dot.")
	useStat          = flag.Bool("stat", false, "Generate stat graph when type is dot.")
	searchForeign    = flag.Bool("foreign", false, "Search definitions in foreign packages.")
	searchUniverse   = flag.Bool("universe", false, "Search definitions in builtin packages.")
	searchPrivate    = flag.Bool("private", false, "Search private definitions.")
	ignoreSelfloop   = flag.Bool("noselfloop", false, "Ignore self references.")
	acceptNameRegex  = flag.String("accept.name", "", "Accept objects whose name matches this.")
	denyNameRegex    = flag.String("deny.name", "", "Deny objects whose name matches this.")
	acceptPkgRegex   = flag.String("accept.pkg", "", "Accept packages whose name matches this.")
	denyPkgRegex     = flag.String("deny.pkg", "", "Deny packages whose name matches this.")
	searchWorkerNum  = flag.Int("worker", 4, "Number of search workers.")
	searchBufferSize = flag.Int("buffer", 1000, "Size of search buffers.")
	minFontsize      = flag.Int("fontsize.min", 8, "Min fontsize used for text in dot.")
	maxFontsize      = flag.Int("fontsize.max", 24, "Max fontsize used for text in dot.")
	minPenwidth      = flag.Int("penwidth.min", 1, "Min penwidth used to draw lines in dot.")
	maxPenwidth      = flag.Int("penwidth.max", 1, "Max penwidth used to draw lines in dot.")
	minWeight        = flag.Int("weight.min", 1, "Min weight for dot.")
	maxWeight        = flag.Int("weight.max", 100, "Max weight for dot.")

	verbosity = flag.String("v", "info", "Logging verbosity. quiet, error, warn, info, verbose or debug.")
	quiet     = flag.Bool("quiet", false, "Quiet logs.")
	logRegexp = flag.String("log.regexp", "", "Regexp to grep logs.")
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
		logger.Errorf("%v", err)
		os.Exit(1)
	}
}

func initLogger() {
	logger.SetLevel(logLevel())
	logger.SetFilter(compileRegex(*logRegexp))
}

func logLevel() logger.Level {
	if *quiet {
		return logger.Quiet
	}
	x := strings.ToLower(*verbosity)
	pref := func(t string) bool { return strings.HasPrefix(x, t) }
	switch {
	case pref("v"):
		return logger.Verbose
	case pref("d"):
		return logger.Debug
	case pref("i"):
		return logger.Info
	case pref("w"):
		return logger.Warn
	case pref("e"):
		return logger.Error
	case pref("q"):
		return logger.Quiet
	default:
		return logger.Info
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

func writerOptions() []display.WriterOption {
	return []display.WriterOption{
		display.WithWriterMinFontsize(*minFontsize),
		display.WithWriterMaxFontsize(*maxFontsize),
		display.WithWriterMinPenwidth(*minPenwidth),
		display.WithWriterMaxPenwidth(*maxPenwidth),
		display.WithWriterMinWeight(*minWeight),
		display.WithWriterMaxWeight(*maxWeight),
	}
}

func newWriter() display.Writer {
	switch *outputType {
	case "dot":
		opt := writerOptions()
		if *useStat {
			return display.NewPackageDotWriter(os.Stdout, opt...)
		}
		return display.NewNodeDotWriter(os.Stdout, opt...)
	default:
		return display.NewJSONWriter(os.Stdout)
	}
}

func ignoreSelfloopOptions() []search.UseSearcherOption {
	if !*ignoreSelfloop {
		return nil
	}
	switch *outputType {
	case "dot":
		if *useStat {
			return []search.UseSearcherOption{search.WithUseSearcherIgnorePkgSelfloop(true)}
		}
		return []search.UseSearcherOption{search.WithUseSearcherIgnoreUseSelfloop(true)}
	default:
		return nil
	}
}

func searcherOptions() []search.UseSearcherOption {
	return append([]search.UseSearcherOption{
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
	}, ignoreSelfloopOptions()...)
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
		search.NewFieldSearcherFromPackages(pkgs),
		search.DefSetFilter(defSetList),
		opt...,
	)
}

func newProfiler() profile.Profiler {
	if *quiet {
		return profile.NewNullProfiler()
	}
	return profile.NewProfiler(profile.NewStopwatch())
}

func main() {
	flag.Usage = Usage
	flag.Parse()

	initLogger()
	profiler := newProfiler()
	profiler.Init()
	logger.Infof("Load packages")
	pkgs := loadPackages()
	profiler.PkgLoaded(pkgs)
	logger.Infof("%d packages loaded", len(pkgs))
	var (
		searcher = newSearcher(
			pkgs,
			searcherOptions()...,
		)
		writer = newWriter()
	)
	logger.Infof("Search and write")
	for result := range searcher.Search() {
		fail(writer.Write(result))
		profiler.Add(result)
	}
	profiler.Searched()
	logger.Infof("Flush")
	fail(writer.Flush())
	profiler.Flushed()
	profiler.Write(os.Stderr)
}
