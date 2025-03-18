package sqltest

import "testing"

func Test_defaultQueryDelimiter(t *testing.T) {
	for _, tt := range []struct {
		name string
		src  string
		want position
	}{
		{name: "empty", src: "", want: position{index: 0}},
		{name: "only-ws", src: " ", want: position{index: 1}},
		{name: "only-delim", src: ";", want: position{index: 0}},
		{name: "ws-delim", src: " ;", want: position{index: 1}},
		{name: "delim-before-nl", src: ";\n", want: position{index: 0}},
		{name: "ws-delim-before-nl", src: " ;\n", want: position{index: 1}},
		{name: "delim-ws-before-nl", src: " ; \n", want: position{index: 1}},
		{name: "only-nl", src: "\n", want: position{index: 1, line: 1}},
		{name: "only-nl-mult", src: "\n\n", want: position{index: 2, line: 2}},
		{name: "delim-after-nl", src: "\n;", want: position{index: 1, line: 1}},
		{name: "word-delim", src: "test;", want: position{index: 4}},
		{name: "delim-word", src: ";test", want: position{index: 5}},
		{name: "delim-ws-word", src: ";\ntest", want: position{index: 0}},
		{name: "alpha-nl-alpha", src: "a\nb", want: position{index: 3, line: 1}},
		{name: "alpha-nl-alpha-delim", src: "a\nb;", want: position{index: 3, line: 1}},
		{name: "alpha-nl-alpha-delim-nl-alpha", src: "a\nb;\nc", want: position{index: 3, line: 1}},
		{name: "fix-offset-overlap", src: "a\nb\nc", want: position{line: 2, index: 5}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := defaultQueryDelimiter([]byte(tt.src))
			if got != tt.want {
				t.Errorf("defaultQueryDelimiter() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_skipEmptyLines(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want position
	}{
		{name: "empty", src: "", want: position{}},
		{name: "newlines", src: "\n\n", want: position{line: 2, index: 2}},
		{name: "whitespaces", src: " \t", want: position{index: 2}},
		{name: "whitespaces+newlines", src: "\t\n ", want: position{line: 1, index: 3}},
		{name: "non-whitespace", src: "\t\n test", want: position{line: 1, index: 3}},
		{name: "empty-delimiter", src: ";\n ", want: position{line: 1, index: 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := skipEmptyLines([]byte(tt.src))
			if got != tt.want {
				t.Errorf("skipEmptyLines() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_skipCommentaries(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want position
	}{
		{name: "empty", src: "", want: position{}},
		{name: "line-comment", src: "--", want: position{index: 2}},
		{name: "line-comment-full", src: "--skip\ntest", want: position{line: 1, index: 7}},
		{name: "block-comment", src: "/*test*/test", want: position{line: 0, index: 8}},
		{name: "block-comment-full", src: "/*sk\nip*/\ntest", want: position{line: 2, index: 10}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := skipCommentaries([]byte(tt.src))
			if got != tt.want {
				t.Errorf("skipCommentaries() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
