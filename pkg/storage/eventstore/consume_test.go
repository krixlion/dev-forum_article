package eventstore

import (
	"context"
	"reflect"
	"testing"

	"github.com/krixlion/dev-forum_article/pkg/event"
)

func Test_Consume(t *testing.T) {
	type args struct {
		ctx   context.Context
		in1   string
		eType event.EventType
	}
	tests := []struct {
		name    string
		args    args
		want    <-chan event.Event
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setUpDB()
			defer db.Close()

			got, err := db.Consume(tt.args.ctx, tt.args.in1, tt.args.eType)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.Consume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DB.Consume() = %v, want %v", got, tt.want)
			}
		})
	}
}
