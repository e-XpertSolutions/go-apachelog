package apachelog

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// Most commonly used log formats.
const (
	CombinedLogFromat = "%h %l %u %t \"%r\" %s %b \"%{Referer}i\" \"%{User-agent}i\""
	CommonLogFormat   = "%h %l %u %t \"%r\" %s %b"
)

// StandardEnglishFormat is the time layout to use for parsing %t format.
const StandardEnglishFormat = "02/Jan/2006:15:04:05 -0700"

// A stateFn extracts an information contained in "line" and starting from
// "pos". The function is supposed to update the entry with the extracted
// information.
type stateFn func(entry *AccessLogEntry, line string, pos int) error

// makeStateFn constructs a chain of state functions from an expression, which
// corresponds to a list of format strings as defined by the Apache
// mod_log_config module documentation:
//    https://httpd.apache.org/docs/2.2/fr/mod/mod_log_config.html#formats
func makeStateFn(expr []string) (stateFn, error) {
	// End of the recursive call, we return nil.
	if expr == nil || len(expr) == 0 {
		return nil, nil
	}

	formatStr := expr[0]

	// Expressions can be quoted, so we keep a track of it and trim the quotes.
	var quoted bool
	if strings.HasPrefix(formatStr, "\"") {
		quoted = true
		formatStr = strings.Trim(formatStr, "\"")
	}

	// Recursive call to determine the next state function.
	// XXX(gilliek): errors are reported right to left
	next, err := makeStateFn(expr[1:])
	if err != nil {
		return nil, err
	}

	switch f := LookupFormat(formatStr); f {
	case REMOTE_HOST:
		return parseRemoteHost(quoted, next), nil
	case REMOTE_LOGNAME:
		return parseRemoteLogname(quoted, next), nil
	case REMOTE_USER:
		return parseRemoteUser(quoted, next), nil
	case TIME:
		return parseTime(quoted, next), nil
	case REQUEST_FIRST_LINE:
		return parseRequestFirstLine(quoted, next), nil
	case STATUS:
		return parseStatus(quoted, next), nil
	case RESPONSE_SIZE:
		return parseResponseSize(quoted, next), nil
	case RESPONSE_SIZE_CLF:
		return parseResponseSizeCLF(quoted, next), nil
	case ELAPSED_TIME_IN_SEC:
		return parseElapsedTimeInSec(quoted, next), nil
	case HEADER:
		hdr := strings.TrimSuffix(strings.TrimPrefix(formatStr, "%{"), "}i")
		return parseHeader(quoted, hdr, next), nil
	case UNKNOWN:
		fallthrough
	default:
		return nil, fmt.Errorf("%q format is not supported", formatStr)
	}
}

// A Parser for parsing Apaache access log files.
type Parser struct {
	br *bufio.Reader
	fn stateFn
}

// CombinedParser creates a new parser that reads from r and that parses log
// entries using the Apache Combined Log format.
func CombinedParser(r io.Reader) (*Parser, error) {
	return CustomParser(r, CombinedLogFromat)
}

// CommonParser creates a new parser that reads from r and that parses log entries
// using the Apache Common Log format.
func CommonParser(r io.Reader) (*Parser, error) {
	return CustomParser(r, CommonLogFormat)
}

// CustomParser creates a new parser that reads from r and that is capable of
// parsing log entries having the given format.
//
// The format is mostly the same as the one defined by the mod_log_config Apache
// module:
//    https://httpd.apache.org/docs/2.2/fr/mod/mod_log_config.html#formats
//
// However, unlike the Apache format, it does not support modifiers of any
// kind (<, >, !, ...).
func CustomParser(r io.Reader, format string) (*Parser, error) {
	if r == nil {
		return nil, errors.New("reader is nil")
	}
	fn, err := makeStateFn(strings.Split(format, " "))
	if err != nil {
		return nil, err
	}
	return &Parser{
		br: bufio.NewReader(r),
		fn: fn,
	}, nil
}

// Parse the next access log entry. If there is no more data to read and parse,
// an io.EOF error is returned.
func (p *Parser) Parse() (*AccessLogEntry, error) {
	line, err := p.br.ReadString('\n')
	if err != nil {
		return nil, err
	}
	entry := AccessLogEntry{
		Cookies: make(map[string]string),
		Headers: make(map[string]string),
		EnvVars: make(map[string]string),
	}
	if err := p.fn(&entry, line, 0); err != nil {
		return nil, err
	}
	return &entry, nil
}

func parseRemoteHost(quoted bool, next stateFn) stateFn {
	return func(entry *AccessLogEntry, line string, pos int) error {
		data, off, err := readString(line, pos, quoted)
		if err != nil {
			return err
		}
		entry.RemoteHost = data
		newPos := pos + off
		if line[newPos] == ' ' {
			newPos++ // jump over next space, if any
		}
		if line[newPos] == '\n' || next == nil {
			// If we reached the final \n character or that there is no further
			// state, we do not call the next function.
			return nil
		}
		return next(entry, line, newPos)
	}
}

func parseRemoteLogname(quoted bool, next stateFn) stateFn {
	return func(entry *AccessLogEntry, line string, pos int) error {
		data, off, err := readString(line, pos, quoted)
		if err != nil {
			return err
		}
		entry.RemoteLogname = data
		newPos := pos + off
		if line[newPos] == ' ' {
			newPos++ // jump over next space, if any
		}
		if line[newPos] == '\n' || next == nil {
			// If we reached the final \n character or that there is no further
			// state, we do not call the next function.
			return nil
		}
		return next(entry, line, newPos)
	}
}

func parseRemoteUser(quoted bool, next stateFn) stateFn {
	return func(entry *AccessLogEntry, line string, pos int) error {
		data, off, err := readString(line, pos, quoted)
		if err != nil {
			return err
		}
		entry.RemoteUser = data
		newPos := pos + off
		if line[newPos] == ' ' {
			newPos++ // jump over next space, if any
		}
		if line[newPos] == '\n' || next == nil {
			// If we reached the final \n character or that there is no further
			// state, we do not call the next function.
			return nil
		}
		return next(entry, line, newPos)
	}
}

func parseTime(quoted bool, next stateFn) stateFn {
	return func(entry *AccessLogEntry, line string, pos int) error {
		data, off, err := readDateTime(line, pos, quoted)
		if err != nil {
			return err
		}
		entry.Time = data
		newPos := pos + off
		if line[newPos] == ' ' {
			newPos++ // jump over next space, if any
		}
		if line[newPos] == '\n' || next == nil {
			// If we reached the final \n character or that there is no further
			// state, we do not call the next function.
			return nil
		}
		return next(entry, line, newPos)
	}
}

func parseRequestFirstLine(quoted bool, next stateFn) stateFn {
	return func(entry *AccessLogEntry, line string, pos int) error {
		data, off, err := readString(line, pos, quoted)
		if err != nil {
			return err
		}
		entry.RequestFirstLine = NewRequestFirstLine(data)
		newPos := pos + off
		if line[newPos] == ' ' {
			newPos++ // jump over next space, if any
		}
		if line[newPos] == '\n' || next == nil {
			// If we reached the final \n character or that there is no further
			// state, we do not call the next function.
			return nil
		}
		return next(entry, line, newPos)
	}
}

func parseStatus(quoted bool, next stateFn) stateFn {
	return func(entry *AccessLogEntry, line string, pos int) error {
		data, off, err := readString(line, pos, quoted)
		if err != nil {
			return err
		}
		entry.Status = data
		newPos := pos + off
		if line[newPos] == ' ' {
			newPos++ // jump over next space, if any
		}
		if line[newPos] == '\n' || next == nil {
			// If we reached the final \n character or that there is no further
			// state, we do not call the next function.
			return nil
		}
		return next(entry, line, newPos)
	}
}

func parseResponseSize(quoted bool, next stateFn) stateFn {
	return func(entry *AccessLogEntry, line string, pos int) error {
		data, off, err := readInt(line, pos, quoted)
		if err != nil {
			return err
		}
		entry.ResponseSize = data
		newPos := pos + off
		if line[newPos] == ' ' {
			newPos++ // jump over next space, if any
		}
		if line[newPos] == '\n' || next == nil {
			// If we reached the final \n character or that there is no further
			// state, we do not call the next function.
			return nil
		}
		return next(entry, line, newPos)
	}
}

func parseResponseSizeCLF(quoted bool, next stateFn) stateFn {
	return func(entry *AccessLogEntry, line string, pos int) error {
		data, off, err := readString(line, pos, quoted)
		if err != nil {
			return err
		}
		if data != "-" {
			if entry.ResponseSize, err = strconv.ParseInt(data, 10, 64); err != nil {
				return errors.New("malformed response size: " + err.Error())
			}
		}
		newPos := pos + off
		if line[newPos] == ' ' {
			newPos++ // jump over next space, if any
		}
		if line[newPos] == '\n' || next == nil {
			// If we reached the final \n character or that there is no further
			// state, we do not call the next function.
			return nil
		}
		return next(entry, line, newPos)
	}
}

func parseElapsedTimeInSec(quoted bool, next stateFn) stateFn {
	return func(entry *AccessLogEntry, line string, pos int) error {
		data, off, err := readInt(line, pos, quoted)
		if err != nil {
			return err
		}
		entry.ElapsedTimeSec = data
		newPos := pos + off
		if line[newPos] == ' ' {
			newPos++ // jump over next space, if any
		}
		if line[newPos] == '\n' || next == nil {
			// If we reached the final \n character or that there is no further
			// state, we do not call the next function.
			return nil
		}
		return next(entry, line, newPos)
	}
}

func parseHeader(quoted bool, hdr string, next stateFn) stateFn {
	return func(entry *AccessLogEntry, line string, pos int) error {
		data, off, err := readString(line, pos, quoted)
		if err != nil {
			return err
		}
		entry.Headers[hdr] = data
		newPos := pos + off
		if line[newPos] == ' ' {
			newPos++ // jump over next space, if any
		}
		if line[newPos] == '\n' || next == nil {
			// If we reached the final \n character or that there is no further
			// state, we do not call the next function.
			return nil
		}
		return next(entry, line, newPos)
	}
}

// extractFromQuotes extract the content of a quoted expression along with the
// ending quote position.
//
// s is expected to be a non empty string starting with a " (double quote)
// character. The closing double quote does not need to be at end of the string.
func extractFromQuotes(s string) (data string, off int, err error) {
	if s[0] != '"' {
		err = fmt.Errorf("got %q, want quote", s[0])
		return
	}
	if off = strings.Index(s[1:], "\""); off == -1 {
		err = errors.New("missing closing quote")
	} else {
		off++ // since we are starting from position 1
		data = s[1:off]
	}
	return
}

// readString reads the next string value from the given position of the line.
//
// It returns the string value as well as the offset between the initial
// position and the next character following the string.
func readString(line string, pos int, quoted bool) (data string, off int, err error) {
	input := line[pos:] // narrow the input to the current position
	if quoted {
		data, off, err = extractFromQuotes(input)
		off++ // go after the "
	} else {
		if off = strings.IndexAny(input, " \n"); off == -1 {
			// Should never happen since line is expected to have a trailing \n.
			data = input
			off = len(input)
		} else {
			data = input[:off]
		}
	}
	return
}

// readDateTime reads and parse the next datetime value, surrounded by square
// brackets ("[", "]"), from the given position of the line. The date is
// expected to be formatted using the same layout as the one defined by the
// StandardEnglishFormat constant.
//
// It returns the time.Time value as well as the offset between the initial
// position and the next character following the date.
func readDateTime(line string, pos int, quoted bool) (d time.Time, off int, err error) {
	input := line[pos:] // narrow the input to the current position
	if quoted {
		if input, off, err = extractFromQuotes(input); err != nil {
			return
		}
	}
	if input[0] != '[' {
		err = fmt.Errorf("got %q, want '['", input[0])
		return
	}
	idx := strings.Index(input, "]")
	if idx == -1 {
		err = errors.New("missing closing ']'")
		return
	}
	if !quoted {
		off = idx + 1
	}
	if d, err = time.Parse(StandardEnglishFormat, input[1:idx]); err != nil {
		err = errors.New("failed to parse datetime: " + err.Error())
	}
	return
}

// readInt reads the next integer value from the given position of the line.
//
// It returns the 64 integer value as well as the offset between the initial
// position and the next character following the integer.
func readInt(line string, pos int, quoted bool) (data int64, off int, err error) {
	input := line[pos:] // narrow the input to the current position
	if input[0] < '0' || input[0] > '9' {
		err = fmt.Errorf("got %q, want digit between 0 and 9", input[0])
		return
	}
	for off = 1; off < len(input); off++ {
		if input[off] < '0' || input[off] > '9' {
			break
		}
	}
	// error should not occur since only digit are kept
	data, err = strconv.ParseInt(input[:off], 10, 64)
	return
}
