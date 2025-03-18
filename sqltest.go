package sqltest

import (
	"context"
	"errors"
	"fmt"
	"io"
)

var errTestEmpty = errors.New("not found queries for test")

// Create new [Test] with queries from reader delimited by ;\n
func New(reader io.Reader, opts ...option) (*Test, error) {
	test := new(Test)
	config := newParserConfig(opts...)
	source, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	test.context = context.Background()
	end := len(source)
	var off position
	for {
		if config.limit == 0 {
			panic("cycle limit reached")
		}
		config.limit--
		if off.index >= end {
			break
		}
		if d := skipEmptyLines(source); d.index > 0 {
			source = source[d.index:]
			off = off.add(d)
		}
		right := config.delimiter(source)
		if right.index == 0 {
			continue
		}
		q := query{
			left:   off,
			right:  off.add(right),
			source: source[:right.index],
		}
		source = source[right.index:]
		off = q.right
		psrc := q.source
		if l := skipCommentaries(q.source); l.index > 0 {
			if len(psrc) == l.index {
				continue
			}
			psrc = psrc[l.index:]
		}
		parsed := false
		for _, parser := range config.parsers {
			nctx, que, err := parser.Parse(test.context, psrc)
			if err != nil {
				return nil, err
			}
			if nctx != nil {
				test.context = nctx
			}
			if que != nil {
				q.querier = que
				test.queries = append(test.queries, q)
			}
			if nctx != nil || que != nil {
				parsed = true
				break
			}
		}
		if !parsed {
			q.querier = &execQuerier{q.source}
			test.queries = append(test.queries, q)
		}
	}
	if len(test.queries) == 0 {
		return nil, errTestEmpty
	}
	return test, nil
}

// Overwrite parsing cycles limit.
func WithLimit(limit int) option {
	return func(pc *parseConfig) {
		pc.limit = limit
	}
}

type Test struct {
	// Test context accumulate test data.
	// This context passed into Query calls.
	context context.Context
	queries []query
}

type query struct {
	left, right position
	source      []byte
	querier     Querier
}

type QueryParser interface {
	// Parse is trying to parse query source to Querier object or updated context.
	//
	// If query source is not related to QueryParser implementation, it should return nils.
	// That is, Parse returns either the context, or the Querier, or nothing.
	// If it returns not nil, then the processing of the current query will be stopped at this QueryParser.
	Parse(context.Context, []byte) (context.Context, Querier, error)
}

// Querier is query'like object. Which whould be called in Test.Run.
type Querier interface {
	// Query called in Test.Run. Querier implementation do some query and check it results.
	Query(ctx context.Context, tx Tx) error
}

// Tx declares database interaction interface.
type Tx interface {
	// Exec query omitting results.
	Exec(ctx context.Context, sql string, args ...any) error

	// Query [Rows] from database.
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
}

// Rows returned by Tx.Query.
type Rows interface {
	Close()
	Next() bool
	Err() error

	// String returns string representation of row values.
	String() (string, error)
}

func (test *Test) Run(tx Tx) error {
	var err error
	for _, q := range test.queries {
		err = q.querier.Query(test.context, tx)
		if err != nil {
			return fmt.Errorf(
				"Query on lines %d:%d at bytes %d:%d fails: %v.\nQuery source: %s",
				q.left.line+1, q.right.line+1, q.left.index, q.right.index, err, q.source,
			)
		}
	}
	return nil
}

type execQuerier struct {
	src []byte
}

// Query implements Querier.
func (e *execQuerier) Query(ctx context.Context, tx Tx) error {
	if err := tx.Exec(ctx, string(e.src)); err != nil {
		return fmt.Errorf("Exec(): %v", err)
	}
	return nil
}

var _ Querier = (*execQuerier)(nil)

type option func(*parseConfig)

func newParserConfig(opts ...option) *parseConfig {
	p := new(parseConfig)
	for _, opt := range opts {
		opt(p)
	}
	if p.limit == 0 {
		p.limit = 100
	}
	if p.delimiter == nil {
		p.delimiter = defaultQueryDelimiter
	}
	if p.parsers == nil {
		p.parsers = []QueryParser{
			&define{},
			&except{},
		}
	}
	return p
}

type parseConfig struct {
	limit     int
	delimiter QueryDelimiter
	parsers   []QueryParser
}
