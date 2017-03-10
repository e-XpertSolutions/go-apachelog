package apachelog

type Format int

const (
	format_beg            Format = iota
	REMOTE_IP_ADDRESS            // %a
	LOCAL_IP_ADDRESS             // %A
	RESPONSE_SIZE                // %B
	RESPONSE_SIZE_CLF            // %b
	COOKIE                       // %{Foobar}C
	ELAPSED_TIME                 // %D
	ENV_VAR                      // %{FOOBAR}e
	FILENAME                     // %f
	REMOTE_HOST                  // %h
	REQUEST_PROTO                // %H
	REMOTE_LOGNAME               // %l
	REQUEST_METHOD               // %m
	PORT                         // %p
	PROCESS_ID                   // %P
	QUERY_STRING                 // %q
	STATUS                       // %s
	TIME                         // %t
	REMOTE_USER                  // %u
	URL_PATH                     // %U
	CANONICAL_SERVER_NAME        // %v
	SERVER_NAME                  // %V
	BYTES_RECEIVED               // %I
	BYTES_SENT                   // %O
	format_end

	UNKNOWN // for errors
)

var formats = [...]string{
	REMOTE_IP_ADDRESS:     "%a",
	LOCAL_IP_ADDRESS:      "%A",
	RESPONSE_SIZE:         "%B",
	RESPONSE_SIZE_CLF:     "%b",
	COOKIE:                "%{Foobar}C",
	ELAPSED_TIME:          "%D",
	ENV_VAR:               "%{FOOBAR}e",
	FILENAME:              "%f",
	REMOTE_HOST:           "%h",
	REQUEST_PROTO:         "%H",
	REMOTE_LOGNAME:        "%l",
	REQUEST_METHOD:        "%m",
	PORT:                  "%p",
	PROCESS_ID:            "%P",
	QUERY_STRING:          "%q",
	STATUS:                "%s",
	TIME:                  "%t",
	REMOTE_USER:           "%u",
	URL_PATH:              "%U",
	CANONICAL_SERVER_NAME: "%v",
	SERVER_NAME:           "%V",
	BYTES_RECEIVED:        "%I",
	BYTES_SENT:            "%O",
	UNKNOWN:               "UNKNOWN",
}

func (f Format) String() string {
	if f > format_beg && f < format_end {
		return formats[f+1]
	}
	return formats[UNKNOWN]
}

var formatsMapping map[string]Format

func init() {
	formatsMapping = make(map[string]Format)
	for i := format_beg + 1; i < format_end; i++ {
		formatsMapping[formats[i]] = i
	}
}

func LookupFormat(format string) Format {
	if f, found := formatsMapping[format]; found {
		return f
	}
	return UNKNOWN
}
