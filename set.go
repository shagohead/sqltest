package sqltest

import (
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"path"
	"path/filepath"
)

type Set struct {
	tests map[string]*Test
}

type TestRunner interface {
	Run(tx Tx) error
}

func (set *Set) All() iter.Seq2[string, TestRunner] {
	return func(yield func(string, TestRunner) bool) {
		for name, test := range set.tests {
			if !yield(name, test) {
				return
			}
		}
	}
}

// NewSet creates set for tests provided by tp.
func NewSet(tp iter.Seq2[string, io.Reader], opts ...option) (*Set, error) {
	set := &Set{tests: make(map[string]*Test)}
	var err error
	for name, reader := range tp {
		set.tests[name], err = New(reader, opts...)
		if err != nil {
			return nil, fmt.Errorf("test %q: %v", name, err)
		}
	}
	if len(set.tests) == 0 {
		return nil, errors.New("test set is empty")
	}
	return set, nil
}

func NewFileSet(pattern string, opts ...option) (*Set, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	return NewSet(func(yield func(string, io.Reader) bool) {
		for _, fname := range matches {
			f, err := os.Open(fname)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			if !yield(fname, f) {
				return
			}
		}
	}, opts...)
}

func DefaultFileSet(opts ...option) (*Set, error) {
	return NewFileSet(path.Join("testdata", "*.sql"), opts...)
}
