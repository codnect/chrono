package chrono

import (
	"errors"
	"fmt"
	"math"
	"math/bits"
	"strconv"
	"strings"
	"time"
)

var (
	months = []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
	days   = []string{"MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"}
)

type cronField string

const (
	cronFieldNanoSecond = "NANO_SECOND"
	cronFieldSecond     = "SECOND"
	cronFieldMinute     = "MINUTE"
	cronFieldHour       = "HOUR"
	cronFieldDayOfMonth = "DAY_OF_MONTH"
	cronFieldMonth      = "MONTH"
	cronFieldDayOfWeek  = "DAY_OF_WEEK"
)

type fieldType struct {
	Field    cronField
	MinValue int
	MaxValue int
}

var (
	nanoSecond = fieldType{cronFieldNanoSecond, 0, 999999999}
	second     = fieldType{cronFieldSecond, 0, 59}
	minute     = fieldType{cronFieldMinute, 0, 59}
	hour       = fieldType{cronFieldHour, 0, 23}
	dayOfMonth = fieldType{cronFieldDayOfMonth, 1, 31}
	month      = fieldType{cronFieldMonth, 1, 12}
	dayOfWeek  = fieldType{cronFieldDayOfWeek, 1, 7}
)

var cronFieldTypes = []fieldType{
	second,
	minute,
	hour,
	dayOfMonth,
	month,
	dayOfWeek,
}

type valueRange struct {
	MinValue int
	MaxValue int
}

func newValueRange(min int, max int) valueRange {
	return valueRange{
		MinValue: min,
		MaxValue: max,
	}
}

type cronFieldBits struct {
	Typ  fieldType
	Bits uint64
}

func newFieldBits(typ fieldType) *cronFieldBits {
	return &cronFieldBits{
		Typ: typ,
	}
}

const maxAttempts = 366
const mask = 0xFFFFFFFFFFFFFFFF

type CronExpression struct {
	fields []*cronFieldBits
}

func newCronExpression() *CronExpression {
	exp := &CronExpression{
		make([]*cronFieldBits, 0),
	}

	nanoSecondBits := newFieldBits(nanoSecond)
	nanoSecondBits.Bits = 1

	exp.fields = append(exp.fields, nanoSecondBits)
	return exp
}

func (expression *CronExpression) NextTime(t time.Time) time.Time {

	t = t.Add(1 * time.Nanosecond)

	for i := 0; i < maxAttempts; i++ {
		result := expression.next(t)

		if result.IsZero() || result.Equal(t) {
			return result
		}

		t = result
	}

	return time.Time{}
}

func (expression *CronExpression) next(t time.Time) time.Time {
	for _, field := range expression.fields {
		t = expression.nextField(field, t)

		if t.IsZero() {
			return t
		}
	}

	return t
}

func (expression *CronExpression) nextField(field *cronFieldBits, t time.Time) time.Time {
	current := getTimeValue(t, field.Typ.Field)
	next := setNextBit(field.Bits, current)

	if next == -1 {
		amount := getFieldMaxValue(t, field.Typ) - current + 1
		t = addTime(t, field.Typ.Field, amount)
		next = setNextBit(field.Bits, 0)
	}

	if next == current {
		return t
	} else {
		count := 0
		current := getTimeValue(t, field.Typ.Field)
		for ; current != next && count < maxAttempts; count++ {
			t = elapseUntil(t, field.Typ, next)
			current = getTimeValue(t, field.Typ.Field)
		}

		if count >= maxAttempts {
			return time.Time{}
		}

		return t
	}
}

func ParseCronExpression(expression string) (*CronExpression, error) {
	if len(expression) == 0 {
		return nil, errors.New("cron expression must not be empty")
	}

	fields := strings.Fields(expression)

	if len(fields) != 6 {
		return nil, fmt.Errorf("cron expression must consist of 6 fields : found %d in \"%s\"", len(fields), expression)
	}

	cronExpression := newCronExpression()

	for index, cronFieldType := range cronFieldTypes {
		value, err := parseField(fields[index], cronFieldType)

		if err != nil {
			return nil, err
		}

		if cronFieldType.Field == cronFieldDayOfWeek && value.Bits&1<<0 != 0 {
			value.Bits |= 1 << 7
			temp := ^(1 << 0)
			value.Bits &= uint64(temp)
		}

		cronExpression.fields = append(cronExpression.fields, value)
	}

	return cronExpression, nil
}

func parseField(value string, fieldType fieldType) (*cronFieldBits, error) {
	if len(value) == 0 {
		return nil, fmt.Errorf("value must not be empty")
	}

	if fieldType.Field == cronFieldMonth {
		value = replaceOrdinals(value, months)
	} else if fieldType.Field == cronFieldDayOfWeek {
		value = replaceOrdinals(value, days)
	}

	cronFieldBits := newFieldBits(fieldType)

	fields := strings.Split(value, ",")

	for _, field := range fields {
		slashPos := strings.Index(field, "/")

		step := -1
		var valueRange valueRange

		if slashPos != -1 {
			rangeStr := field[0:slashPos]

			var err error
			valueRange, err = parseRange(rangeStr, fieldType)

			if err != nil {
				return nil, err
			}

			if strings.Index(rangeStr, "-") == -1 {
				valueRange = newValueRange(valueRange.MinValue, fieldType.MaxValue)
			}

			stepStr := field[slashPos+1:]

			step, err = strconv.Atoi(stepStr)

			if err != nil {
				return nil, fmt.Errorf("step must be number : \"%s\"", stepStr)
			}

			if step <= 0 {
				return nil, fmt.Errorf("step must be 1 or higher in \"%s\"", value)
			}

		} else {
			var err error
			valueRange, err = parseRange(field, fieldType)

			if err != nil {
				return nil, err
			}
		}

		if step > 1 {
			for index := valueRange.MinValue; index <= valueRange.MaxValue; index += step {
				cronFieldBits.Bits |= 1 << index
			}
			continue
		}

		if valueRange.MinValue == valueRange.MaxValue {
			cronFieldBits.Bits |= 1 << valueRange.MinValue
		} else {
			cronFieldBits.Bits |= ^(math.MaxUint64 << (valueRange.MaxValue + 1)) & (math.MaxUint64 << valueRange.MinValue)
		}
	}

	return cronFieldBits, nil
}

func parseRange(value string, fieldType fieldType) (valueRange, error) {
	if value == "*" {
		return newValueRange(fieldType.MinValue, fieldType.MaxValue), nil
	} else {
		hyphenPos := strings.Index(value, "-")

		if hyphenPos == -1 {
			result, err := checkValidValue(value, fieldType)

			if err != nil {
				return valueRange{}, err
			}

			return newValueRange(result, result), nil
		} else {
			maxStr := value[hyphenPos+1:]
			minStr := value[0:hyphenPos]

			min, err := checkValidValue(minStr, fieldType)

			if err != nil {
				return valueRange{}, err
			}
			var max int
			max, err = checkValidValue(maxStr, fieldType)

			if err != nil {
				return valueRange{}, err
			}

			if fieldType.Field == cronFieldDayOfWeek && min == 7 {
				min = 0
			}

			return newValueRange(min, max), nil
		}
	}
}

func replaceOrdinals(value string, list []string) string {
	value = strings.ToUpper(value)

	for index := 0; index < len(list); index++ {
		replacement := strconv.Itoa(index + 1)
		value = strings.ReplaceAll(value, list[index], replacement)
	}

	return value
}

func checkValidValue(value string, fieldType fieldType) (int, error) {
	result, err := strconv.Atoi(value)

	if err != nil {
		return 0, fmt.Errorf("the value in field %s must be number : %s", fieldType.Field, value)
	}

	if fieldType.Field == cronFieldDayOfWeek && result == 0 {
		return result, nil
	}

	if result >= fieldType.MinValue && result <= fieldType.MaxValue {
		return result, nil
	}

	return 0, fmt.Errorf("the value in field %s must be between %d and %d", fieldType.Field, fieldType.MinValue, fieldType.MaxValue)
}

func getTimeValue(t time.Time, field cronField) int {

	switch field {
	case cronFieldNanoSecond:
		return t.Nanosecond()
	case cronFieldSecond:
		return t.Second()
	case cronFieldMinute:
		return t.Minute()
	case cronFieldHour:
		return t.Hour()
	case cronFieldDayOfMonth:
		return t.Day()
	case cronFieldMonth:
		return int(t.Month())
	case cronFieldDayOfWeek:
		if t.Weekday() == 0 {
			return 7
		}
		return int(t.Weekday())
	}

	panic("unreachable code!")
}

func addTime(t time.Time, field cronField, value int) time.Time {
	switch field {
	case cronFieldNanoSecond:
		return t.Add(time.Duration(value) * time.Nanosecond)
	case cronFieldSecond:
		return t.Add(time.Duration(value) * time.Second)
	case cronFieldMinute:
		return t.Add(time.Duration(value) * time.Minute)
	case cronFieldHour:
		return t.Add(time.Duration(value) * time.Hour)
	case cronFieldDayOfMonth:
		return t.AddDate(0, 0, value)
	case cronFieldMonth:
		return t.AddDate(0, value, 0)
	case cronFieldDayOfWeek:
		return t.AddDate(0, 0, value)
	}

	panic("unreachable code!")
}

func setNextBit(bitsValue uint64, index int) int {
	result := bitsValue & (mask << index)

	if result != 0 {
		return bits.TrailingZeros64(result)
	}

	return -1
}

func elapseUntil(t time.Time, fieldType fieldType, value int) time.Time {
	current := getTimeValue(t, fieldType.Field)

	maxValue := getFieldMaxValue(t, fieldType)

	if current >= value {
		amount := value + maxValue - current + 1 - fieldType.MinValue
		return addTime(t, fieldType.Field, amount)
	}

	if value >= fieldType.MinValue && value <= maxValue {
		return with(t, fieldType.Field, value)
	}

	return addTime(t, fieldType.Field, value-current)
}

func with(t time.Time, field cronField, value int) time.Time {
	switch field {
	case cronFieldNanoSecond:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), value, time.Local)
	case cronFieldSecond:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), value, t.Nanosecond(), time.Local)
	case cronFieldMinute:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), value, t.Second(), t.Nanosecond(), time.Local)
	case cronFieldHour:
		return time.Date(t.Year(), t.Month(), t.Day(), value, t.Minute(), t.Second(), t.Nanosecond(), time.Local)
	case cronFieldDayOfMonth:
		return time.Date(t.Year(), t.Month(), value, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local)
	case cronFieldMonth:
		return time.Date(t.Year(), time.Month(value), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local)
	case cronFieldDayOfWeek:
		return t.AddDate(0, 0, value-int(t.Weekday()))
	}

	panic("unreachable code!")
}

func getFieldMaxValue(t time.Time, fieldType fieldType) int {

	if cronFieldDayOfMonth == fieldType.Field {
		switch int(t.Month()) {
		case 2:
			if isLeapYear(t.Year()) {
				return 29
			}
			return 28
		case 4:
			return 30
		case 6:
			return 30
		case 9:
			return 30
		case 11:
			return 30
		default:
			return 31
		}
	}

	return fieldType.MaxValue
}

func isLeapYear(year int) bool {
	return year%400 == 0 || year%100 != 0 && year%4 == 0
}
