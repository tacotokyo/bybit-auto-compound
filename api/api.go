package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const baseUrl = "https://api.bybit.com"

var c = http.Client{
	Timeout: time.Second * 3,
}

type BybitApi struct {
	key    string
	secret string
}

func New(key, secret string) *BybitApi {
	return &BybitApi{
		key:    key,
		secret: secret,
	}
}

type walletResult struct {
	RetCode float64 `json:"ret_code"`
	Result  map[string]struct {
		AvailableBalance float64 `json:"available_balance"`
	} `json:"result"`
}

func (b *BybitApi) AvailableBalance(coin string) (float64, error) {
	param := b.getParameterMap()
	coinUpperCase := strings.ToUpper(coin)
	param["coin"] = coinUpperCase
	body, err := b.httpGet("/v2/private/wallet/balance", param)
	if err != nil {
		return 0.0, err
	}
	var m walletResult
	err = json.Unmarshal(body, &m)
	if m.RetCode != 0 {
		return 0.0, fmt.Errorf("Api Wallet Balance result error: %#v", string(body))
	}
	return m.Result[coinUpperCase].AvailableBalance, nil
}

type positionResult struct {
	RetCode float64 `json:"ret_code"`
	Result  struct {
		Symbol              string  `json:"symbol"`
		Side                string  `json:"side"`
		Price               string  `json:"entry_price"`
		Size                float64 `json:"size"`
		Leverage            string  `json:"leverage"`
		EffectiveLeverage   string  `json:"effective_leverage"`
		DeleverageIndicator float64 `json:"deleverage_indicator"`
	} `json:"result"`
}

type Position struct {
	Symbol              string
	Side                string
	Price               float64
	Size                float64
	Leverage            float64
	EffectiveLeverage   float64
	DeleverageIndicator float64
}

func (b *BybitApi) Positions(coin string) (ret Position, err error) {
	param := b.getParameterMap()
	param["symbol"] = strings.ToUpper(coin) + "USD"
	body, err := b.httpGet("/v2/private/position/list", param)
	if err != nil {
		return ret, err
	}
	var m positionResult
	err = json.Unmarshal(body, &m)
	if m.RetCode != 0 {
		return ret, fmt.Errorf("Api Positions List error: %#v", string(body))
	}
	ret = Position{
		Symbol:              m.Result.Symbol,
		Side:                m.Result.Side,
		Size:                m.Result.Size,
		Price:               toFloat64(m.Result.Price),
		Leverage:            toFloat64(m.Result.Leverage),
		EffectiveLeverage:   toFloat64(m.Result.EffectiveLeverage),
		DeleverageIndicator: m.Result.DeleverageIndicator,
	}
	return ret, err
}

type cancelResult struct {
	RetCode float64 `json:"ret_code"`
}

func (b *BybitApi) CancelAll(coin string) error {
	param := b.getParameterMap()
	param["symbol"] = strings.ToUpper(coin) + "USD"
	body, err := b.httpPost("/v2/private/order/cancelAll", param)
	if err != nil {
		return err
	}
	var m cancelResult
	err = json.Unmarshal(body, &m)
	if m.RetCode != 0 {
		return fmt.Errorf("Api Cancel All error: %#v", string(body))
	}
	return err
}

type leverageResult struct {
	RetCode float64 `json:"ret_code"`
}

func (b *BybitApi) LeverageX1(coin string) error {
	param := b.getParameterMap()
	param["symbol"] = strings.ToUpper(coin) + "USD"
	param["leverage"] = 1
	body, err := b.httpPost("/v2/private/position/leverage/save", param)
	if err != nil {
		return err
	}
	var m cancelResult
	err = json.Unmarshal(body, &m)
	if m.RetCode != 0 {
		return fmt.Errorf("Api Position Leverage error: %#v", string(body))
	}
	return err
}

type orderResult struct {
	RetCode float64 `json:"ret_code"`
	Result  struct {
		OrderID string `json:"order_id"`
	} `json:"result"`
}

func (b *BybitApi) Sell(coin string, price float64, size int64) (string, error) {
	param := b.getParameterMap()
	param["side"] = "Sell"
	param["symbol"] = strings.ToUpper(coin) + "USD"
	param["order_type"] = "Limit"
	param["qty"] = size
	param["price"] = price
	param["time_in_force"] = "PostOnly"
	body, err := b.httpPost("/v2/private/order/create", param)
	if err != nil {
		return "", err
	}
	var m orderResult
	err = json.Unmarshal(body, &m)
	if m.RetCode != 0 {
		return "", fmt.Errorf("Api Order error: %#v", string(body))
	}
	return m.Result.OrderID, err
}

type tickerResult struct {
	RetCode float64 `json:"ret_code"`
	Result  []struct {
		Symbol   string `json:"symbol"`
		BidPrice string `json:"bid_price"`
		AskPrice string `json:"ask_price"`
	} `json:"result"`
}

func (b *BybitApi) Price(coin string) (ask, bid float64, err error) {
	param := map[string]interface{}{
		"symbol": strings.ToUpper(coin) + "USD",
	}
	body, err := b.httpGetNoAuth("/v2/public/tickers", param)
	if err != nil {
		return
	}
	var m tickerResult
	err = json.Unmarshal(body, &m)
	if m.RetCode != 0 {
		err = fmt.Errorf("Api Tickers error: %#v", string(body))
		return
	}
	return toFloat64(m.Result[0].AskPrice), toFloat64(m.Result[0].BidPrice), err
}

func toFloat64(v string) float64 {
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return math.NaN()
	}
	return f
}

func (b *BybitApi) getParameterMap() map[string]interface{} {
	return map[string]interface{}{
		"api_key":     b.key,
		"timestamp":   nowUnixMilliSec(),
		"recv_window": "3000",
	}
}

func nowUnixMilliSec() string {
	t := time.Now().UnixNano() / 1e6
	return strconv.FormatInt(t, 10)
}

func (b *BybitApi) httpGet(path string, query map[string]interface{}) ([]byte, error) {
	sign, querystr := b.getSignature(query)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s?%s&sign=%s", baseUrl, path, querystr, sign), nil)
	if err != nil {
		log.Println("http req error:", err)
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		log.Println("http res error:", err)
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("http res body error:", err)
		return nil, err
	}
	return body, nil
}

func (b *BybitApi) httpGetNoAuth(path string, query map[string]interface{}) ([]byte, error) {
	querystr := ""
	for _, k := range sortedKeys(query) {
		querystr += fmt.Sprintf("%v=%v&", k, query[k])
	}
	querystr = querystr[0 : len(querystr)-1]
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s?%s", baseUrl, path, querystr), nil)
	if err != nil {
		log.Println("http req error:", err)
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		log.Println("http res error:", err)
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("http res body error:", err)
		return nil, err
	}
	return body, nil
}

func (b *BybitApi) httpPost(path string, query map[string]interface{}) ([]byte, error) {
	sign, querystr := b.getSignature(query)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s?%s&sign=%s", baseUrl, path, querystr, sign), nil)
	if err != nil {
		log.Println("http req error:", err)
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		log.Println("http res error:", err)
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("http res body error:", err)
		return nil, err
	}
	return body, nil
}

func (b *BybitApi) getSignature(params map[string]interface{}) (string, string) {
	keys := sortedKeys(params)
	_val := ""
	for _, k := range keys {
		_val += fmt.Sprintf("%v=%v&", k, params[k])
	}
	_val = _val[0 : len(_val)-1]
	h := hmac.New(sha256.New, []byte(b.secret))
	h.Write([]byte(_val))
	return hex.EncodeToString(h.Sum(nil)), _val
}

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}
