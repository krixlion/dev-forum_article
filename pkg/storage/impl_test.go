package storage_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/event"
	"github.com/krixlion/dev-forum_article/pkg/helpers/gentest"
	"github.com/krixlion/dev-forum_article/pkg/helpers/nulls"
	"github.com/krixlion/dev-forum_article/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Get(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}

	testCases := []struct {
		desc    string
		query   mockQuery
		args    args
		want    entity.Article
		wantErr bool
	}{
		{
			desc: "Test if method is invoked",
			args: args{
				ctx: context.Background(),
				id:  "",
			},
			want: entity.Article{},
			query: func() mockQuery {
				m := mockQuery{new(mock.Mock)}
				m.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(entity.Article{}, nil).Once()
				return m
			}(),
		},
		{
			desc: "Test if method forwards an error",
			args: args{
				ctx: context.Background(),
				id:  "",
			},
			want:    entity.Article{},
			wantErr: true,
			query: func() mockQuery {
				m := mockQuery{new(mock.Mock)}
				m.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(entity.Article{}, errors.New("test err")).Once()
				return m
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			db := storage.NewStorage(mockCmd{}, tC.query, nulls.NullLogger{})
			got, err := db.Get(tC.args.ctx, tC.args.id)
			if (err != nil) != tC.wantErr {
				t.Errorf("storage.Get():\n error = %+v\n wantErr = %+v\n", err, tC.wantErr)
				return
			}

			if !cmp.Equal(got, tC.want) {
				t.Errorf("storage.Get():\n got = %+v\n want = %+v\n", got, tC.want)
				return
			}
			assert.True(t, tC.query.AssertCalled(t, "Get", mock.Anything, tC.args.id))
		})
	}
}
func Test_GetMultiple(t *testing.T) {
	type args struct {
		ctx    context.Context
		offset string
		limit  string
	}

	testCases := []struct {
		desc    string
		query   mockQuery
		args    args
		want    []entity.Article
		wantErr bool
	}{
		{
			desc: "Test if method is invoked",
			args: args{
				ctx:    context.Background(),
				limit:  "",
				offset: "",
			},
			want: []entity.Article{},
			query: func() mockQuery {
				m := mockQuery{new(mock.Mock)}
				m.On("GetMultiple", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return([]entity.Article{}, nil).Once()
				return m
			}(),
		},
		{
			desc: "Test if method forwards an error",
			args: args{
				ctx:    context.Background(),
				limit:  "",
				offset: "",
			},
			want:    []entity.Article{},
			wantErr: true,
			query: func() mockQuery {
				m := mockQuery{new(mock.Mock)}
				m.On("GetMultiple", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return([]entity.Article{}, errors.New("test err")).Once()
				return m
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			db := storage.NewStorage(mockCmd{}, tC.query, nulls.NullLogger{})
			got, err := db.GetMultiple(tC.args.ctx, tC.args.offset, tC.args.limit)
			if (err != nil) != tC.wantErr {
				t.Errorf("storage.GetMultiple():\n error = %+v\n wantErr = %+v\n", err, tC.wantErr)
				return
			}

			if !cmp.Equal(got, tC.want, cmpopts.EquateEmpty()) {
				t.Errorf("storage.GetMultiple():\n got = %+v\n want = %+v\n", got, tC.want)
				return
			}

			assert.True(t, tC.query.AssertCalled(t, "GetMultiple", mock.Anything, tC.args.offset, tC.args.limit))
		})
	}
}
func Test_Create(t *testing.T) {
	type args struct {
		ctx     context.Context
		article entity.Article
	}

	testCases := []struct {
		desc    string
		cmd     mockCmd
		args    args
		wantErr bool
	}{
		{
			desc: "Test if method is invoked",
			args: args{
				ctx:     context.Background(),
				article: entity.Article{},
			},

			cmd: func() mockCmd {
				m := mockCmd{new(mock.Mock)}
				m.On("Create", mock.Anything, mock.AnythingOfType("entity.Article")).Return(nil).Once()
				return m
			}(),
		},
		{
			desc: "Test if an error is forwarded",
			args: args{
				ctx:     context.Background(),
				article: entity.Article{},
			},
			wantErr: true,
			cmd: func() mockCmd {
				m := mockCmd{new(mock.Mock)}
				m.On("Create", mock.Anything, mock.AnythingOfType("entity.Article")).Return(errors.New("test err")).Once()
				return m
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			db := storage.NewStorage(tC.cmd, mockQuery{}, nulls.NullLogger{})
			err := db.Create(tC.args.ctx, tC.args.article)
			if (err != nil) != tC.wantErr {
				t.Errorf("storage.Create():\n error = %+v\n wantErr = %+v\n", err, tC.wantErr)
				return
			}
			assert.True(t, tC.cmd.AssertCalled(t, "Create", mock.Anything, tC.args.article))
		})
	}
}
func Test_Update(t *testing.T) {
	type args struct {
		ctx     context.Context
		article entity.Article
	}

	testCases := []struct {
		desc    string
		cmd     mockCmd
		args    args
		wantErr bool
	}{
		{
			desc: "Test if method is invoked",
			args: args{
				ctx:     context.Background(),
				article: entity.Article{},
			},

			cmd: func() mockCmd {
				m := mockCmd{new(mock.Mock)}
				m.On("Update", mock.Anything, mock.AnythingOfType("entity.Article")).Return(nil).Once()
				return m
			}(),
		},
		{
			desc: "Test if error is forwarded",
			args: args{
				ctx:     context.Background(),
				article: entity.Article{},
			},
			wantErr: true,
			cmd: func() mockCmd {
				m := mockCmd{new(mock.Mock)}
				m.On("Update", mock.Anything, mock.AnythingOfType("entity.Article")).Return(errors.New("test err")).Once()
				return m
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			db := storage.NewStorage(tC.cmd, mockQuery{}, nulls.NullLogger{})
			err := db.Update(tC.args.ctx, tC.args.article)
			if (err != nil) != tC.wantErr {
				t.Errorf("storage.Update():\n error = %+v\n wantErr = %+v\n", err, tC.wantErr)
				return
			}
			assert.True(t, tC.cmd.AssertCalled(t, "Update", mock.Anything, tC.args.article))
		})
	}
}
func Test_Delete(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}

	testCases := []struct {
		desc    string
		cmd     mockCmd
		args    args
		wantErr bool
	}{
		{
			desc: "Test if method is invoked",
			args: args{
				ctx: context.Background(),
				id:  "",
			},

			cmd: func() mockCmd {
				m := mockCmd{new(mock.Mock)}
				m.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()
				return m
			}(),
		},
		{
			desc: "Test if error is forwarded",
			args: args{
				ctx: context.Background(),
				id:  "",
			},
			wantErr: true,
			cmd: func() mockCmd {
				m := mockCmd{new(mock.Mock)}
				m.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(errors.New("test err")).Once()
				return m
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			db := storage.NewStorage(tC.cmd, mockQuery{}, nulls.NullLogger{})
			err := db.Delete(tC.args.ctx, tC.args.id)
			if (err != nil) != tC.wantErr {
				t.Errorf("storage.Delete():\n error = %+v\n wantErr = %+v\n", err, tC.wantErr)
				return
			}
			assert.True(t, tC.cmd.AssertCalled(t, "Delete", mock.Anything, tC.args.id))
			assert.True(t, tC.cmd.AssertExpectations(t))
		})
	}
}

func Test_CatchUp(t *testing.T) {
	testCases := []struct {
		desc   string
		arg    event.Event
		query  mockQuery
		method string
	}{
		{
			desc: "Test if Update method is invoked on ArticleUpdated event",
			arg: event.Event{
				Type: event.ArticleUpdated,
				Body: gentest.RandomJSONArticle(2, 3),
			},
			method: "Update",
			query: func() mockQuery {
				m := mockQuery{new(mock.Mock)}
				m.On("Update", mock.Anything, mock.AnythingOfType("entity.Article")).Return(nil).Once()
				return m
			}(),
		},
		{
			desc: "Test if Create method is invoked on ArticleCreated event",
			arg: event.Event{
				Type: event.ArticleCreated,
				Body: gentest.RandomJSONArticle(2, 3),
			},
			method: "Create",
			query: func() mockQuery {
				m := mockQuery{new(mock.Mock)}
				m.On("Create", mock.Anything, mock.AnythingOfType("entity.Article")).Return(nil).Once()
				return m
			}(),
		},
		{
			desc: "Test if Delete method is invoked on ArticleDeleted event",
			arg: event.Event{
				Type: event.ArticleDeleted,
				Body: func() []byte {
					id, err := json.Marshal(gentest.RandomString(5))
					if err != nil {
						t.Fatalf("Failed to marshal random ID to JSON. Error: %+v", err)
					}
					return id
				}(),
			},
			method: "Delete",
			query: func() mockQuery {
				m := mockQuery{new(mock.Mock)}
				m.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()
				return m
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			db := storage.NewStorage(mockCmd{}, tC.query, nulls.NullLogger{})
			db.CatchUp(tC.arg)

			switch tC.method {
			case "Delete":
				var id string
				err := json.Unmarshal(tC.arg.Body, &id)
				if err != nil {
					t.Errorf("Failed to unmarshal random JSON ID. Error: %+v", err)
					return
				}

				assert.True(t, tC.query.AssertCalled(t, tC.method, mock.Anything, id))

			default:
				var article entity.Article
				err := json.Unmarshal(tC.arg.Body, &article)
				if err != nil {
					t.Errorf("Failed to unmarshal random JSON article. Error: %+v", err)
					return
				}

				assert.True(t, tC.query.AssertCalled(t, tC.method, mock.Anything, article))
			}

			assert.True(t, tC.query.AssertExpectations(t))
		})
	}
}
