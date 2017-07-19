package prestgo

const (
	// This type captures boolean values true and false
	Boolean = "boolean"

	// A 64-bit signed two’s complement integer with a minimum value of -2^63 and a maximum value of 2^63 - 1.
	BigInt = "bigint"

	// Integer assumed to be an alias for BigInt.
	Integer = "integer"

	// A double is a 64-bit inexact, variable-precision implementing the IEEE Standard 754 for Binary Floating-Point Arithmetic.
	Double = "double"

	// Variable length character data.
	VarChar = "varchar"

	// Variable length binary data.
	VarBinary = "varbinary"

	// Variable length json data.
	JSON = "json"

	// Calendar date (year, month, day).
	// Example: DATE '2001-08-22'
	Date = "date"

	// Time of day (hour, minute, second, millisecond) without a time zone. Values of this type are parsed and rendered in the session time zone.
	// Example: TIME '01:02:03.456'
	Time = "time"

	// Instant in time that includes the date and time of day without a time zone. Values of this type are parsed and rendered in the session time zone.
	// Example: TIMESTAMP '2001-08-22 03:04:05.321'
	Timestamp = "timestamp"

	// Instant in time that includes the date and time of day with a time zone. Values of this type are parsed and rendered in the provided time zone.
	// Example: TIMESTAMP '2001-08-22 03:04:05.321' AT TIME ZONE 'America/Los_Angeles'
	TimestampWithTimezone = "timestamp with time zone"
)

type stmtResponse struct {
	ID      string    `json:"id"`
	InfoURI string    `json:"infoUri"`
	NextURI string    `json:"nextUri"`
	Stats   stmtStats `json:"stats"`
	Error   stmtError `json:"error"`
}

type stmtStats struct {
	State           string    `json:"state"`
	Scheduled       bool      `json:"scheduled"`
	Nodes           int       `json:"nodes"`
	TotalSplits     int       `json:"totalSplits"`
	QueuesSplits    int       `json:"queuedSplits"`
	RunningSplits   int       `json:"runningSplits"`
	CompletedSplits int       `json:"completedSplits"`
	UserTimeMillis  int       `json:"userTimeMillis"`
	CPUTimeMillis   int       `json:"cpuTimeMillis"`
	WallTimeMillis  int       `json:"wallTimeMillis"`
	ProcessedRows   int       `json:"processedRows"`
	ProcessedBytes  int       `json:"processedBytes"`
	RootStage       stmtStage `json:"rootStage"`
}

type stmtError struct {
	Message       string               `json:"message"`
	ErrorCode     int                  `json:"errorCode"`
	ErrorLocation stmtErrorLocation    `json:"errorLocation"`
	FailureInfo   stmtErrorFailureInfo `json:"failureInfo"`
	// Other ***REMOVED***elds omitted
}

type stmtErrorLocation struct {
	LineNumber   int `json:"lineNumber"`
	ColumnNumber int `json:"columnNumber"`
}

type stmtErrorFailureInfo struct {
	Type string `json:"type"`
	// Other ***REMOVED***elds omitted
}

func (e stmtError) Error() string {
	return e.FailureInfo.Type + ": " + e.Message
}

type stmtStage struct {
	StageID         string      `json:"stageId"`
	State           string      `json:"state"`
	Done            bool        `json:"done"`
	Nodes           int         `json:"nodes"`
	TotalSplits     int         `json:"totalSplits"`
	QueuedSplits    int         `json:"queuedSplits"`
	RunningSplits   int         `json:"runningSplits"`
	CompletedSplits int         `json:"completedSplits"`
	UserTimeMillis  int         `json:"userTimeMillis"`
	CPUTimeMillis   int         `json:"cpuTimeMillis"`
	WallTimeMillis  int         `json:"wallTimeMillis"`
	ProcessedRows   int         `json:"processedRows"`
	ProcessedBytes  int         `json:"processedBytes"`
	SubStages       []stmtStage `json:"subStages"`
}

type queryResponse struct {
	ID               string        `json:"id"`
	InfoURI          string        `json:"infoUri"`
	PartialCancelURI string        `json:"partialCancelUri"`
	NextURI          string        `json:"nextUri"`
	Columns          []queryColumn `json:"columns"`
	Data             []queryData   `json:"data"`
	Stats            stmtStats     `json:"stats"`
	Error            stmtError     `json:"error"`
}

type queryColumn struct {
	Name          string        `json:"name"`
	Type          string        `json:"type"`
	TypeSignature typeSignature `json:"typeSignature"`
}

type queryData []interface{}

type typeSignature struct {
	RawType          string        `json:"rawType"`
	TypeArguments    []interface{} `json:"typeArguments"`
	LiteralArguments []interface{} `json:"literalArguments"`
}

type infoResponse struct {
	QueryID string `json:"queryId"`
	State   string `json:"state"`
}

const (
	QueryStateQueued   = "QUEUED"
	QueryStatePlanning = "PLANNING"
	QueryStateStarting = "STARTING"
	QueryStateRunning  = "RUNNING"
	QueryStateFinished = "FINISHED"
	QueryStateCanceled = "CANCELED"
	QueryStateFailed   = "FAILED"
)
