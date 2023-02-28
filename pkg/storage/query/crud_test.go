package query_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-article/pkg/helpers/gentest"
	"github.com/krixlion/dev_forum-article/pkg/storage/query"
	"github.com/krixlion/dev_forum-lib/env"
	"github.com/krixlion/dev_forum-lib/nulls"
)

var (
	port string
	host string
	pass string
)

func init() {
	env.Load("app")
	port = os.Getenv("DB_READ_PORT")
	host = os.Getenv("DB_READ_HOST")
	pass = os.Getenv("DB_READ_PASS")
}

func setUpDB() query.DB {
	db, err := query.MakeDB(host, port, pass, nulls.NullLogger{}, nulls.NullTracer{})
	if err != nil {
		log.Fatalf("Failed to make DB, err: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = db.Ping(ctx)
	if err != nil {
		log.Fatalf("Failed to ping to DB: %v", err)
	}

	return db
}

func Test_GetMultiple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping GetMultiple() integration test")
	}

	type args struct {
		offset string
		limit  string
	}
	tests := []struct {
		desc string
		args args
		want []entity.Article
	}{
		{
			desc: "Test if correctly returns keys added on K8s entrypoint",
			args: args{
				limit: "3",
			},
			want: []entity.Article{
				{
					Id:     "18",
					UserId: "18",
					Title:  "title-18",
					Body:   "body-18",
					CreatedAt: func() time.Time {
						time, err := time.Parse(time.RFC3339, "2023-01-31T22:58:24Z")
						if err != nil {
							panic(err)
						}
						return time
					}(),
					UpdatedAt: func() time.Time {
						time, err := time.Parse(time.RFC3339, "2023-01-31T22:58:24Z")
						if err != nil {
							panic(err)
						}
						return time
					}(),
				},
				{
					Id:     "17",
					UserId: "17",
					Title:  "title-17",
					Body:   "body-17",
					CreatedAt: func() time.Time {
						time, err := time.Parse(time.RFC3339, "2023-01-31T22:58:24Z")
						if err != nil {
							panic(err)
						}
						return time
					}(),
					UpdatedAt: func() time.Time {
						time, err := time.Parse(time.RFC3339, "2023-01-31T22:58:24Z")
						if err != nil {
							panic(err)
						}
						return time
					}(),
				},
				{
					Id:     "16",
					UserId: "16",
					Title:  "title-16",
					Body:   "body-16",
					CreatedAt: func() time.Time {
						time, err := time.Parse(time.RFC3339, "2023-01-31T22:58:24Z")
						if err != nil {
							panic(err)
						}
						return time
					}(),
					UpdatedAt: func() time.Time {
						time, err := time.Parse(time.RFC3339, "2023-01-31T22:58:24Z")
						if err != nil {
							panic(err)
						}
						return time
					}(),
				},
			},
		},
		{
			desc: "Test if correctly offset and sorting on multiple keys",
			args: args{
				offset: "2",
				limit:  "3",
			},
			want: []entity.Article{
				{
					Id:     "16",
					UserId: "16",
					Title:  "title-16",
					Body:   "body-16",
					CreatedAt: func() time.Time {
						time, err := time.Parse(time.RFC3339, "2023-01-31T22:58:24Z")
						if err != nil {
							panic(err)
						}
						return time
					}(),
					UpdatedAt: func() time.Time {
						time, err := time.Parse(time.RFC3339, "2023-01-31T22:58:24Z")
						if err != nil {
							panic(err)
						}
						return time
					}(),
				},
				{
					Id:     "15",
					UserId: "15",
					Title:  "title-15",
					Body:   "body-15",
					CreatedAt: func() time.Time {
						time, err := time.Parse(time.RFC3339, "2023-01-31T22:58:24Z")
						if err != nil {
							panic(err)
						}
						return time
					}(),
					UpdatedAt: func() time.Time {
						time, err := time.Parse(time.RFC3339, "2023-01-31T22:58:24Z")
						if err != nil {
							panic(err)
						}
						return time
					}(),
				},
				{
					Id:     "14",
					UserId: "14",
					Title:  "title-14",
					Body:   "body-14",
					CreatedAt: func() time.Time {
						time, err := time.Parse(time.RFC3339, "2023-01-31T22:58:24Z")
						if err != nil {
							panic(err)
						}
						return time
					}(),
					UpdatedAt: func() time.Time {
						time, err := time.Parse(time.RFC3339, "2023-01-31T22:58:24Z")
						if err != nil {
							panic(err)
						}
						return time
					}(),
				},
			},
		},
		{
			desc: "Test if correctly applies offset",
			args: args{
				offset: "2",
				limit:  "1",
			},
			want: []entity.Article{
				{
					Id:     "16",
					UserId: "16",
					Title:  "title-16",
					Body:   "body-16",
					CreatedAt: func() time.Time {
						time, err := time.Parse(time.RFC3339, "2023-01-31T22:58:24Z")
						if err != nil {
							panic(err)
						}
						return time
					}(),
					UpdatedAt: func() time.Time {
						time, err := time.Parse(time.RFC3339, "2023-01-31T22:58:24Z")
						if err != nil {
							panic(err)
						}
						return time
					}(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			db := setUpDB()
			defer db.Close()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

			got, err := db.GetMultiple(ctx, tt.args.offset, tt.args.limit)
			if err != nil {
				t.Errorf("db.GetMultiple() error = %+v\n", err)
				return
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("db.GetMultiple():\n got = %+v\n want = %+v\n", got, tt.want)
				return
			}
		})
	}
}

func Test_Get(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Get() integration test")
	}
	tests := []struct {
		desc    string
		arg     string
		want    entity.Article
		wantErr bool
	}{
		{
			desc: "Test if works on simple data",
			arg:  "12",
			want: entity.Article{
				Id:     "12",
				UserId: "12",
				Title:  "title-12",
				Body:   "body-12",
				CreatedAt: func() time.Time {
					time, err := time.Parse(time.RFC3339, "2023-01-31T22:58:24Z")
					if err != nil {
						panic(err)
					}
					return time
				}(),
				UpdatedAt: func() time.Time {
					time, err := time.Parse(time.RFC3339, "2023-01-31T22:58:24Z")
					if err != nil {
						panic(err)
					}
					return time
				}(),
			},
		},
		{
			desc:    "Test if returns error on non-existent key",
			arg:     gentest.RandomString(50),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			db := setUpDB()
			defer db.Close()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			got, err := db.Get(ctx, tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("db.Get():\n error = %+v wantErr = %+v\n", err, tt.wantErr)
				return
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("db.Get():\n got = %+v\n want = %+v\n", got, tt.want)
				return
			}
		})
	}
}

func Test_GetBelongingIDs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Create() integration test")
	}

	tests := []struct {
		name    string
		userId  string
		want    []string
		wantErr bool
	}{
		// {
		// 	name:   "",
		// 	userId: "",
		// 	want:   []string{},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setUpDB()
			defer db.Close()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			got, err := db.GetBelongingIDs(ctx, tt.userId)

			if (err != nil) != tt.wantErr {
				t.Errorf("db.GetBelongingIDs() error = %+v\n, wantErr %+v\n", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("db.GetBelongingIDs() = %+v\n want %+v\n diff = %+v\n", got, tt.want, cmp.Diff(got, tt.want))
				return
			}
		})
	}
}

func Test_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Create() integration test")
	}

	tests := []struct {
		desc    string
		arg     entity.Article
		wantErr bool
	}{
		{
			desc: "Test if works on simple data",
			arg: entity.Article{
				Id:     "test",
				UserId: "test",
				Title:  "title-test",
				Body:   "body-test",
				CreatedAt: func() time.Time {
					time, err := time.Parse(time.RFC3339, "2023-01-31T22:58:24Z")
					if err != nil {
						panic(err)
					}
					return time
				}(),
				UpdatedAt: func() time.Time {
					time, err := time.Parse(time.RFC3339, "2023-01-31T22:58:24Z")
					if err != nil {
						panic(err)
					}
					return time
				}(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			db := setUpDB()
			defer db.Close()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			err := db.Create(ctx, tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("db.Create() error = %+v", err)
				return
			}

			got, err := db.Get(ctx, tt.arg.Id)
			if err != nil {
				t.Errorf("db.Create() failed to db.Get() article:\n error = %+v\n wantErr = %+v\n", err, tt.wantErr)
				return
			}

			if !cmp.Equal(got, tt.arg) {
				t.Errorf("db.Create():\n got = %+v\n want = %+v\n", got, tt.arg)
				return
			}
		})
	}
}
func Test_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Update() integration test")
	}

	tests := []struct {
		desc    string
		arg     entity.Article
		wantErr bool
	}{
		{
			desc: "Test if works on simple data",
			arg: entity.Article{
				Id:     "test",
				UserId: "test",
				Title:  "title-test",
				Body:   "body-test: " + gentest.RandomString(2),
				CreatedAt: func() time.Time {
					time, err := time.Parse(time.RFC3339, "2023-01-31T22:58:24Z")
					if err != nil {
						panic(err)
					}
					return time
				}(),
				UpdatedAt: func() time.Time {
					time, err := time.Parse(time.RFC3339, "2023-01-31T22:58:24Z")
					if err != nil {
						panic(err)
					}
					return time
				}(),
			},
		},
		{
			desc: "Test if returns error on non-existent key",
			arg: entity.Article{
				Id:     "z" + gentest.RandomString(50),
				UserId: "test",
				Title:  "title-test",
				Body:   "body-test: " + gentest.RandomString(2),
				CreatedAt: func() time.Time {
					time, err := time.Parse(time.RFC3339, "2023-01-31T22:58:24Z")
					if err != nil {
						panic(err)
					}
					return time
				}(),
				UpdatedAt: func() time.Time {
					time, err := time.Parse(time.RFC3339, "2023-01-31T22:58:24Z")
					if err != nil {
						panic(err)
					}
					return time
				}(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			db := setUpDB()
			defer db.Close()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			err := db.Update(ctx, tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("db.Update():\n error = %+v wantErr = %+v\n", err, tt.wantErr)
				return
			}

			got, err := db.Get(ctx, tt.arg.Id)
			if (err != nil) != tt.wantErr {
				t.Errorf("db.Update() failed to db.Get() article:\n error = %+v\n", err)
				return
			}

			if !cmp.Equal(got, tt.arg) && !tt.wantErr {
				t.Errorf("db.Update():\n got = %+v\n want = %+v\n", got, tt.arg)
				return
			}
		})
	}
}

func Test_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Delete() integration test")
	}

	tests := []struct {
		desc string
		arg  string
	}{
		{
			desc: "Test if works on simple data",
			arg:  "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			db := setUpDB()
			defer db.Close()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			err := db.Delete(ctx, tt.arg)
			if err != nil {
				t.Errorf("db.Delete() error = %+v", err)
				return
			}

			_, err = db.Get(ctx, tt.arg)
			if err == nil {
				t.Errorf("db.Get() after db.Delete() returned nil error.")
				return
			}
		})
	}
}
