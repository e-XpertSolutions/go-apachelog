package apachelog

// Format supported by the Apache mod_log_config module.
// For more information, see:
//    https://httpd.apache.org/docs/2.2/fr/mod/mod_log_config.html#formats
type Format int

// Supported formats.
//
// TODO(gilliek): move complex format, such as COOKIE, at the bottom of the list
// in order to treat them separately.
const (
	format_beg            Format = iota
	REMOTE_IP_ADDRESS            // %a
	LOCAL_IP_ADDRESS             // %A
	RESPONSE_SIZE                // %B
	RESPONSE_SIZE_CLF            // %b
	COOKIE                       // %{Foobar}C
	ELAPSED_TIME                 // %D
	ENV_VAR                      // %{FOOBAR}e
	HEADER                       // %{Foobar}i
	FILENAME                     // %f
	REMOTE_HOST                  // %h
	REQUEST_PROTO                // %H
	REMOTE_LOGNAME               // %l
	REQUEST_METHOD               // %m
	PORT                         // %p
	PROCESS_ID                   // %P
	QUERY_STRING                 // %q
	REQUEST_FIRST_LINE           // %r
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
	COOKIE:                "%{...}C",
	ELAPSED_TIME:          "%D",
	ENV_VAR:               "%{...}e",
	HEADER:                "%{...}i",
	FILENAME:              "%f",
	REMOTE_HOST:           "%h",
	REQUEST_PROTO:         "%H",
	REMOTE_LOGNAME:        "%l",
	REQUEST_METHOD:        "%m",
	PORT:                  "%p",
	PROCESS_ID:            "%P",
	QUERY_STRING:          "%q",
	REQUEST_FIRST_LINE:    "%r",
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

// LookupFormat retrieves the format corresponding to the given format string.
func LookupFormat(format string) Format {
	if f, found := formatsMapping[format]; found {
		return f
	}
	return UNKNOWN
}
