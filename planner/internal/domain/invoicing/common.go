package invoicing

import "time"

// DaysInMonth returns the number of days in the specified year and month.
func DaysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// MonthNameInPortuguese returns the lowercase Portuguese name for the given month.
func MonthNameInPortuguese(m time.Month) string {
	switch m {
	case time.January:
		return "janeiro"
	case time.February:
		return "fevereiro"
	case time.March:
		return "março"
	case time.April:
		return "abril"
	case time.May:
		return "maio"
	case time.June:
		return "junho"
	case time.July:
		return "julho"
	case time.August:
		return "agosto"
	case time.September:
		return "setembro"
	case time.October:
		return "outubro"
	case time.November:
		return "novembro"
	case time.December:
		return "dezembro"
	default:
		return ""
	}
}

