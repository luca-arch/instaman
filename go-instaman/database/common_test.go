package database_test

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/luca-arch/instaman/database"
	"github.com/luca-arch/instaman/database/models"
	"github.com/luca-arch/instaman/instaproxy"
	"github.com/stretchr/testify/mock"
)

type mockQuerier struct {
	mock.Mock
}

// Count calls the Count function to return the number of counted records.
func (q *mockQuerier) Count(ctx context.Context, db *database.Database, sql string, args ...any) (int32, error) {
	allArgs := make([]any, 0)
	allArgs = append(allArgs, ctx, db, oneLineSQL(sql))
	allArgs = append(allArgs, args...)

	funcArgs := q.Called(allArgs...)

	return funcArgs.Get(0).(int32), funcArgs.Error(1)
}

// Execute calls the Execute function to return any error that might occur.
func (q *mockQuerier) Execute(ctx context.Context, db *database.Database, sql string, args ...any) error {
	allArgs := make([]any, 0)
	allArgs = append(allArgs, ctx, db, oneLineSQL(sql))
	allArgs = append(allArgs, args...)

	funcArgs := q.Called(allArgs...)

	return funcArgs.Error(0)
}

// SelectJob calls the SelectOne function to return a `Job` objects.
func (q *mockQuerier) SelectJob(ctx context.Context, db *database.Database, sql string, args ...any) (*models.Job, error) {
	allArgs := make([]any, 0)
	allArgs = append(allArgs, ctx, db, oneLineSQL(sql))
	allArgs = append(allArgs, args...)

	funcArgs := q.Called(allArgs...)

	return funcArgs.Get(0).(*models.Job), funcArgs.Error(1)
}

// SelectJobs calls the Select function to return a list of `Job` objects.
func (q *mockQuerier) SelectJobs(ctx context.Context, db *database.Database, sql string, args ...any) ([]models.Job, error) {
	allArgs := make([]any, 0)
	allArgs = append(allArgs, ctx, db, oneLineSQL(sql))
	allArgs = append(allArgs, args...)

	funcArgs := q.Called(allArgs...)

	return funcArgs.Get(0).([]models.Job), funcArgs.Error(1)
}

// SelectUsers calls the Select function to return a list of `User` objects.
func (q *mockQuerier) SelectUsers(ctx context.Context, db *database.Database, sql string, args ...any) ([]models.User, error) {
	allArgs := make([]any, 0)
	allArgs = append(allArgs, ctx, db, oneLineSQL(sql))
	allArgs = append(allArgs, args...)

	funcArgs := q.Called(allArgs...)

	return funcArgs.Get(0).([]models.User), funcArgs.Error(1)
}

func oneLineSQL(sql string) string {
	s := strings.ReplaceAll(sql, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")

	return strings.Join(strings.Fields(s), " ")
}

func strPtr(str string) *string {
	return &str
}

func urlField(t *testing.T, s string) *instaproxy.URLField {
	t.Helper()

	u, err := url.Parse(s)
	if err != nil {
		t.Fatal(err)
	}

	return &instaproxy.URLField{URL: *u}
}
