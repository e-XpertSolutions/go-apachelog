package apachelog

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

var makeStateFnTests = []struct {
	expr []string // input
	err  error    // expected error
}{
	{
		expr: []string{"%h"},
		err:  nil,
	},
	{
		expr: []string{},
		err:  nil,
	},
	{
		expr: nil,
		err:  nil,
	},
	{
		expr: []string{"foo"},
		err:  errors.New("\"foo\" format is not supported"),
	},
	{
		expr: []string{"%h", "foo"},
		err:  errors.New("\"foo\" format is not supported"),
	},
}

func TestMakeStateFn(t *testing.T) {
	for _, test := range makeStateFnTests {
		_, err := makeStateFn(test.expr)
		switch {
		case err == nil && test.err != nil:
			t.Errorf("makeSateFn(%v): expected error %q; got none", test.expr, test.err.Error())
		case err != nil && test.err == nil:
			t.Errorf("makeSateFn(%v): unexpected error %q", test.expr, err.Error())
		case err != nil && test.err != nil && err.Error() != test.err.Error():
			t.Errorf("makeSateFn(%v): expected error %q; got %q",
				test.expr, test.err.Error(), err.Error())
		}
	}
}

func TestParseRemoteHost(t *testing.T) {
	type testCase struct {
		quoted bool
		line   string
		err    error
	}

	testCases := []testCase{
		{quoted: true, line: "\"foobar\" 42", err: nil},
		{quoted: false, line: "foobar 42", err: nil},
	}

	for i, test := range testCases {
		var entry AccessLogEntry
		err := parseRemoteHost(test.quoted, nil)(&entry, test.line, 0)
		switch {
		case err == nil && test.err != nil:
			t.Errorf("%d. parseRemoteHost(%v, nil)({}, %q, 0): expected error %q; got none",
				i, test.quoted, test.line, test.err.Error())
			continue
		case err != nil && test.err == nil:
			t.Errorf("%d. parseRemoteHost(%v, nil)({}, %q, 0): unexpected error %q",
				i, test.quoted, test.line, test.err.Error())
			continue
		case err != nil && test.err != nil && err.Error() != test.err.Error():
			t.Errorf("%d. parseRemoteHost(%v, nil)({}, %q, 0): expected error %q; got %q",
				i, test.quoted, test.line, test.err.Error(), err.Error())
			continue
		}
		if got := entry.RemoteHost; got != "foobar" {
			t.Errorf("%d. parseRemoteHost(%v, nil)({}, %q, 0): expected RemoteHost %q, got %q",
				i, test.quoted, test.line, "foobar", got)
		}
	}
}

func TestParseRemoteLogname(t *testing.T) {
	type testCase struct {
		quoted bool
		line   string
		err    error
	}

	testCases := []testCase{
		{quoted: true, line: "\"foobar\" 42", err: nil},
		{quoted: false, line: "foobar 42", err: nil},
	}

	for i, test := range testCases {
		var entry AccessLogEntry
		err := parseRemoteHost(test.quoted, nil)(&entry, test.line, 0)
		switch {
		case err == nil && test.err != nil:
			t.Errorf("%d. parseRemoteHost(%v, nil)({}, %q, 0): expected error %q; got none",
				i, test.quoted, test.line, test.err.Error())
			continue
		case err != nil && test.err == nil:
			t.Errorf("%d. parseRemoteHost(%v, nil)({}, %q, 0): unexpected error %q",
				i, test.quoted, test.line, test.err.Error())
			continue
		case err != nil && test.err != nil && err.Error() != test.err.Error():
			t.Errorf("%d. parseRemoteHost(%v, nil)({}, %q, 0): expected error %q; got %q",
				i, test.quoted, test.line, test.err.Error(), err.Error())
			continue
		}
		if got := entry.RemoteHost; got != "foobar" {
			t.Errorf("%d. parseRemoteHost(%v, nil)({}, %q, 0): expected RemoteHost %q, got %q",
				i, test.quoted, test.line, "foobar", got)
		}
	}
}

func TestReadWithQuotes(t *testing.T) {
	type testCase struct {
		in, out string
		pos     int
		err     error
	}

	testCases := []testCase{
		{in: "\"foo\" bar", out: "foo", pos: 4, err: nil},
		{in: "\"foo bar", out: "", pos: 1, err: errors.New("missing closing quote")},
		{in: "foo\" bar", out: "", pos: 0, err: fmt.Errorf("got 'f', want quote")},
	}

	for i, test := range testCases {
		data, pos, err := readWithQuotes(test.in)
		if data != test.out {
			t.Errorf("%d. readWithQuotes(%q) got data %q, want %q", i, test.in, data, test.out)
		}
		if pos != test.pos {
			t.Errorf("%d. readWithQuotes(%q) got position %d, want %d", i, test.in, pos, test.pos)
		}
		switch {
		case err == nil && test.err != nil:
			t.Errorf("%d. readWithQuotes(%q): expected error %q; got none",
				i, test.in, test.err.Error())
		case err != nil && test.err == nil:
			t.Errorf("%d. readWithQuotes(%q): unexpected error %q",
				i, test.in, err.Error())
		case err != nil && test.err != nil && err.Error() != test.err.Error():
			t.Errorf("%d. readWithQuotes(%q): expected error %q; got %q",
				i, test.in, test.err.Error(), err.Error())
		}
	}
}

func TestReadWithoutQuotes(t *testing.T) {
	type testCase struct {
		in, out string
		pos     int
		err     error
	}

	testCases := []testCase{
		{in: "127.0.0.1\n", out: "127.0.0.1", pos: 9, err: nil},
		{in: "foo bar", out: "foo", pos: 3, err: nil},
		{in: "foobar", out: "", pos: 0, err: errors.New("malformed input string")},
	}

	for i, test := range testCases {
		data, pos, err := readWithoutQuotes(test.in)
		if data != test.out {
			t.Errorf("%d. readWithoutQuotes(%q) got data %q, want %q", i, test.in, data, test.out)
		}
		if pos != test.pos {
			t.Errorf("%d. readWithoutQuotes(%q) got position %d, want %d", i, test.in, pos, test.pos)
		}
		switch {
		case err == nil && test.err != nil:
			t.Errorf("%d. readWithoutQuotes(%q): expected error %q; got none",
				i, test.in, test.err.Error())
		case err != nil && test.err == nil:
			t.Errorf("%d. readWithoutQuotes(%q): unexpected error %q",
				i, test.in, err.Error())
		case err != nil && test.err != nil && err.Error() != test.err.Error():
			t.Errorf("%d. readWithoutQuotes(%q): expected error %q; got %q",
				i, test.in, test.err.Error(), err.Error())
		}
	}
}

func TestReadDateTime(t *testing.T) {
	type testCase struct {
		in  string
		out time.Time
		pos int
		err error
	}

	// switzerland timezone
	tz, err := time.LoadLocation("Europe/Zurich")
	if err != nil {
		// should not happen...
		t.Fatal(err)
	}

	testCases := []testCase{
		{
			in:  "[16/Nov/2016:09:25:05 +0100] foobar",
			out: time.Date(2016, 11, 16, 9, 25, 5, 0, tz),
			pos: 27,
			err: nil,
		},
		{in: "foobar]", out: time.Time{}, pos: 0, err: errors.New("got 'f', want '['")},
		{in: "[foobar", out: time.Time{}, pos: 1, err: errors.New("missing closing ']'")},
		{
			in:  "[2016-11-16 09:25:05 +0100]",
			out: time.Time{},
			pos: 26,
			err: errors.New("failed to parse datetime: parsing time \"2016-11-16 09:25:05 +0100\" as \"02/Jan/2006:15:04:05 -0700\": cannot parse \"16-11-16 09:25:05 +0100\" as \"/\""),
		},
	}

	for i, test := range testCases {
		d, pos, err := readDateTime(test.in)
		if !d.Equal(test.out) {
			t.Errorf("%d. readDateTime(%q) got date %q, want %q",
				i, test.in, d.Format(StandardEnglishFormat), test.out.Format(StandardEnglishFormat))
		}
		if pos != test.pos {
			t.Errorf("%d. readDateTime(%q) got position %d, want %d", i, test.in, pos, test.pos)
		}
		switch {
		case err == nil && test.err != nil:
			t.Errorf("%d. readDateTime(%q): expected error %q; got none",
				i, test.in, test.err.Error())
		case err != nil && test.err == nil:
			t.Errorf("%d. readDateTime(%q): unexpected error %q",
				i, test.in, err.Error())
		case err != nil && test.err != nil && err.Error() != test.err.Error():
			t.Errorf("%d. readDateTime(%q): expected error %q; got %q",
				i, test.in, test.err.Error(), err.Error())
		}
	}
}

func TestReadInt(t *testing.T) {
	type testCase struct {
		in  string
		out int64
		pos int
		err error
	}

	testCases := []testCase{
		{in: "1234567890 foo", out: 1234567890, pos: 10, err: nil},
		{in: "1234567890foo", out: 1234567890, pos: 10, err: nil},
		{in: "1234567890\n", out: 1234567890, pos: 10, err: nil},
		{in: "foo123", out: 0, pos: 0, err: errors.New("got 'f', want digit between 0 and 9")},
	}

	for i, test := range testCases {
		data, pos, err := readInt(test.in)
		if data != test.out {
			t.Errorf("%d. readInt(%q) got data %d, want %d", i, test.in, data, test.out)
		}
		if pos != test.pos {
			t.Errorf("%d. readInt(%q) got position %d, want %d", i, test.in, pos, test.pos)
		}
		switch {
		case err == nil && test.err != nil:
			t.Errorf("%d. readInt(%q): expected error %q; got none",
				i, test.in, test.err.Error())
		case err != nil && test.err == nil:
			t.Errorf("%d. readInt(%q): unexpected error %q",
				i, test.in, err.Error())
		case err != nil && test.err != nil && err.Error() != test.err.Error():
			t.Errorf("%d. readInt(%q): expected error %q; got %q",
				i, test.in, test.err.Error(), err.Error())
		}
	}
}
