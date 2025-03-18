package sqltest

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const (
	defineKey = "define "
	defineLen = len(defineKey)
	assertKey = "assert "
	assertLen = len(assertKey)
)

type define struct{}

var (
	errDefineWoNl   = errors.New("missing newline after define statement")
	errDefineWoKey  = errors.New("missing key in define statement")
	errDefineDouble = errors.New("duplicate key in define statement for test")
	errDefineWoQue  = errors.New("missing query in define statement")
	errDefineKeyWs  = errors.New("key in define should not contain spaces")
	errAssertWoKey  = errors.New("missing key in assert statement")
	errAssertWoVal  = errors.New("missing expected value in assert statement")
	errAssertUndef  = errors.New("assert uses not defined key")
)

type ctxKey int

const (
	ctxKeyDefine ctxKey = iota
)

func parseDefine(ctx context.Context, src string) (context.Context, error) {
	nsrc := len(src)
	nl := strings.IndexByte(src, '\n')
	if nl == -1 {
		return nil, errDefineWoNl
	}
	if nl == defineLen {
		return nil, errDefineWoKey
	}
	if nl+1 == nsrc {
		return nil, errDefineWoQue
	}
	key := strings.Trim(src[defineLen:nl], " \t")
	if strings.ContainsAny(key, " \t") {
		return nil, errDefineKeyWs
	}
	que := src[nl+1:]
	var defs map[string]string
	if d := ctx.Value(ctxKeyDefine); d != nil {
		defs = d.(map[string]string)
		if _, ok := defs[key]; ok {
			return nil, errDefineDouble
		}
	} else {
		defs = make(map[string]string)
		ctx = context.WithValue(ctx, ctxKeyDefine, defs)
	}
	defs[key] = que
	return ctx, nil
}

func parseAssert(ctx context.Context, src string) (Querier, error) {
	nsrc := len(src)
	if nsrc == assertLen {
		return nil, errAssertWoKey
	}
	src = strings.Trim(src[assertLen:], " \t")
	ws := strings.IndexByte(src, ' ')
	if ws == -1 {
		return nil, errAssertWoVal
	}
	if ws+1 == nsrc {
		return nil, errAssertWoVal
	}
	key := src[:ws]
	want := src[ws+1:]
	def := ctx.Value(ctxKeyDefine)
	if def == nil {
		return nil, errAssertUndef
	}
	mdef := def.(map[string]string)
	que, ok := mdef[key]
	if !ok {
		return nil, errAssertUndef
	}
	return &assertQuerier{query: que, want: want}, nil
}

type assertQuerier struct {
	query string
	want  string
}

// Query implements Querier.
func (a *assertQuerier) Query(ctx context.Context, tx Tx) error {
	rows, err := tx.Query(ctx, a.query)
	if err != nil {
		return err
	}
	var rs []string
	for rows.Next() {
		v, err := rows.String()
		if err != nil {
			return err
		}
		rs = append(rs, v)
	}
	if got := strings.Join(rs, " "); got != a.want {
		return fmt.Errorf("defined query returns %q, want %q", got, a.want)
	}
	return nil
}

var _ Querier = (*assertQuerier)(nil)

// Parse implements QueryParser.
func (d *define) Parse(ctx context.Context, src []byte) (context.Context, Querier, error) {
	source := string(src)
	if strings.HasPrefix(source, defineKey) {
		ctx, err := parseDefine(ctx, source)
		return ctx, nil, err
	}
	if strings.HasPrefix(source, assertKey) {
		query, err := parseAssert(ctx, source)
		return nil, query, err
	}
	return nil, nil, nil
}

var _ QueryParser = (*define)(nil)
