package go_ethutils

import (
	"os"
	"reflect"
	"testing"
)

func TestEthCrypto(t *testing.T) {
	type args struct {
		publicKey  string
		privateKey string
		message    string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test crypto",
			args: args{
				publicKey:  os.Getenv("PUBKEY"),
				privateKey: os.Getenv("PRIVKEY"),
				message:    "hello world",
			},
			want:    "hello world",
			wantErr: false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := EncryptByPubKey(tt.args.publicKey, tt.args.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncryptByPubKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			stringify, err := encrypted.Stringify()
			if err != nil {
				t.Errorf("Stringify() error = %v", err)
				return
			}
			result, err := DecryptByKey(Decode(stringify), tt.args.privateKey)
			if !reflect.DeepEqual(string(result), tt.want) {
				t.Errorf("EncryptByPubKey() got = %v, want %v", result, tt.want)
			} else {
				t.Log(string(result))
			}
		})
	}
}
