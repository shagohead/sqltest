package sqltest

import (
	"context"
	"fmt"
	"testing"
)

func Test_parseDefine(t *testing.T) {
	emptyCtx := context.Background()
	tests := []struct {
		name    string
		ctx     context.Context
		src     string
		want    map[string]string
		wantErr bool
	}{
		{
			name:    "missing-key",
			src:     "define \nSELECT 1",
			wantErr: true,
		},
		{
			name:    "missing-query",
			src:     "define A",
			wantErr: true,
		},
		{
			name:    "missing-query-after-nl",
			src:     "define A\n",
			wantErr: true,
		},
		{
			name:    "key-with-space",
			src:     "define A A\nSELECT 1",
			wantErr: true,
		},
		{
			name:    "key-with-tab",
			src:     "define A\tA\nSELECT 1",
			wantErr: true,
		},
		{
			name:    "key-trim",
			src:     "define   ABC\t  \nSELECT 1",
			want:    map[string]string{"ABC": "SELECT 1"},
			wantErr: false,
		},
		{
			name:    "initial-query",
			src:     "define A\nSELECT 1",
			want:    map[string]string{"A": "SELECT 1"},
			wantErr: false,
		},
		{
			name:    "append-query",
			ctx:     context.WithValue(emptyCtx, ctxKeyDefine, map[string]string{"B": "SELECT 2"}),
			src:     "define A\nSELECT 1",
			want:    map[string]string{"A": "SELECT 1", "B": "SELECT 2"},
			wantErr: false,
		},
		{
			name:    "conflict-query",
			ctx:     context.WithValue(emptyCtx, ctxKeyDefine, map[string]string{"A": "SELECT 2"}),
			src:     "define A\nSELECT 1",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx == nil {
				tt.ctx = emptyCtx
			}
			got, gotErr := parseDefine(tt.ctx, tt.src)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("parseDefine() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("parseDefine() succeeded unexpectedly")
			}
			def := got.Value(ctxKeyDefine)
			if def == nil {
				t.Fatal("parseDefine().Value(ctxKeyDefine) = nil")
			}
			mdef, ok := def.(map[string]string)
			if !ok {
				t.Fatalf("parseDefine().Value(ctxKeyDefine).(type) = %T", def)
			}
			if g, w := fmt.Sprintf("%v", mdef), fmt.Sprintf("%v", tt.want); g != w {
				t.Fatalf("parseDefine().Value(ctxKeyDefine) = %q, want %q", g, w)
			}
		})
	}
}

func Test_parseAssert(t *testing.T) {
	emptyCtx := context.Background()
	tests := []struct {
		name    string
		ctx     context.Context
		src     string
		want    assertQuerier
		wantErr bool
	}{
		{
			name:    "missing-key",
			src:     "assert ",
			wantErr: true,
		},
		{
			name:    "missing-value",
			src:     "assert A",
			wantErr: true,
		},
		{
			name:    "missing-query",
			src:     "assert A []",
			wantErr: true,
		},
		{
			name:    "ok",
			ctx:     context.WithValue(emptyCtx, ctxKeyDefine, map[string]string{"A": "SELECT 1"}),
			src:     "assert A []",
			want:    assertQuerier{query: "SELECT 1", want: "[]"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx == nil {
				tt.ctx = emptyCtx
			}
			got, gotErr := parseAssert(tt.ctx, tt.src)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("parseAssert() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("parseAssert() succeeded unexpectedly")
			}
			if got == nil {
				t.Fatal("parseAssert() = nil")
			}
			gotq := got.(*assertQuerier)
			if g, w := fmt.Sprintf("%#v", *gotq), fmt.Sprintf("%#v", tt.want); g != w {
				t.Errorf("parseAssert() = %q, want %q", g, w)
			}
		})
	}
}

func Test_define_Parse(t *testing.T) {
	emptyCtx := context.Background()
	tests := []struct {
		name    string
		ctx     context.Context
		src     string
		wantCtx bool
		wantQue bool
		wantErr bool
	}{
		{
			name:    "empty",
			src:     "",
			wantErr: false,
		},
		{
			name:    "unrecognized",
			src:     "unrecognized A []",
			wantErr: false,
		},
		{
			name:    "define-err",
			src:     "define ",
			wantErr: true,
		},
		{
			name:    "assert-err",
			src:     "assert ",
			wantErr: true,
		},
		{
			name:    "define-ok",
			src:     "define A\nSELECT 1",
			wantCtx: true,
		},
		{
			name:    "assert-ok",
			ctx:     context.WithValue(emptyCtx, ctxKeyDefine, map[string]string{"A": "SELECT 1"}),
			src:     "assert A []",
			wantQue: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx == nil {
				tt.ctx = emptyCtx
			}
			gotCtx, gotQue, gotErr := (&define{}).Parse(tt.ctx, []byte(tt.src))
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Parse() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Parse() succeeded unexpectedly")
			}
			if gotCtx := gotCtx != nil; gotCtx != tt.wantCtx {
				t.Errorf("Parse() got context = %v, want %v", gotCtx, tt.wantCtx)
			}
			if gotQue := gotQue != nil; gotQue != tt.wantQue {
				t.Errorf("Parse() got querier = %v, want %v", gotQue, tt.wantQue)
			}
		})
	}
}
