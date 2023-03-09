package go_ethutils

import (
	"reflect"
	"testing"
)

func TestGetAccountFromMnemonic(t *testing.T) {
	type args struct {
		mnemonic string
		index    int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"test_length_15",
			args{
				mnemonic: "index zero inhale insane vapor boss isolate swear pool quarter fuel helmet parent badge interest",
				index:    0,
			},
			"0xfcC887E8574412F824F42D70dea7B4BC5a844015",
		},
		{
			"test_length_18",
			args{
				mnemonic: "this roast more tackle pretty moon security essence fade whip chest awake multiply tag smile west write company",
				index:    1,
			},
			"0x6153cCd54D70ce709C23D194d5A274c075Fa8E2f",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAccountFromMnemonic(tt.args.mnemonic, tt.args.index); !reflect.DeepEqual(got.Address.String(), tt.want) {
				t.Errorf("GetAccountFromMnemonic() = %v, want %v", got, tt.want)
			}
		})
	}
}
