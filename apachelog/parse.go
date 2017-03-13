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

// Time layout for parsing %t format.
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
		strings.Trim(formatStr, "\"")
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
	case UNKNOWN:
		fallthrough
	default:
		return nil, fmt.Errorf("%q format is not supported", formatStr)
	}
}

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

func (p *Parser) Parse() (*AccessLogEntry, error) {
	line, err := p.br.ReadString('\n')
	if err != nil {
		return nil, err
	}
	var entry AccessLogEntry
	if err := p.fn(&entry, line, 0); err != nil {
		return nil, err
	}
	return &entry, nil
}

func parseRemoteHost(quoted bool, next stateFn) stateFn {
	return func(entry *AccessLogEntry, line string, pos int) error {
		input := line[pos:] // narrow the input to the current position
		newPos := pos

		// handle quotes
		var (
			data string
			off  int
			err  error
		)
		if quoted {
			data, off, err = readWithQuotes(input)
			off++ // go after the "
		} else {
			data, off, err = readWithoutQuotes(input)
		}
		if err != nil {
			return err
		}
		newPos += off
		entry.RemoteHost = data

		// jump over next space, if any
		if line[newPos] == ' ' {
			newPos++
		}

		// If we reached the final \n character or that there is no further
		// state, we do not call the next function.
		if line[newPos] == '\n' || next == nil {
			return nil
		}

		return next(entry, line, newPos)
	}
}

func parseRemoteLogname(quoted bool, next stateFn) stateFn {
	return func(entry *AccessLogEntry, line string, pos int) error {
		input := line[pos:] // narrow the input to the current position
		newPos := pos

		// handle quotes
		var (
			data string
			off  int
			err  error
		)
		if quoted {
			data, off, err = readWithQuotes(input)
			off++ // go after the "
		} else {
			data, off, err = readWithoutQuotes(input)
		}
		if err != nil {
			return err
		}
		newPos += off
		entry.RemoteLogname = data

		// jump over next space, if any
		if line[newPos] == ' ' {
			newPos++
		}

		// If we reached the final \n character or that there is no further
		// state, we do not call the next function.
		if line[newPos] == '\n' || next == nil {
			return nil
		}

		return next(entry, line, newPos)
	}
}

// readWithQuotes extract the content of a quoted expression along with the
// ending quote position.
//
// s is expected to be a non empty string starting with a " (double quote)
// character. The closing double quote does not need to be at end of the string.
func readWithQuotes(s string) (data string, pos int, err error) {
	if s[0] != '"' {
		err = fmt.Errorf("got %q, want quote", s[0])
		return
	}
	pos++
	if idx := strings.Index(s[pos:], "\""); idx == -1 {
		err = errors.New("missing closing quote")
	} else {
		pos += idx
	}
	data = s[1:pos]
	return
}

func readWithoutQuotes(s string) (data string, pos int, err error) {
	if idx := strings.IndexAny(s, " \n"); idx == -1 {
		// should never happen
		err = errors.New("malformed input string")
	} else {
		pos += idx
	}
	data = s[:pos]
	return
}

func readDateTime(s string) (d time.Time, pos int, err error) {
	if s[0] != '[' {
		err = fmt.Errorf("got %q, want '['", s[0])
		return
	}
	pos++
	if idx := strings.Index(s[pos:], "]"); idx == -1 {
		err = errors.New("missing closing ']'")
	} else {
		pos += idx
		if d, err = time.Parse(StandardEnglishFormat, s[1:pos]); err != nil {
			err = errors.New("failed to parse datetime: " + err.Error())
		}
	}
	return
}

func readInt(s string) (data int64, pos int, err error) {
	if s[0] < '0' || s[0] > '9' {
		err = fmt.Errorf("got %q, want digit between 0 and 9", s[0])
		return
	}
	for pos = 1; pos < len(s); pos++ {
		if s[pos] < '0' || s[pos] > '9' {
			break
		}
	}
	// error should not occur since only digit are kept
	data, err = strconv.ParseInt(s[:pos], 10, 64)
	return
}
