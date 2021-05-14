package chrono

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

/**
 * Cron and CronParser make use of algorithm and approach of https://github.com/robfig/cron.
 */

const cronExpressionFieldCount = 6
const valueListSeparator = ","
const rangeStepSeparator = "/"
const rangeSeparator = "-"
const asterisk = "*"
const questionMark = "?"

const CronEveryFlag = 1 << 63

type CronField struct {
	Index  uint
	Min    uint
	Max    uint
	Values map[string]uint
}

var (
	Second     = CronField{0, 0, 59, nil}
	Minute     = CronField{1, 0, 59, nil}
	Hour       = CronField{2, 0, 23, nil}
	DayOfMonth = CronField{3, 1, 31, nil}
	Month      = CronField{4, 1, 12, map[string]uint{
		"JAN": 1,
		"FEB": 2,
		"MAR": 3,
		"APR": 4,
		"MAY": 5,
		"JUN": 6,
		"JUL": 7,
		"AUG": 8,
		"SEP": 9,
		"OCT": 10,
		"NOV": 11,
		"DEC": 12,
	}}
	DayOfWeek = CronField{5, 0, 6, map[string]uint{
		"SUN": 0,
		"MON": 1,
		"TUE": 2,
		"WED": 3,
		"THU": 4,
		"FRI": 5,
		"SAT": 6,
	}}
)

var cronFields = []CronField{
	Second,
	Minute,
	Hour,
	DayOfMonth,
	Month,
	DayOfWeek,
}

type Cron struct {
	Second     uint64
	Minute     uint64
	Hour       uint64
	DayOfMonth uint64
	Month      uint64
	DayOfWeek  uint64
	Location   *time.Location
}

func NewCron() *Cron {
	return &Cron{
		Location: time.Local,
	}
}

func (cron *Cron) Set(cronField CronField, value uint64) {

	switch cronField.Index {
	case Second.Index:
		cron.Second = value
	case Minute.Index:
		cron.Minute = value
	case Hour.Index:
		cron.Hour = value
	case DayOfMonth.Index:
		cron.DayOfMonth = value
	case Month.Index:
		cron.Month = value
	case DayOfWeek.Index:
		cron.DayOfWeek = value
	}

}

func (cron *Cron) NextTime(t time.Time) time.Time {
	location := cron.Location

	if location == nil {
		location = time.Local
	}

	if location == time.Local {
		location = t.Location()
	} else {
		t = t.In(cron.Location)
	}

	originLocation := t.Location()

	t = t.Add(1*time.Second - time.Duration(t.Nanosecond())*time.Nanosecond)

	added := false

repeat:

	for 1<<uint(t.Month())&cron.Month == 0 {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, location)
		}

		t = t.AddDate(0, 1, 0)

		if t.Month() == time.January {
			goto repeat
		}
	}

	for !cron.matches(t) {

		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, location)
		}

		t = t.AddDate(0, 0, 1)

		if t.Hour() != 0 {
			if t.Hour() > 12 {
				t = t.Add(time.Duration(24-t.Hour()) * time.Hour)
			} else {
				t = t.Add(time.Duration(-t.Hour()) * time.Hour)
			}
		}

		if t.Day() == 1 {
			goto repeat
		}
	}

	for 1<<uint(t.Hour())&cron.Hour == 0 {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, location)
		}

		t = t.Add(1 * time.Hour)

		if t.Hour() == 0 {
			goto repeat
		}
	}

	for 1<<uint(t.Minute())&cron.Minute == 0 {
		if !added {
			added = true
			t = t.Truncate(time.Minute)
		}

		t = t.Add(1 * time.Minute)

		if t.Minute() == 0 {
			goto repeat
		}
	}

	for 1<<uint(t.Second())&cron.Second == 0 {
		if !added {
			added = true
			t = t.Truncate(time.Second)
		}

		t = t.Add(1 * time.Second)

		if t.Second() == 0 {
			goto repeat
		}
	}

	return t.In(originLocation)
}

func (cron *Cron) matches(t time.Time) bool {

	dayOfMonthMatch := 1<<uint(t.Day())&cron.DayOfMonth > 0
	dayOfWeekMatch := 1<<uint(t.Weekday())&cron.DayOfWeek > 0

	if cron.DayOfMonth&(CronEveryFlag) > 0 || cron.DayOfWeek&(CronEveryFlag) > 0 {
		return dayOfMonthMatch && dayOfWeekMatch
	}

	return dayOfMonthMatch || dayOfWeekMatch
}

var cronParser = NewCronParser()

type CronParser struct {
}

func NewCronParser() *CronParser {
	return &CronParser{}
}

func (parser *CronParser) Parse(expression string) (*Cron, error) {

	if len(expression) == 0 {
		return nil, errors.New("cron expression must not be empty")
	}

	fields := strings.Fields(expression)

	if len(fields) != cronExpressionFieldCount {
		return nil, fmt.Errorf("expected %d cron expression fields, found %d: %s", cronExpressionFieldCount, len(fields), fields)
	}

	var err error
	cron := NewCron()

	for _, cronField := range cronFields {

		var value uint64
		value, err = parser.parseField(cronField, fields[cronField.Index])

		if err != nil {
			return nil, fmt.Errorf("cron expression is not valid : %s, message : %s", expression, err.Error())
		}

		cron.Set(cronField, value)
	}

	return cron, nil
}

func (parser *CronParser) parseField(cronField CronField, fieldExpression string) (uint64, error) {

	var fieldValue uint64
	values := strings.Split(fieldExpression, valueListSeparator)

	for _, value := range values {

		bits, err := cronParser.getBitsValue(cronField, value)

		if err != nil {
			return fieldValue, err
		}

		fieldValue |= bits
	}

	return fieldValue, nil
}

func (parser *CronParser) getBitsValue(cronField CronField, value string) (uint64, error) {

	var err error
	var min, max, step uint

	rangeAndStep := strings.Split(value, rangeStepSeparator)

	if len(rangeAndStep) > 2 {
		return 0, fmt.Errorf("range-step format is wrong %s", value)
	}

	minAndMax := strings.Split(rangeAndStep[0], rangeSeparator)

	if len(minAndMax) > 2 {
		return 0, fmt.Errorf("range format is wrong %s", value)
	}

	var everyFlag uint64

	if minAndMax[0] == asterisk || minAndMax[0] == questionMark {
		min = cronField.Min
		max = cronField.Max
		everyFlag = CronEveryFlag
	} else {
		min, max, err = cronParser.getRanges(minAndMax, cronField.Values)

		if err != nil {
			return 0, err
		}
	}

	switch len(rangeAndStep) {
	case 1:
		step = 1
	case 2:
		step, err = parser.getActualValue(rangeAndStep[1], nil)

		if err != nil {
			return 0, err
		}

		if len(minAndMax) == 1 {
			max = cronField.Max
		}
	}

	if step > 1 {
		everyFlag = 0
	}

	if min < cronField.Min || max > cronField.Max {
		return 0, fmt.Errorf("value out of range for %s, expected: (%d-%d) actual:(%d-%d)", value,
			cronField.Min, cronField.Max, min, max)
	}

	if min > max {
		return 0, fmt.Errorf("min value (%d) cannot be greater than max value (%d) : %s", min, max, value)
	}

	var bits uint64

	if step == 1 {
		bits = ^(math.MaxUint64 << (max + 1)) & (math.MaxUint64 << min)
	}

	for index := min; index <= max; index += step {
		bits |= 1 << index
	}

	return bits | everyFlag, nil
}

func (parser *CronParser) getRanges(minAndMax []string, values map[string]uint) (uint, uint, error) {
	var err error
	var min, max uint

	min, err = parser.getActualValue(minAndMax[0], values)

	if err != nil {
		return 0, 0, err
	}

	switch len(minAndMax) {
	case 1:
		max = min
	case 2:
		max, err = parser.getActualValue(minAndMax[1], values)
	}

	return min, max, err
}

func (parser *CronParser) getActualValue(expression string, values map[string]uint) (uint, error) {

	if values != nil {

		if actualValue, ok := values[strings.ToUpper(expression)]; ok {
			return actualValue, nil

		}
	}

	actualValue, err := strconv.Atoi(expression)

	if err != nil {
		return 0, err
	}

	if actualValue < 0 {
		return 0, fmt.Errorf("negative value (%d) is not allowed: %s", actualValue, expression)
	}

	return uint(actualValue), nil
}
