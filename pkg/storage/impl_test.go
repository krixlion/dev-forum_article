package storage_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-article/pkg/helpers/gentest"
	"github.com/krixlion/dev_forum-article/pkg/storage"
	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/mocks"
	"github.com/krixlion/dev_forum-lib/nulls"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Get(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}

	tests := []struct {
		desc    string
		query   mocks.Query[entity.Article]
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
			query: func() mocks.Query[entity.Article] {
				m := mocks.NewQuery[entity.Article]()
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
			query: func() mocks.Query[entity.Article] {
				m := mocks.NewQuery[entity.Article]()
				m.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(entity.Article{}, errors.New("test err")).Once()
				return m
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			db := storage.NewCQRStorage(mocks.Cmd[entity.Article]{}, tt.query, nulls.NullLogger{}, nulls.NullTracer{})
			got, err := db.Get(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("storage.Get():\n error = %+v\n wantErr = %+v\n", err, tt.wantErr)
				return
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("storage.Get():\n got = %+v\n want = %+v\n", got, tt.want)
				return
			}
			assert.True(t, tt.query.AssertCalled(t, "Get", mock.Anything, tt.args.id))
		})
	}
}
func Test_GetMultiple(t *testing.T) {
	type args struct {
		ctx    context.Context
		offset string
		limit  string
	}

	tests := []struct {
		desc    string
		query   mocks.Query[entity.Article]
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
			query: func() mocks.Query[entity.Article] {
				m := mocks.NewQuery[entity.Article]()
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
			query: func() mocks.Query[entity.Article] {
				m := mocks.NewQuery[entity.Article]()
				m.On("GetMultiple", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return([]entity.Article{}, errors.New("test err")).Once()
				return m
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			db := storage.NewCQRStorage(mocks.Cmd[entity.Article]{}, tt.query, nulls.NullLogger{}, nulls.NullTracer{})
			got, err := db.GetMultiple(tt.args.ctx, tt.args.offset, tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("storage.GetMultiple():\n error = %+v\n wantErr = %+v\n", err, tt.wantErr)
				return
			}

			if !cmp.Equal(got, tt.want, cmpopts.EquateEmpty()) {
				t.Errorf("storage.GetMultiple():\n got = %+v\n want = %+v\n", got, tt.want)
				return
			}

			assert.True(t, tt.query.AssertCalled(t, "GetMultiple", mock.Anything, tt.args.offset, tt.args.limit))
		})
	}
}
func Test_Create(t *testing.T) {
	type args struct {
		ctx     context.Context
		article entity.Article
	}

	tests := []struct {
		desc    string
		cmd     mocks.Cmd[entity.Article]
		args    args
		wantErr bool
	}{
		{
			desc: "Test if method is invoked",
			args: args{
				ctx:     context.Background(),
				article: entity.Article{},
			},

			cmd: func() mocks.Cmd[entity.Article] {
				m := mocks.Cmd[entity.Article]{Mock: new(mock.Mock)}
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
			cmd: func() mocks.Cmd[entity.Article] {
				m := mocks.Cmd[entity.Article]{Mock: new(mock.Mock)}
				m.On("Create", mock.Anything, mock.AnythingOfType("entity.Article")).Return(errors.New("test err")).Once()
				return m
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			db := storage.NewCQRStorage(tt.cmd, mocks.Query[entity.Article]{}, nulls.NullLogger{}, nulls.NullTracer{})
			err := db.Create(tt.args.ctx, tt.args.article)
			if (err != nil) != tt.wantErr {
				t.Errorf("storage.Create():\n error = %+v\n wantErr = %+v\n", err, tt.wantErr)
				return
			}
			assert.True(t, tt.cmd.AssertCalled(t, "Create", mock.Anything, tt.args.article))
		})
	}
}
func Test_Update(t *testing.T) {
	type args struct {
		ctx     context.Context
		article entity.Article
	}

	tests := []struct {
		desc    string
		cmd     mocks.Cmd[entity.Article]
		args    args
		wantErr bool
	}{
		{
			desc: "Test if method is invoked",
			args: args{
				ctx:     context.Background(),
				article: entity.Article{},
			},

			cmd: func() mocks.Cmd[entity.Article] {
				m := mocks.Cmd[entity.Article]{Mock: new(mock.Mock)}
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
			cmd: func() mocks.Cmd[entity.Article] {
				m := mocks.Cmd[entity.Article]{Mock: new(mock.Mock)}
				m.On("Update", mock.Anything, mock.AnythingOfType("entity.Article")).Return(errors.New("test err")).Once()
				return m
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			db := storage.NewCQRStorage(tt.cmd, mocks.Query[entity.Article]{}, nulls.NullLogger{}, nulls.NullTracer{})
			err := db.Update(tt.args.ctx, tt.args.article)
			if (err != nil) != tt.wantErr {
				t.Errorf("storage.Update():\n error = %+v\n wantErr = %+v\n", err, tt.wantErr)
				return
			}
			assert.True(t, tt.cmd.AssertCalled(t, "Update", mock.Anything, tt.args.article))
		})
	}
}
func Test_Delete(t *testing.T) {
	type args struct {
		ctx context.Context
		id  string
	}

	tests := []struct {
		desc    string
		cmd     mocks.Cmd[entity.Article]
		args    args
		wantErr bool
	}{
		{
			desc: "Test if method is invoked",
			args: args{
				ctx: context.Background(),
				id:  "",
			},

			cmd: func() mocks.Cmd[entity.Article] {
				m := mocks.Cmd[entity.Article]{Mock: new(mock.Mock)}
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
			cmd: func() mocks.Cmd[entity.Article] {
				m := mocks.Cmd[entity.Article]{Mock: new(mock.Mock)}
				m.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(errors.New("test err")).Once()
				return m
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			db := storage.NewCQRStorage(tt.cmd, mocks.Query[entity.Article]{}, nulls.NullLogger{}, nulls.NullTracer{})
			err := db.Delete(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("storage.Delete():\n error = %+v\n wantErr = %+v\n", err, tt.wantErr)
				return
			}
			assert.True(t, tt.cmd.AssertCalled(t, "Delete", mock.Anything, tt.args.id))
			assert.True(t, tt.cmd.AssertExpectations(t))
		})
	}
}

func Test_CatchUp(t *testing.T) {
	tests := []struct {
		desc   string
		arg    event.Event
		query  mocks.Query[entity.Article]
		method string
	}{
		{
			desc: "Test if Update method is invoked on ArticleUpdated event",
			arg: event.Event{
				Type: event.ArticleUpdated,
				Body: gentest.RandomJSONArticle(2, 3),
			},
			method: "Update",
			query: func() mocks.Query[entity.Article] {
				m := mocks.NewQuery[entity.Article]()
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
			query: func() mocks.Query[entity.Article] {
				m := mocks.NewQuery[entity.Article]()
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
			query: func() mocks.Query[entity.Article] {
				m := mocks.NewQuery[entity.Article]()
				m.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()
				return m
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			db := storage.NewCQRStorage(mocks.Cmd[entity.Article]{}, tt.query, nulls.NullLogger{}, nulls.NullTracer{})
			db.CatchUp(tt.arg)

			switch tt.method {
			case "Delete":
				var id string
				err := json.Unmarshal(tt.arg.Body, &id)
				if err != nil {
					t.Errorf("Failed to unmarshal random JSON ID. Error: %+v", err)
					return
				}

				assert.True(t, tt.query.AssertCalled(t, tt.method, mock.Anything, id))

			default:
				var article entity.Article
				err := json.Unmarshal(tt.arg.Body, &article)
				if err != nil {
					t.Errorf("Failed to unmarshal random JSON article. Error: %+v", err)
					return
				}

				assert.True(t, tt.query.AssertCalled(t, tt.method, mock.Anything, article))
			}

			assert.True(t, tt.query.AssertExpectations(t))
		})
	}
}
