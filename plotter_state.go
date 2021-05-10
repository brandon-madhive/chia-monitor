package main

import (
	"regexp"
	"strconv"
	"time"
)

type PhaseTime struct {
	Phase    string
	Run      int
	Duration time.Duration
}

type PlotterState struct {
	State       map[string]string
	Pid         int
	Completions int
	PhaseTimes  []PhaseTime

	Phase       string
	Table       string
	Bucket      string
	PlotSz      int
	BucketCount int
	MaxRamMb    int
	MaxThread   int
	Duration    time.Duration
}

var processors = map[string]*regexp.Regexp{
	"plotSize":   regexp.MustCompile(`Plot size is: (\d+)`),
	"maxRam":     regexp.MustCompile(`Buffer size is: (\d+)MiB`),
	"bucketSize": regexp.MustCompile(`Using (\d+) buckets`),
	"phase":      regexp.MustCompile(`.*Starting phase (\d)/*.`),
	"table":      regexp.MustCompile(`.*table (\d)`),
	"bucket":     regexp.MustCompile(`.*Bucket (\d+)`),
}

var runCounter = regexp.MustCompile(`Total time = (\d+)`)
var phaseTime = regexp.MustCompile(`Time for phase (\d) = (\d+)`)
var copyTime = regexp.MustCompile(`Copy time = (\d+)`)

func checkRegex(s string, r *regexp.Regexp) ([]string, bool) {
	if r.Match([]byte(s)) {
		matches := r.FindStringSubmatch(s)
		if len(matches) > 1 {
			return matches[1:], true
		}
	}

	return nil, false
}

func (s *PlotterState) Update(entry *logEntry) {
	for k, r := range processors {
		if val, valid := checkRegex(entry.msg, r); valid {
			s.State[k] = val[0]
		}
	}

	if val, valid := checkRegex(entry.msg, phaseTime); valid {
		dur, _ := strconv.Atoi(val[1])
		ps := PhaseTime{
			Phase:    val[0],
			Run:      s.Completions,
			Duration: time.Second * time.Duration(dur),
		}
		// phase times
		s.PhaseTimes = append(s.PhaseTimes, ps)
	}

	if val, valid := checkRegex(entry.msg, copyTime); valid {
		dur, _ := strconv.Atoi(val[0])
		ps := PhaseTime{
			Phase:    "copy",
			Run:      s.Completions - 1, // copy happens after the run finshes
			Duration: time.Second * time.Duration(dur),
		}
		s.PhaseTimes = append(s.PhaseTimes, ps)
	}

	if val, valid := checkRegex(entry.msg, runCounter); valid {
		dur, _ := strconv.Atoi(val[0])
		ps := PhaseTime{
			Phase:    "final",
			Run:      s.Completions,
			Duration: time.Second * time.Duration(dur),
		}
		s.Completions++
		s.PhaseTimes = append(s.PhaseTimes, ps)
	}

	s.State["last"] = entry.msg
}
