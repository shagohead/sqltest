package sqltest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
)

type except struct{}

const (
	exceptKey = "except "
	exceptLen = len(exceptKey)
)

var (
	errExceptMissQuery = errors.New("missing query in except statement")
	errExceptMissStr   = errors.New("missing exception substring in except statement")
)

// Parse implements QueryParser.
func (e *except) Parse(ctx context.Context, src []byte) (context.Context, Querier, error) {
	if !bytes.HasPrefix(src, []byte(exceptKey)) {
		return nil, nil, nil
	}
	nl := bytes.IndexByte(src, '\n')
	if nl == -1 {
		return nil, nil, errExceptMissQuery
	}
	if nl == exceptLen {
		return nil, nil, errExceptMissStr
	}
	if nl+1 == len(src) {
		return nil, nil, errExceptMissQuery
	}
	return nil, &exceptQuerier{
		query:  string(src[nl+1:]),
		except: string(src[exceptLen:nl]),
	}, nil
}

var _ QueryParser = (*except)(nil)

type exceptQuerier struct {
	query  string
	except string
}

// Query implements Querier.
func (e *exceptQuerier) Query(ctx context.Context, tx Tx) error {
	err := tx.Exec(ctx, e.query)
	if strings.Contains(err.Error(), e.except) {
		return nil
	}
	return fmt.Errorf("expected error contained %q, but got %v", e.except, err)
}
