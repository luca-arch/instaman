/*
 * Instaman - Simple Instagram account manager.
 *
 * Copyright (C) 2024 Luca Contini
 *
 * This program is free software: you can redistribute it and/or modify it
 * under the terms of the GNU General Public License as published by the Free
 * Software Foundation, either version 3 of the License, or (at your option)
 * any later version.
 *
 * This program is distributed in the hope that it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for
 * more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * this program. If not, see <http://www.gnu.org/licenses/>.
 */

package database_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/luca-arch/instaman/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockdb struct {
	mock.Mock
}

func (m *mockdb) Query(ctx context.Context, sql string, _ ...any) (pgx.Rows, error) { //nolint:ireturn // mock ok
	a := m.Called(ctx, sql, 123)

	return a.Get(0).(pgx.Rows), a.Error(1)
}

// Need to mock the whole pgx.Rows interface...
type mockrows struct {
	expectedNextCalls int
	countNextCalls    int
	onScan            func(...any) error
}

func (m *mockrows) Close() {}

func (m *mockrows) Err() error {
	return nil
}

func (m *mockrows) CommandTag() pgconn.CommandTag {
	return pgconn.NewCommandTag("mock")
}

func (m *mockrows) FieldDescriptions() []pgconn.FieldDescription {
	return make([]pgconn.FieldDescription, 0)
}

func (m *mockrows) Next() bool {
	if m.countNextCalls == m.expectedNextCalls {
		return false
	}

	m.countNextCalls++

	return true
}

func (m *mockrows) Scan(dest ...any) error {
	return m.onScan(dest...)
}

func (m *mockrows) Values() ([]any, error) {
	return nil, nil
}

func (m *mockrows) RawValues() [][]byte {
	return make([][]byte, 0)
}

func (m *mockrows) Conn() *pgx.Conn {
	return nil
}

func TestCount(t *testing.T) {
	// FIXME: this test is complicated and useless.
	t.Parallel()

	ctx := context.TODO()

	mock := &mockdb{}
	rows := &mockrows{
		expectedNextCalls: 1,
		onScan: func(dest ...any) error {
			t.Helper()

			i := dest[0]
			p, ok := i.(*int)

			if !ok {
				t.Fatalf("Expected *int, got %#v", p)
			}

			*p = 400

			return nil
		},
	}

	mock.On("Query", ctx, "mock sql statement", 123).Return(rows, nil)

	db := database.NewPool(ctx, "postgres://postgres:123@127.0.0.1:5432/dummy").WithPool(mock)

	res, err := database.Count(ctx, db, "mock sql statement", 123)

	assert.NoError(t, err)
	assert.Equal(t, 400, res)
}
