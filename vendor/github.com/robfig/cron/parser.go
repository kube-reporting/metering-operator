package cron

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// Con***REMOVED***guration options for creating a parser. Most options specify which
// ***REMOVED***elds should be included, while others enable features. If a ***REMOVED***eld is not
// included the parser will assume a default value. These options do not change
// the order ***REMOVED***elds are parse in.
type ParseOption int

const (
	Second      ParseOption = 1 << iota // Seconds ***REMOVED***eld, default 0
	Minute                              // Minutes ***REMOVED***eld, default 0
	Hour                                // Hours ***REMOVED***eld, default 0
	Dom                                 // Day of month ***REMOVED***eld, default *
	Month                               // Month ***REMOVED***eld, default *
	Dow                                 // Day of week ***REMOVED***eld, default *
	DowOptional                         // Optional day of week ***REMOVED***eld, default *
	Descriptor                          // Allow descriptors such as @monthly, @weekly, etc.
)

var places = []ParseOption{
	Second,
	Minute,
	Hour,
	Dom,
	Month,
	Dow,
}

var defaults = []string{
	"0",
	"0",
	"0",
	"*",
	"*",
	"*",
}

// A custom Parser that can be con***REMOVED***gured.
type Parser struct {
	options   ParseOption
	optionals int
}

// Creates a custom Parser with custom options.
//
//  // Standard parser without descriptors
//  specParser := NewParser(Minute | Hour | Dom | Month | Dow)
//  sched, err := specParser.Parse("0 0 15 */3 *")
//
//  // Same as above, just excludes time ***REMOVED***elds
//  subsParser := NewParser(Dom | Month | Dow)
//  sched, err := specParser.Parse("15 */3 *")
//
//  // Same as above, just makes Dow optional
//  subsParser := NewParser(Dom | Month | DowOptional)
//  sched, err := specParser.Parse("15 */3")
//
func NewParser(options ParseOption) Parser {
	optionals := 0
	if options&DowOptional > 0 {
		options |= Dow
		optionals++
	}
	return Parser{options, optionals}
}

// Parse returns a new crontab schedule representing the given spec.
// It returns a descriptive error if the spec is not valid.
// It accepts crontab specs and features con***REMOVED***gured by NewParser.
func (p Parser) Parse(spec string) (Schedule, error) {
	if len(spec) == 0 {
		return nil, fmt.Errorf("Empty spec string")
	}
	if spec[0] == '@' && p.options&Descriptor > 0 {
		return parseDescriptor(spec)
	}

	// Figure out how many ***REMOVED***elds we need
	max := 0
	for _, place := range places {
		if p.options&place > 0 {
			max++
		}
	}
	min := max - p.optionals

	// Split ***REMOVED***elds on whitespace
	***REMOVED***elds := strings.Fields(spec)

	// Validate number of ***REMOVED***elds
	if count := len(***REMOVED***elds); count < min || count > max {
		if min == max {
			return nil, fmt.Errorf("Expected exactly %d ***REMOVED***elds, found %d: %s", min, count, spec)
		}
		return nil, fmt.Errorf("Expected %d to %d ***REMOVED***elds, found %d: %s", min, max, count, spec)
	}

	// Fill in missing ***REMOVED***elds
	***REMOVED***elds = expandFields(***REMOVED***elds, p.options)

	var err error
	***REMOVED***eld := func(***REMOVED***eld string, r bounds) uint64 {
		if err != nil {
			return 0
		}
		var bits uint64
		bits, err = getField(***REMOVED***eld, r)
		return bits
	}

	var (
		second     = ***REMOVED***eld(***REMOVED***elds[0], seconds)
		minute     = ***REMOVED***eld(***REMOVED***elds[1], minutes)
		hour       = ***REMOVED***eld(***REMOVED***elds[2], hours)
		dayofmonth = ***REMOVED***eld(***REMOVED***elds[3], dom)
		month      = ***REMOVED***eld(***REMOVED***elds[4], months)
		dayofweek  = ***REMOVED***eld(***REMOVED***elds[5], dow)
	)
	if err != nil {
		return nil, err
	}

	return &SpecSchedule{
		Second: second,
		Minute: minute,
		Hour:   hour,
		Dom:    dayofmonth,
		Month:  month,
		Dow:    dayofweek,
	}, nil
}

func expandFields(***REMOVED***elds []string, options ParseOption) []string {
	n := 0
	count := len(***REMOVED***elds)
	expFields := make([]string, len(places))
	copy(expFields, defaults)
	for i, place := range places {
		if options&place > 0 {
			expFields[i] = ***REMOVED***elds[n]
			n++
		}
		if n == count {
			break
		}
	}
	return expFields
}

var standardParser = NewParser(
	Minute | Hour | Dom | Month | Dow | Descriptor,
)

// ParseStandard returns a new crontab schedule representing the given standardSpec
// (https://en.wikipedia.org/wiki/Cron). It differs from Parse requiring to always
// pass 5 entries representing: minute, hour, day of month, month and day of week,
// in that order. It returns a descriptive error if the spec is not valid.
//
// It accepts
//   - Standard crontab specs, e.g. "* * * * ?"
//   - Descriptors, e.g. "@midnight", "@every 1h30m"
func ParseStandard(standardSpec string) (Schedule, error) {
	return standardParser.Parse(standardSpec)
}

var defaultParser = NewParser(
	Second | Minute | Hour | Dom | Month | DowOptional | Descriptor,
)

// Parse returns a new crontab schedule representing the given spec.
// It returns a descriptive error if the spec is not valid.
//
// It accepts
//   - Full crontab specs, e.g. "* * * * * ?"
//   - Descriptors, e.g. "@midnight", "@every 1h30m"
func Parse(spec string) (Schedule, error) {
	return defaultParser.Parse(spec)
}

// getField returns an Int with the bits set representing all of the times that
// the ***REMOVED***eld represents or error parsing ***REMOVED***eld value.  A "***REMOVED***eld" is a comma-separated
// list of "ranges".
func getField(***REMOVED***eld string, r bounds) (uint64, error) {
	var bits uint64
	ranges := strings.FieldsFunc(***REMOVED***eld, func(r rune) bool { return r == ',' })
	for _, expr := range ranges {
		bit, err := getRange(expr, r)
		if err != nil {
			return bits, err
		}
		bits |= bit
	}
	return bits, nil
}

// getRange returns the bits indicated by the given expression:
//   number | number "-" number [ "/" number ]
// or error parsing range.
func getRange(expr string, r bounds) (uint64, error) {
	var (
		start, end, step uint
		rangeAndStep     = strings.Split(expr, "/")
		lowAndHigh       = strings.Split(rangeAndStep[0], "-")
		singleDigit      = len(lowAndHigh) == 1
		err              error
	)

	var extra uint64
	if lowAndHigh[0] == "*" || lowAndHigh[0] == "?" {
		start = r.min
		end = r.max
		extra = starBit
	} ***REMOVED*** {
		start, err = parseIntOrName(lowAndHigh[0], r.names)
		if err != nil {
			return 0, err
		}
		switch len(lowAndHigh) {
		case 1:
			end = start
		case 2:
			end, err = parseIntOrName(lowAndHigh[1], r.names)
			if err != nil {
				return 0, err
			}
		default:
			return 0, fmt.Errorf("Too many hyphens: %s", expr)
		}
	}

	switch len(rangeAndStep) {
	case 1:
		step = 1
	case 2:
		step, err = mustParseInt(rangeAndStep[1])
		if err != nil {
			return 0, err
		}

		// Special handling: "N/step" means "N-max/step".
		if singleDigit {
			end = r.max
		}
	default:
		return 0, fmt.Errorf("Too many slashes: %s", expr)
	}

	if start < r.min {
		return 0, fmt.Errorf("Beginning of range (%d) below minimum (%d): %s", start, r.min, expr)
	}
	if end > r.max {
		return 0, fmt.Errorf("End of range (%d) above maximum (%d): %s", end, r.max, expr)
	}
	if start > end {
		return 0, fmt.Errorf("Beginning of range (%d) beyond end of range (%d): %s", start, end, expr)
	}
	if step == 0 {
		return 0, fmt.Errorf("Step of range should be a positive number: %s", expr)
	}

	return getBits(start, end, step) | extra, nil
}

// parseIntOrName returns the (possibly-named) integer contained in expr.
func parseIntOrName(expr string, names map[string]uint) (uint, error) {
	if names != nil {
		if namedInt, ok := names[strings.ToLower(expr)]; ok {
			return namedInt, nil
		}
	}
	return mustParseInt(expr)
}

// mustParseInt parses the given expression as an int or returns an error.
func mustParseInt(expr string) (uint, error) {
	num, err := strconv.Atoi(expr)
	if err != nil {
		return 0, fmt.Errorf("Failed to parse int from %s: %s", expr, err)
	}
	if num < 0 {
		return 0, fmt.Errorf("Negative number (%d) not allowed: %s", num, expr)
	}

	return uint(num), nil
}

// getBits sets all bits in the range [min, max], modulo the given step size.
func getBits(min, max, step uint) uint64 {
	var bits uint64

	// If step is 1, use shifts.
	if step == 1 {
		return ^(math.MaxUint64 << (max + 1)) & (math.MaxUint64 << min)
	}

	// Else, use a simple loop.
	for i := min; i <= max; i += step {
		bits |= 1 << i
	}
	return bits
}

// all returns all bits within the given bounds.  (plus the star bit)
func all(r bounds) uint64 {
	return getBits(r.min, r.max, 1) | starBit
}

// parseDescriptor returns a prede***REMOVED***ned schedule for the expression, or error if none matches.
func parseDescriptor(descriptor string) (Schedule, error) {
	switch descriptor {
	case "@yearly", "@annually":
		return &SpecSchedule{
			Second: 1 << seconds.min,
			Minute: 1 << minutes.min,
			Hour:   1 << hours.min,
			Dom:    1 << dom.min,
			Month:  1 << months.min,
			Dow:    all(dow),
		}, nil

	case "@monthly":
		return &SpecSchedule{
			Second: 1 << seconds.min,
			Minute: 1 << minutes.min,
			Hour:   1 << hours.min,
			Dom:    1 << dom.min,
			Month:  all(months),
			Dow:    all(dow),
		}, nil

	case "@weekly":
		return &SpecSchedule{
			Second: 1 << seconds.min,
			Minute: 1 << minutes.min,
			Hour:   1 << hours.min,
			Dom:    all(dom),
			Month:  all(months),
			Dow:    1 << dow.min,
		}, nil

	case "@daily", "@midnight":
		return &SpecSchedule{
			Second: 1 << seconds.min,
			Minute: 1 << minutes.min,
			Hour:   1 << hours.min,
			Dom:    all(dom),
			Month:  all(months),
			Dow:    all(dow),
		}, nil

	case "@hourly":
		return &SpecSchedule{
			Second: 1 << seconds.min,
			Minute: 1 << minutes.min,
			Hour:   all(hours),
			Dom:    all(dom),
			Month:  all(months),
			Dow:    all(dow),
		}, nil
	}

	const every = "@every "
	if strings.HasPre***REMOVED***x(descriptor, every) {
		duration, err := time.ParseDuration(descriptor[len(every):])
		if err != nil {
			return nil, fmt.Errorf("Failed to parse duration %s: %s", descriptor, err)
		}
		return Every(duration), nil
	}

	return nil, fmt.Errorf("Unrecognized descriptor: %s", descriptor)
}
