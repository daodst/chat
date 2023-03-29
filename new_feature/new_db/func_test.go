package new_db

import (
	"github.com/matrix-org/dendrite/new_feature"
	"testing"
)

func TestCheckMortgage(t *testing.T) {
	InitDb("postgresql://postgres:postgres@127.0.0.1:5432/chat?sslmode=disable")
	type args struct {
		userLocal string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			args: args{
				userLocal: "123456",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := new_feature.CheckMortgage(tt.args.userLocal); got != tt.want {
				t.Errorf("CheckMortgage() = %v, want %v", got, tt.want)
			}
		})
	}
}
