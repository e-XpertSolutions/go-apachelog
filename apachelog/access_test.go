package apachelog

import "testing"

func TestNewRequestFirstLine(t *testing.T) {
	if got := NewRequestFirstLine("foo"); got.raw != "foo" {
		t.Errorf("NewRequestFirstLine(%q): got raw %q; want %q", "foo", got.raw, "foo")
	}
}

func TestRequestFirstLine_Method(t *testing.T) {
	rfl := RequestFirstLine{
		method: "GET",
	}
	if got := rfl.Method(); got != rfl.method {
		t.Errorf("RequestFirstLine({method: %q}).Method(): got %q; want %q", rfl.method, got, rfl.method)
	}
}

func TestRequestFirstLine_Path(t *testing.T) {
	rfl := RequestFirstLine{
		path: "/test",
	}
	if got := rfl.Path(); got != rfl.path {
		t.Errorf("RequestFirstLine({path: %q}).Path(): got %q; want %q", rfl.path, got, rfl.path)
	}
}

func TestRequestFirstLine_Protocol(t *testing.T) {
	rfl := RequestFirstLine{
		proto: "HTTP/1.1",
	}
	if got := rfl.Protocol(); got != rfl.proto {
		t.Errorf("RequestFirstLine({proto: %q}).Method(): got %q; want %q", rfl.proto, got, rfl.proto)
	}
}
func TestRequestFirstLine_String(t *testing.T) {
	if got := NewRequestFirstLine("test").String(); got != "test" {
		t.Errorf("RequestFirstLine({raw: %q}).String(): got %q; want %q", "test", got, "test")
	}
}
func TestRequestFirstLine_parse(t *testing.T) {
	type testCase struct {
		raw        string
		parsed     bool
		wantMethod string
		wantPath   string
		wantProto  string
	}
	tests := []testCase{
		{
			raw:        "GET /test HTTP/1.1",
			wantMethod: "GET",
			wantPath:   "/test",
			wantProto:  "HTTP/1.1",
		},
		{
			raw:    "GET /test HTTP/1.1",
			parsed: true,
		},
		{
			raw: "",
		},
		{
			raw: "GET /test",
		},
	}
	for i, test := range tests {
		rfl := NewRequestFirstLine(test.raw)
		rfl.parsed = test.parsed
		rfl.parse()

		if got := rfl.method; got != test.wantMethod {
			t.Errorf("%d. NewRequestFirstLine({ raw: %q, parsed: %v}).parse(): got method %q; want %q", i, test.raw, test.parsed, got, test.wantMethod)
		}
		if got := rfl.path; got != test.wantPath {
			t.Errorf("%d. NewRequestFirstLine({ raw: %q, parsed: %v}).parse(): got path %q; want %q", i, test.raw, test.parsed, got, test.wantPath)
		}
		if got := rfl.proto; got != test.wantProto {
			t.Errorf("%d. NewRequestFirstLine({ raw: %q, parsed: %v}).parse(): got proto %q; want %q", i, test.raw, test.parsed, got, test.wantProto)
		}
	}
}
