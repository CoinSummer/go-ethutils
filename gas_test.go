package ethutils

import (
	"reflect"
	"testing"
)

func TestGetSuggestGasPrice(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{
			name: "normal",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSuggestGasPrice(); !reflect.DeepEqual(got.Rapid.Int64() > 0, tt.want) {
				t.Errorf("GetSuggestGasPrice() = %v, want %v", got, tt.want)
			}
		})
	}
}
