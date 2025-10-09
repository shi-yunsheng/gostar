package date

type DateFormat string

const (
	FORMAT_DATE        DateFormat = "2006-01-02"
	FORMAT_TIME        DateFormat = "15:04:05"
	FORMAT_DATETIME    DateFormat = "2006-01-02 15:04:05"
	FORMAT_DATETIME_MS DateFormat = "2006-01-02 15:04:05.000"
)
