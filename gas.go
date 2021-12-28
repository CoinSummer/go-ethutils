package ethutils

import (
	"github.com/imroc/req"
	"math/big"
)

type GasPrice struct {
	Rapid    *big.Int `json:"rapid"`
	Fast     *big.Int `json:"fast"`
	Standard *big.Int `json:"standard"`
	Low      *big.Int `json:"slow"`
}

type gasNowResp struct {
	Code int       `json:"code"`
	Data *GasPrice `json:"data"`
}

func GetSuggestGasPrice() *GasPrice {
	resp, err := req.Get("https://etherchain.org/api/gasnow")
	if err != nil {
		return nil
	}
	r := &gasNowResp{}
	err = resp.ToJSON(r)
	if err != nil || r.Code != 200 {
		return nil
	}
	return r.Data
}
