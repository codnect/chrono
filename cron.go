package chrono

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

var (
	months = []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}
	days   = []string{"MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"}
)

type cronField string

const (
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

func (fieldType fieldType) checkValidValue(value string) int {
	result, err := strconv.Atoi(value)

	if err != nil {
		panic(err)
	}

	if fieldType.Field == cronFieldDayOfWeek && result == 0 {
		return result
	}

	if result >= fieldType.MinValue && result <= fieldType.MaxValue {
		return result
	}

	panic("value is not valid")
}

var (
	second     = fieldType{cronFieldSecond, 0, 59}
	minute     = fieldType{cronFieldMinute, 0, 59}
	hour       = fieldType{cronFieldHour, 0, 23}
	dayOfMonth = fieldType{cronFieldDayOfMonth, 1, 31}
	month      = fieldType{cronFieldMonth, 1, 12}
	dayOfWeek  = fieldType{cronFieldDayOfWeek, 0, 6}
)

var cronFields = []fieldType{
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
	typ  fieldType
	bits uint64
}

func newFieldBits(typ fieldType) *cronFieldBits {
	return &cronFieldBits{
		typ: typ,
	}
}

func (bits *cronFieldBits) setBits(valueRange valueRange) {
	if valueRange.MinValue == valueRange.MaxValue {
		bits.bits |= 1 << valueRange.MinValue
	} else {
		bits.bits = ^(math.MaxUint64 << (valueRange.MaxValue + 1)) & (math.MaxUint64 << valueRange.MinValue)
	}
}

func (bits *cronFieldBits) setBitsWithStep(valueRange valueRange, step int) {
	if step == 1 {
		bits.setBits(valueRange)
	} else {
		for index := valueRange.MinValue; index <= valueRange.MaxValue; index += step {
			bits.bits |= 1 << index
		}
	}
}

type CronExpression struct {
	fields []*cronFieldBits
}

func newCronExpression() *CronExpression {
	return &CronExpression{
		make([]*cronFieldBits, 0),
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

	for index, cronField := range cronFields {
		value, err := parseField(fields[index], cronField)

		if err != nil {
			return nil, err
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

		if slashPos != -1 {
			rangeStr := field[0:slashPos]

			valueRange := parseRange(rangeStr, fieldType)

			if strings.Index(rangeStr, "-") == -1 {
				valueRange = newValueRange(valueRange.MinValue, fieldType.MaxValue)
			}

			stepStr := field[slashPos+1:]

			step, err := strconv.Atoi(stepStr)

			if err != nil {
				panic(err)
			}

			if step <= 0 {
				panic("step must be 1 or higher")
			}

			cronFieldBits.setBitsWithStep(valueRange, step)
		} else {
			valueRange := parseRange(field, fieldType)
			cronFieldBits.setBits(valueRange)
		}
	}

	return cronFieldBits, nil
}

func parseRange(value string, fieldType fieldType) valueRange {
	if value == "*" {
		return newValueRange(fieldType.MinValue, fieldType.MaxValue)
	} else {
		hyphenPos := strings.Index(value, "-")

		if hyphenPos == -1 {
			result := fieldType.checkValidValue(value)

			return newValueRange(result, result)
		} else {
			maxStr := value[hyphenPos+1:]
			minStr := value[0:hyphenPos]

			min := fieldType.checkValidValue(minStr)
			max := fieldType.checkValidValue(maxStr)

			if fieldType.Field == cronFieldDayOfWeek && min == 7 {
				min = 0
			}

			return newValueRange(min, max)
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
