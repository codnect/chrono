package chrono

import "time"

type TimeFunction func() time.Time

func Time(year int, month time.Month, day, hour, min, sec, nsec int) TimeFunction {
	return func() time.Time {
		return time.Date(year, month, day, hour, min, sec, nsec, time.Local)
	}
}
