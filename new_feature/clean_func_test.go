package new_feature

import (
	"freemasonry.cc/chat/new_feature/new_db"
	"testing"
)

func TestCleanExpireHistoryMsgs(t *testing.T) {
	err := new_db.InitDb("postgresql://postgres:postgres@127.0.0.1:5432/chat?sslmode=disable")
	if err != nil {
		t.Error("err init db")
	}
	type args struct {
		interval int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "1", args: args{1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CleanExpireHistoryMsgs(tt.args.interval); (err != nil) != tt.wantErr {
				t.Errorf("CleanExpireHistoryMsgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
