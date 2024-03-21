package go_ethutils

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
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
				message:    os.Getenv("MESSAGE"),
			},
			want:    os.Getenv("MESSAGE"),
			wantErr: false,
		},
		{
			name: "test crypto 1111",
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
			//encrypted, err := EncryptByPubKey(tt.args.publicKey, tt.args.message)
			//if (err != nil) != tt.wantErr {
			//	t.Errorf("EncryptByPubKey() error = %v, wantErr %v", err, tt.wantErr)
			//	return
			//}
			//
			//marshal, err := json.Marshal(encrypted)
			//if err != nil {
			//	t.Errorf("json.Marshal() error = %v", err)
			//	return
			//}
			//t.Log(string(marshal))
			//
			//stringify, err := encrypted.Stringify()
			//if err != nil {
			//	t.Errorf("Stringify() error = %v", err)
			//	return
			//}
			//
			//t.Log("code: ", stringify)

			iv, err := hexutil.Decode("0x3f7d07b4cbb31909de526fab97cb0a33")
			if err != nil {
				t.Errorf("hexutil.Decode() error = %v", err)
				return
			}
			ephemPublicKey, err := hexutil.Decode("0x0479049255f0c294cf0000ba57451b449a5cdb3e6c41c0c0df30d7af9e35da232a948bd4ca792aaefe49fdcb05dbaca7df308eeece231aaffb4f7e01738c335ac7")
			if err != nil {
				t.Errorf("hexutil.Decode() error = %v", err)
				return
			}

			ciphertext, _ := hexutil.Decode("0xf89161182fe22a7b644976d1a53a7c81")
			mac, _ := hexutil.Decode("0x6227f04f023a742a6663ec5232638001a7b939377a772ce4d078216297a11c98")
			opt := &EncryptOption{
				Iv:             iv,
				EphemPublicKey: ephemPublicKey,
				Ciphertext:     ciphertext,
				Mac:            mac,
			}

			//result, _ := DecryptByKey(Decode(stringify), tt.args.privateKey)
			result, err := DecryptByKey(opt, tt.args.privateKey)
			if err != nil {
				t.Errorf("DecryptByKey() error = %v", err)
				return
			}
			if !reflect.DeepEqual(string(result), tt.want) {
				t.Errorf("EncryptByPubKey() got = %v, want %v", result, tt.want)
			} else {
				t.Log(string(result))
			}
		})
	}
}
