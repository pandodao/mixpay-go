package mixpay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

type Client struct {
	host string
}

func New() *Client {
	return &Client{
		host: "https://api.mixpay.me",
	}
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`

	StatusCode int    `json:"-"`
	Raw        string `json:"-"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("statusCode: %d, raw: %s, code: %d, message: %s", e.StatusCode, e.Raw, e.Code, e.Message)
}

func parseResponse(r *http.Response, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	r.Body = io.NopCloser(bytes.NewReader(body))

	var resp struct {
		Error
		Data        json.RawMessage `json:"data"`
		TimestampMs int64           `json:"timestampMs"`
		Success     bool            `json:"success"`
	}

	resp.StatusCode = r.StatusCode
	if err := json.Unmarshal(body, &resp); err != nil {
		resp.Success = false
		resp.Code = r.StatusCode
		resp.Message = http.StatusText(r.StatusCode)
	}

	if resp.Success {
		if v != nil {
			return json.Unmarshal(resp.Data, v)
		}

		return nil
	}

	resp.Raw = string(body)
	return &resp.Error
}

type CreateOneTimePaymentRequest struct {
	// required
	PayeeId           string
	QuoteAmount       string
	QuoteAssetId      string
	SettlementAssetId string
	OrderId           string

	// optional
	StrictMode       bool
	PaymentAssetId   string
	Remark           string
	ExpireSeconds    int64
	TraceId          string
	SettlementMemo   string
	ReturnTo         string
	FailedReturnTo   string
	CallbackUrl      string
	ExpiredTimestamp int64
}

type CreateOneTimePaymentResponse struct {
	Code    string `json:"code"`
	CodeURL string `json:"codeUrl"`
}

func (r CreateOneTimePaymentResponse) PaymentLink() string {
	return fmt.Sprintf("https://mixpay.me/code/%s", r.Code)
}

func (c *Client) CreateOneTimePayment(ctx context.Context, req CreateOneTimePaymentRequest) (*CreateOneTimePaymentResponse, error) {
	values := url.Values{}
	values.Set("payeeId", req.PayeeId)
	values.Set("quoteAmount", req.QuoteAmount)
	values.Set("quoteAssetId", req.QuoteAssetId)
	values.Set("settlementAssetId", req.SettlementAssetId)
	values.Set("orderId", req.OrderId)

	if req.StrictMode {
		values.Set("strictMode", "true")
	}
	if req.PaymentAssetId != "" {
		values.Set("paymentAssetId", req.PaymentAssetId)
	}
	if req.Remark != "" {
		values.Set("remark", req.Remark)
	}
	if req.ExpireSeconds != 0 {
		values.Set("expireSeconds", strconv.FormatInt(req.ExpireSeconds, 10))
	}
	if req.TraceId != "" {
		values.Set("traceId", req.TraceId)
	}
	if req.SettlementMemo != "" {
		values.Set("settlementMemo", req.SettlementMemo)
	}
	if req.ReturnTo != "" {
		values.Set("returnTo", req.ReturnTo)
	}
	if req.FailedReturnTo != "" {
		values.Set("failedReturnTo", req.FailedReturnTo)
	}
	if req.CallbackUrl != "" {
		values.Set("callbackUrl", req.CallbackUrl)
	}
	if req.ExpiredTimestamp != 0 {
		values.Set("expiredTimestamp", strconv.FormatInt(req.ExpiredTimestamp, 10))
	}

	r, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.host+"/v1/one_time_payment", strings.NewReader(values.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result CreateOneTimePaymentResponse
	if err := parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

type GetPaymentResultRequest struct {
	TraceId string `json:"traceId"`
	OrderId string `json:"orderId"`
	PayeeId string `json:"payeeId"`
}

type GetPaymentResultResponse struct {
	Raw              string `json:"-"`
	Status           string `json:"status"`
	QuoteAmount      string `json:"quoteAmount"`
	QuoteSymbol      string `json:"quoteSymbol"`
	QuoteAssetID     string `json:"quoteAssetId"`
	PaymentAmount    string `json:"paymentAmount"`
	PaymentSymbol    string `json:"paymentSymbol"`
	PaymentAssetID   string `json:"paymentAssetId"`
	Payee            string `json:"payee"`
	PayeeID          string `json:"payeeId"`
	PayeeMixinNumber string `json:"payeeMixinNumber"`
	PayeeAvatarURL   string `json:"payeeAvatarUrl"`
	Txid             string `json:"txid"`
	Date             any    `json:"date"`
	SurplusAmount    string `json:"surplusAmount"`
	SurplusStatus    string `json:"surplusStatus"`
	Confirmations    int    `json:"confirmations"`
	PayableAmount    string `json:"payableAmount"`
	FailureCode      string `json:"failureCode"`
	FailureReason    string `json:"failureReason"`
	ReturnTo         string `json:"returnTo"`
	TraceID          string `json:"traceId"`
}

func (c *Client) GetPaymentResult(ctx context.Context, req GetPaymentResultRequest) (*GetPaymentResultResponse, error) {
	value := url.Values{}
	if req.TraceId != "" {
		value.Set("traceId", req.TraceId)
	}
	if req.OrderId != "" {
		value.Set("orderId", req.OrderId)
	}
	if req.PayeeId != "" {
		value.Set("payeeId", req.PayeeId)
	}

	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.host+"/v1/payments_result"+"?"+value.Encode(), nil)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result GetPaymentResultResponse
	if err := parseResponse(resp, &result); err != nil {
		return nil, err
	}

	raw, _ := io.ReadAll(resp.Body)
	result.Raw = string(raw)

	return &result, nil
}

type Asset struct {
	Name           string          `json:"name"`
	Symbol         string          `json:"symbol"`
	IconUrl        string          `json:"iconUrl"`
	AssetId        string          `json:"assetId"`
	IsAsset        bool            `json:"isAsset"`
	Network        string          `json:"network"`
	IsAvailable    bool            `json:"isAvailable"`
	QuoteSymbol    string          `json:"quoteSymbol"`
	MinQuoteAmount decimal.Decimal `json:"minQuoteAmount"`
	MaxQuoteAmount decimal.Decimal `json:"maxQuoteAmount"`
	ChainAsset     ChainAsset      `json:"chainAsset"`
}

type ChainAsset struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Symbol  string `json:"symbol"`
	IconUrl string `json:"iconUrl"`
}

type ListSettlementAssetsRequest struct {
	PayeeID string
	// QuoteAssetID is assetId of quote cryptocurrency.
	QuoteAssetID string
	// QuoteAmount is the amount of quoteAssetId
	QuoteAmount decimal.Decimal
}

func (c *Client) ListSettlementAssets(ctx context.Context, req *ListSettlementAssetsRequest) ([]*Asset, error) {
	value := url.Values{}
	if req != nil {
		if req.PayeeID != "" {
			value.Set("payeeId", req.PayeeID)
		}

		if req.QuoteAssetID != "" && req.QuoteAmount.IsPositive() {
			value.Set("quoteAssetId", req.QuoteAssetID)
			value.Set("quoteAmount", req.QuoteAmount.String())
		}
	}

	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.host+"/v1/setting/settlement_assets"+"?"+value.Encode(), nil)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var assets []*Asset
	if err := parseResponse(resp, &assets); err != nil {
		return nil, err
	}

	return assets, nil
}
