package date

type DateFormat string

const (
	FORMAT_DATE        DateFormat = "2006-01-02"
	FORMAT_TIME        DateFormat = "15:04:05"
	FORMAT_DATETIME    DateFormat = "2006-01-02 15:04:05"
	FORMAT_DATETIME_MS DateFormat = "2006-01-02 15:04:05.000"
	FORMAT_ISO8601     DateFormat = "2006-01-02T15:04:05Z"
	FORMAT_ISO8601_MS  DateFormat = "2006-01-02T15:04:05.000Z"
	FORMAT_RFC2822     DateFormat = "Mon, 02 Jan 2006 15:04:05 -0700"
)
