package profile

import (
	"fmt"
	"time"

	"github.com/berquerant/gotypegraph/use"
	"github.com/berquerant/gotypegraph/util"
	"golang.org/x/tools/go/packages"
)

type Profiler struct {
	sw      Stopwatch
	pkgs    []*packages.Package
	results []*use.Result
}

func NewProfiler(sw Stopwatch) *Profiler {
	return &Profiler{
		sw:      sw,
		pkgs:    []*packages.Package{},
		results: []*use.Result{},
	}
}

// Init initializes the profiler.
// This must be called before any other methods of Profiler.
func (s *Profiler) Init() {
	s.sw.Init()
}

func (s *Profiler) Add(result *use.Result) {
	s.results = append(s.results, result)
}
func (s *Profiler) PkgLoaded(pkgs []*packages.Package) {
	s.pkgs = pkgs
	s.sw.Memory("PkgLoaded")
}
func (s *Profiler) Searched() {
	s.sw.Memory("Searched")
}
func (s *Profiler) Flushed() {
	s.sw.Memory("Flushed")
}

// Result generates the profile.
// This must be called after Init(), Add(), PkgLoaded(), Searched() and Flushed().
func (s *Profiler) Result() *Profile {
	var (
		profile Profile
		logs    = s.sw.Result()
	)

	profile.LoadedPkgNum = len(s.pkgs)
	profile.LoadPkgTime = logs["PkgLoaded"].ElapsedSegment
	for _, pkg := range s.pkgs {
		profile.LoadedDefsNum += len(pkg.TypesInfo.Defs)
		profile.LoadedUsesNum += len(pkg.TypesInfo.Uses)
	}

	profile.SearchedTime = logs["Searched"].ElapsedSegment
	profile.SearchedNum = len(s.results)
	var (
		searchedPkg = util.NewStringSet()
		searchedDef = util.NewStringSet()
		searchedRef = util.NewStringSet()
	)
	for _, result := range s.results {
		r := result.RefPair
		d := result.DefPair
		searchedPkg.Add(d.PkgName)
		searchedPkg.Add(r.Pkg.Name)
		searchedDef.Add(fmt.Sprintf("%s__%s", d.PkgName, d.Obj.Name()))
		searchedRef.Add(fmt.Sprintf("%s__%s", r.Pkg.Name, r.NodeName()))
	}
	profile.SearchedPkgNum = searchedPkg.Len()
	profile.SearchedDefNum = searchedDef.Len()
	profile.SearchedRefNum = searchedRef.Len()

	profile.FlushedTime = logs["Flushed"].ElapsedSegment

	profile.ElapsedTime = logs["Flushed"].Elapsed
	return &profile
}

type Profile struct {
	LoadedPkgNum   int
	LoadPkgTime    time.Duration
	LoadedDefsNum  int
	LoadedUsesNum  int
	SearchedNum    int
	SearchedPkgNum int
	SearchedRefNum int
	SearchedDefNum int
	SearchedTime   time.Duration
	FlushedTime    time.Duration
	ElapsedTime    time.Duration
}

func (s *Profile) String() string {
	var b util.StringBuilder
	b.Writelnf("Loaded pkgs:\t%d", s.LoadedPkgNum)
	b.Writelnf("Loaded defs:\t%d", s.LoadedDefsNum)
	b.Writelnf("Loaded uses:\t%d", s.LoadedUsesNum)
	b.Writelnf("Searched:\t%d", s.SearchedNum)
	b.Writelnf("Searched pkgs:\t%d", s.SearchedPkgNum)
	b.Writelnf("Searched defs:\t%d", s.SearchedDefNum)
	b.Writelnf("Searched refs:\t%d", s.SearchedRefNum)
	b.Writelnf("Load pkg time:\t%v", s.LoadPkgTime)
	b.Writelnf("Search time:\t%v", s.SearchedTime)
	b.Writelnf("Flush time:\t%v", s.FlushedTime)
	b.Writelnf("Elapsed time:\t%v", s.ElapsedTime)
	return b.String()
}

type Stopwatch interface {
	// Init starts the stopwatch.
	Init()
	// Memory records the elapsed time.
	Memory(id string)
	Result() map[string]*StopwatchRecord
}

func NewStopwatch() Stopwatch { return &stopwatch{} }

type StopwatchRecord struct {
	ID string
	// Elapsed is the elapsed time from Stopwatch.Init().
	Elapsed time.Duration
	// ElapsedSegment is the elapsed time from the previous Stopwatch.Memory()
	// If Stopwatch.Memory() is not called, this is the elapsed time from the Stopwatch.Init().
	ElapsedSegment time.Duration
}

type stopwatch struct {
	startTime        time.Time
	segmentStartTime time.Time
	records          map[string]*StopwatchRecord
}

func (s *stopwatch) Init() {
	now := time.Now()
	s.startTime = now
	s.segmentStartTime = now
	s.records = map[string]*StopwatchRecord{}
}

func (s *stopwatch) Memory(id string) {
	now := time.Now()
	s.records[id] = &StopwatchRecord{
		ID:             id,
		Elapsed:        now.Sub(s.startTime),
		ElapsedSegment: now.Sub(s.segmentStartTime),
	}
	s.segmentStartTime = now
}

func (s *stopwatch) Result() map[string]*StopwatchRecord { return s.records }
