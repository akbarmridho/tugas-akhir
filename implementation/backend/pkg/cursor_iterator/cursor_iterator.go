package cursor_iterator

// Source https://github.com/Eun/go-pgx-cursor-iterator
// Copied because the published package are outdated and use outdated packages that used in this project
// Package cursoriterator provides functionality to iterate over big batches of postgres rows.

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"reflect"
	"sync"
)

// CursorIterator will be returned by NewCursorIterator().
// It provides functionality to loop over postgres rows and
// holds all necessary internal information for the functionality.
type CursorIterator struct {
	connector PgxConnector
	query     string
	args      []interface{}

	fetchQuery string

	values       []interface{}
	valuesPos    int
	valuesMaxPos int

	err error

	tx pgx.Tx

	mu         sync.Mutex
	cursorName string
}

// PgxConnector implements the Begin() function from the pgx and pgxpool packages.
type PgxConnector interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// NewCursorIterator can be used to create a new iterator.
// Required parameters:
//
//	connector                 most likely a *pgx.Conn or *pgxpool.Pool, needed to start a transaction on the database
//	values                    a slice where the fetched values should be stored in.
//	maxDatabaseExecutionTime  how long should one database operation be allowed to run.
//	query                     the query to fetch the rows
//	args                      arguments for the query
//
// Example Usage:
//
//	values := make([]User, 1000)
//	iter, err := NewCursorIterator(pool, values, time.Minute, "SELECT * FROM users WHERE role = $1", "Guest")
//	if err != nil {
//		panic(err)
//	}
//	defer iter.Close()
//	for iter.Next() {
//		fmt.Printf("Name: %s\n", values[iter.ValueIndex()].Name)
//	}
//	if err := iter.Error(); err != nil {
//		panic(err)
//	}
func NewCursorIterator(
	connector PgxConnector,
	values interface{},
	query string, args ...interface{},
) (*CursorIterator, error) {
	if connector == nil {
		return nil, errors.New("connector cannot be nil")
	}
	if values == nil {
		return nil, errors.New("values cannot be nil")
	}
	rv := reflect.ValueOf(values)
	if !rv.IsValid() {
		return nil, errors.New("values is invalid")
	}

	if rv.Kind() != reflect.Slice {
		return nil, errors.New("values must be a slice")
	}

	valuesCapacity := rv.Cap()

	if valuesCapacity <= 0 {
		return nil, errors.New("values must have a capacity bigger than 0")
	}

	valuesSlice := make([]interface{}, valuesCapacity)
	for i := 0; i < valuesCapacity; i++ {
		elem := rv.Index(i)
		if !elem.CanAddr() {
			return nil, errors.Errorf("unable to reference %s", elem.Type().String())
		}
		elem = elem.Addr()
		if !elem.CanInterface() {
			return nil, errors.Errorf("unable to get interface of %s", elem.Type().String())
		}
		valuesSlice[i] = elem.Interface()
	}

	cursorID := uuid.New()
	cursorName := hex.EncodeToString(cursorID[:])
	return &CursorIterator{
		connector:  connector,
		query:      query,
		args:       args,
		cursorName: cursorName,

		fetchQuery: fmt.Sprintf("FETCH %d IN %q", valuesCapacity, cursorName),

		values:       valuesSlice,
		valuesPos:    -2,
		valuesMaxPos: valuesCapacity - 1,

		err: nil,

		tx: nil,
	}, nil
}

func (iter *CursorIterator) fetchNextRows(ctx context.Context) {
	rows, err := iter.tx.Query(ctx, iter.fetchQuery)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			iter.close(ctx)
			return
		}
		iter.err = err
		return
	}

	scanner := pgxscan.NewRowScanner(rows)

	i := 0
	for rows.Next() {
		if i > iter.valuesMaxPos {
			iter.close(ctx)
			iter.err = errors.New("database returned more rows than expected")
			return
		}
		if err := scanner.Scan(iter.values[i]); err != nil {
			iter.close(ctx)
			iter.err = errors.Wrap(err, "unable to scan into values element")
			return
		}
		i++
	}

	if err := rows.Err(); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			iter.close(ctx)
			return
		}
		iter.close(ctx)
		iter.err = errors.Wrap(err, "unable to fetch rows")
		return
	}
	if i == 0 {
		iter.close(ctx)
		return
	}
	iter.valuesPos = 0
	iter.valuesMaxPos = i
}

// Next will return true if there is a next value available, false if there is no next value available.
// Next will also fetch next values when all current values have been iterated.
func (iter *CursorIterator) Next(ctx context.Context) bool {
	iter.mu.Lock()
	defer iter.mu.Unlock()
	// it is not the first row, and we already iterated over all rows: early exit
	if iter.valuesPos == -1 {
		return false
	}

	if iter.valuesPos == -2 {
		// first call:
		// start a transaction
		// and declare the cursor
		// start a transaction
		iter.tx, iter.err = iter.connector.Begin(ctx)
		if iter.err != nil {
			iter.err = errors.Wrap(iter.err, "unable to start transaction")
			return false
		}

		// declare cursor
		query := fmt.Sprintf("DECLARE %q CURSOR FOR %s", iter.cursorName, iter.query)
		if _, err := iter.tx.Exec(ctx, query, iter.args...); err != nil {
			iter.err = errors.Wrap(err, "unable to declare cursor")
			return false
		}
		// fetch the initial rows
		iter.fetchNextRows(ctx)
		// return true if we have rows
		return iter.valuesPos == 0
	}

	// do we still have items in the cache?
	if iter.valuesPos+1 < iter.valuesMaxPos {
		iter.valuesPos++
		return true
	}

	// we hit the end: fetch the next chunk of rows
	iter.fetchNextRows(ctx)
	return iter.valuesPos == 0
}

// ValueIndex will return the current value index that can be used to fetch the current value.
// Notice that it will return values below 0 when there is no next value available or the iteration didn't started yet.
func (iter *CursorIterator) ValueIndex() int {
	iter.mu.Lock()
	i := iter.valuesPos
	iter.mu.Unlock()
	return i
}

// Error will return the last error that appeared during fetching.
func (iter *CursorIterator) Error() error {
	iter.mu.Lock()
	err := iter.err
	iter.mu.Unlock()
	return err
}

func (iter *CursorIterator) close(ctx context.Context) {
	if iter.tx == nil {
		iter.err = nil
		return
	}

	iter.err = iter.tx.Rollback(ctx)
	iter.tx = nil
	iter.valuesPos = -1
}

// Close will close the iterator and all Next() calls will return false.
// After Close the iterator is unusable and can not be used again.
func (iter *CursorIterator) Close(ctx context.Context) error {
	iter.mu.Lock()
	defer iter.mu.Unlock()
	iter.close(ctx)
	return iter.err
}
