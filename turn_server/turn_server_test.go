package turn_server

import "testing"

func TestAddTmpAuth(t *testing.T) {
	type args struct {
		username string
		realm    string
		password string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				username: "123",
				realm:    "dendrite",
				password: "sdfgfsd",
			},
			wantErr: false,
		},
		{
			name: "test2",
			args: args{
				username: "456",
				realm:    "dendrite",
				password: "sdfgfsd",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := AddTmpAuth(tt.args.username, tt.args.realm, tt.args.password, 5); (err != nil) != tt.wantErr {
				t.Errorf("AddTmpAuth() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDelTmpAuth(t *testing.T) {
	type args struct {
		username string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "1",
			args:    args{username: "123"},
			wantErr: false,
		},
		{
			name:    "2",
			args:    args{username: "456"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DelTmpAuth(tt.args.username); (err != nil) != tt.wantErr {
				t.Errorf("DelTmpAuth() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
