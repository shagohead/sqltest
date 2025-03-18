package sqltest

import (
	"context"
	"fmt"
	"testing"
)

func Test_except_Parse(t *testing.T) {
	cmp := func(v Querier) string {
		if v == nil {
			return "<nil>"
		}
		return fmt.Sprintf("%+v", v.(*exceptQuerier))
	}

	tests := []struct {
		name    string
		src     string
		wantQue *exceptQuerier
		wantErr bool
	}{
		{
			name:    "empty",
			src:     "",
			wantQue: nil,
		},
		{
			name:    "missing-substr",
			src:     "except \nSELECT 1",
			wantErr: true,
		},
		{
			name:    "missing-query",
			src:     "except ERROR",
			wantErr: true,
		},
		{
			name:    "missing-query-with-nl",
			src:     "except ERROR\n",
			wantErr: true,
		},
		{
			name:    "ok",
			src:     "except ERROR\nSELECT 1",
			wantQue: &exceptQuerier{query: "SELECT 1", except: "ERROR"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotQue, gotErr := (&except{}).Parse(context.Background(), []byte(tt.src))
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Parse() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Parse() succeeded unexpectedly")
			}
			if g, w := cmp(gotQue), cmp(tt.wantQue); g != w {
				t.Errorf("Parse() querier = %s, want %s", g, w)
			}
		})
	}
}
