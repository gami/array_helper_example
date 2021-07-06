package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGenerator_Run(t *testing.T) {
	type args struct {
		dir string
		typ string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "OK",
			args: args{
				dir: "./testdata/chapter2",
				typ: "User",
			},
			want:    "testdata/golden/user_gen.go.golden",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := os.Stat(tt.args.dir)
			if err != nil {
				fmt.Println(err)
			}

			g, err := NewGenerator(tt.args.dir, tt.args.typ)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generator.NewGenerator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got, err := g.Run()
			if (err != nil) != tt.wantErr {
				t.Errorf("Generator.Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			want, err := ioutil.ReadFile(tt.want)
			if err != nil {
				t.Fatal("failed to read golden file")
			}

			if diff := cmp.Diff(got, want); diff != "" {
				t.Errorf("Generator.Run() is not match (-got +want):\n%s", diff)
			}
		})
	}
}
