package sqltest

import (
	"io"
	"iter"
	"slices"
	"strings"
	"testing"
)

func TestNewSet(t *testing.T) {
	provider := func(src map[string]string) iter.Seq2[string, io.Reader] {
		return func(yield func(string, io.Reader) bool) {
			for key, src := range src {
				if !yield(key, strings.NewReader(src)) {
					return
				}
			}
		}
	}

	tests := []struct {
		name    string
		tp      iter.Seq2[string, io.Reader]
		opts    []option
		want    map[string]*Test
		wantErr bool
	}{
		{
			name:    "empty",
			tp:      provider(map[string]string{}),
			wantErr: true,
		},
		{
			name: "ok",
			tp:   provider(map[string]string{"a": "SELECT 1", "b": "SELECT 2"}),
			want: map[string]*Test{
				"a": {
					queries: []query{
						{
							left:    position{index: 0},
							right:   position{index: 8},
							source:  []byte(`SELECT 1`),
							querier: &execQuerier{src: []byte(`SELECT 1`)},
						},
					},
				},
				"b": {
					queries: []query{
						{
							left:    position{index: 0},
							right:   position{index: 8},
							source:  []byte(`SELECT 2`),
							querier: &execQuerier{src: []byte(`SELECT 2`)},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := NewSet(tt.tp, tt.opts...)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("NewSet() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("NewSet() succeeded unexpectedly")
			}
			if got == nil {
				t.Fatal("NewSet() = nil")
			}
			if g, w := cmpTests(got.tests), cmpTests(tt.want); g != w {
				t.Errorf("NewSet() = %q, want %q", g, w)
			}
		})
	}
}

func TestNewFileSet(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		opts    []option
		want    map[string]*Test
		wantErr bool
	}{
		{
			name:    "not-found",
			pattern: "not_exists.sql",
			wantErr: true,
		},
		{
			name:    "concrete-file",
			pattern: "testdata/first.sql",
			want: map[string]*Test{
				"testdata/first.sql": {
					queries: []query{
						{
							left:    position{index: 0},
							right:   position{index: 27},
							source:  []byte(`INSERT INTO data VALUES (1)`),
							querier: &execQuerier{src: []byte(`INSERT INTO data VALUES (1)`)},
						},
					},
				},
			},
		},
		{
			name:    "glob-pattern",
			pattern: "testdata/*.sql",
			want: map[string]*Test{
				"testdata/first.sql": {
					queries: []query{
						{
							left:    position{index: 0},
							right:   position{index: 27},
							source:  []byte(`INSERT INTO data VALUES (1)`),
							querier: &execQuerier{src: []byte(`INSERT INTO data VALUES (1)`)},
						},
					},
				},
				"testdata/second.sql": {
					queries: []query{
						{
							left:    position{index: 0},
							right:   position{index: 8},
							source:  []byte(`SELECT 1`),
							querier: &execQuerier{src: []byte(`SELECT 1`)},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := NewFileSet(tt.pattern, tt.opts...)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("NewFileSet() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("NewFileSet() succeeded unexpectedly")
			}
			if got == nil {
				t.Fatal("NewFileSet() = nil")
			}
			if g, w := cmpTests(got.tests), cmpTests(tt.want); g != w {
				t.Errorf("NewFileSet() = %q, want %q", g, w)
			}
		})
	}
}

func TestDefaultFileSet(t *testing.T) {
	got, gotErr := DefaultFileSet()
	if gotErr != nil {
		t.Fatalf("DefaultFileSet() failed: %v", gotErr)
	}
	if got == nil {
		t.Fatal("DefaultFileSet() = nil")
	}
	var keys []string
	for k := range got.All() {
		keys = append(keys, k)
	}
	if g, w := strings.Join(keys, ", "), "testdata/first.sql, testdata/second.sql"; g != w {
		t.Fatalf("DefaultFileSet() keys: %q, want %q", g, w)
	}
}

// string representation of tests map for comparision in tests.
func cmpTests(tests map[string]*Test) string {
	s := &strings.Builder{}
	s.WriteByte('[')
	cnt := len(tests)
	keys := make([]string, 0, cnt)
	for k := range tests {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		s.WriteString(k)
		s.WriteString(": ")
		s.WriteString(cmpQueries(tests[k].queries))
		cnt--
		if cnt > 0 {
			s.WriteRune(' ')
		}
	}
	s.WriteByte(']')
	return s.String()
}
