package sqltest

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		want    []query
		wantErr bool
		// Check context have expected values.
		wantCtx func(t *testing.T, ctx context.Context)
	}{
		{
			name:    "empty",
			src:     "",
			wantErr: true,
		},
		{
			name:    "whitespaces",
			src:     " \t\n",
			wantErr: true,
		},
		{
			name: "exec",
			src:  "SELECT 1",
			want: []query{
				{
					left:   position{index: 0},
					right:  position{index: 8},
					source: []byte("SELECT 1"),
					querier: &execQuerier{
						src: []byte("SELECT 1"),
					},
				},
			},
		},
		{
			name: "exec-multiple",
			src:  "SELECT 1;\nSELECT 2",
			want: []query{
				{
					left:    position{index: 0},
					right:   position{index: 8},
					source:  []byte("SELECT 1"),
					querier: &execQuerier{src: []byte("SELECT 1")},
				},
				{
					left:    position{line: 1, index: 10},
					right:   position{line: 1, index: 18},
					source:  []byte("SELECT 2"),
					querier: &execQuerier{src: []byte("SELECT 2")},
				},
			},
		},
		{
			name: "exec-trailing-ws",
			src:  "SELECT 1;\n \t;\n",
			want: []query{
				{
					left:   position{index: 0},
					right:  position{index: 8},
					source: []byte("SELECT 1"),
					querier: &execQuerier{
						src: []byte("SELECT 1"),
					},
				},
			},
		},
		{
			name: "exec-skipping-ws",
			src:  "SELECT 1;\n \t;\nSELECT 2",
			want: []query{
				{
					left:   position{index: 0},
					right:  position{index: 8},
					source: []byte("SELECT 1"),
					querier: &execQuerier{
						src: []byte("SELECT 1"),
					},
				},
				{
					left:   position{line: 2, index: 14},
					right:  position{line: 2, index: 22},
					source: []byte("SELECT 2"),
					querier: &execQuerier{
						src: []byte("SELECT 2"),
					},
				},
			},
		},
		{
			name: "parser(define+assert)",
			src:  "define TEST\nSELECT 1;\n\nassert TEST [1]",
			want: []query{
				{
					left:    position{line: 3, index: 23},
					right:   position{line: 3, index: 38},
					source:  []byte("assert TEST [1]"),
					querier: &assertQuerier{query: "SELECT 1", want: "[1]"},
				},
			},
			wantCtx: func(t *testing.T, ctx context.Context) {
				def := ctx.Value(ctxKeyDefine)
				if def == nil {
					t.Fatal("Value(ctxKeyDefine) = nil")
				}
				mdef, ok := def.(map[string]string)
				if !ok {
					t.Fatalf("Value(ctxKeyDefine).(type) = %T", def)
				}
				if g, w := fmt.Sprintf("%v", mdef), "map[TEST:SELECT 1]"; g != w {
					t.Fatalf("Value(ctxKeyDefine) = %q, want %q", g, w)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := New(strings.NewReader(tt.src), WithLimit(5))
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("New() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("New() succeeded unexpectedly")
			}
			if g, w := len(got.queries), len(tt.want); g != w {
				t.Errorf("len(New().queries) = %d, want %d", g, w)
			}
			if g, w := cmpQueries(got.queries), cmpQueries(tt.want); g != w {
				t.Errorf("New() = %q, want %q", g, w)
			}
			if tt.wantCtx != nil {
				if got.context == nil {
					t.Fatal("New().context = nil")
				}
				tt.wantCtx(t, got.context)
			}
		})
	}
}

// string representation of queries for comparision in tests.
func cmpQueries(qs []query) string {
	if qs == nil {
		return "<nil>"
	}
	s := &strings.Builder{}
	end := len(qs) - 1
	for i, q := range qs {
		s.WriteRune('{')
		cmpPos(s, "l", q.left)
		cmpPos(s, "r", q.right)
		s.WriteString(string(q.source))
		s.WriteRune(' ')
		switch t := q.querier.(type) {
		case *execQuerier:
			s.WriteString("*execQuerier{")
			s.WriteString(string(t.src))
			s.WriteRune('}')
		case *assertQuerier:
			s.WriteString("*assertQuerier{query: ")
			s.WriteString(t.query)
			s.WriteString(", want: ")
			s.WriteString(t.want)
			s.WriteRune('}')
		default:
			panic(fmt.Sprintf("unexpected sqltest.Querier: %#v", t))
		}
		s.WriteRune('}')
		if i != end {
			s.WriteString("  ")
		}
	}
	return s.String()
}

// string representation of position for comparision in tests.
func cmpPos(s *strings.Builder, key string, p position) {
	s.WriteString(key)
	s.WriteString(": ")
	s.WriteString(strconv.Itoa(p.line))
	s.WriteRune('.')
	s.WriteString(strconv.Itoa(p.index))
	s.WriteRune(' ')
}
