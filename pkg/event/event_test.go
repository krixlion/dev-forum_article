package event

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-article/pkg/helpers/gentest"
)

func TestMakeEvent(t *testing.T) {
	randString := gentest.RandomString(5)
	randArticle := gentest.RandomArticle(1, 2)
	type args struct {
		eType EventType
		data  interface{}
	}
	testCases := []struct {
		name string
		args args
		want Event
	}{
		{
			name: "Test is correctly serializes ArticleDeleted event with simple data",
			args: args{
				eType: ArticleDeleted,
				data:  randString,
			},
			want: Event{
				AggregateId: "article",
				Type:        ArticleDeleted,
				Body: func() []byte {
					data, err := json.Marshal(randString)
					if err != nil {
						panic(err)
					}
					return data
				}(),
				Timestamp: time.Now(),
			},
		},
		{
			name: "",
			args: args{
				eType: ArticleUpdated,
				data:  randArticle,
			},
			want: Event{
				AggregateId: "article",
				Type:        ArticleUpdated,
				Body: func() []byte {
					data, err := json.Marshal(randArticle)
					if err != nil {
						panic(err)
					}
					return data
				}(),
				Timestamp: time.Now(),
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			if got := MakeEvent(tC.args.eType, tC.args.data); !cmp.Equal(got, tC.want, cmpopts.EquateApproxTime(time.Millisecond*5)) {
				t.Errorf("MakeEvent():\n got = %+v\n want = %+v\n diff = %+v\n", got, tC.want, cmp.Diff(got, tC.want))
				return
			}
		})
	}
}
