# mixpay-go

A Simple Go SDK package for https://mixpay.me

## Example: Create paylink

```go
package main

import (
	"context"
	"fmt"

	"github.com/pandodao/mixpay-go"
)

func main() {
	client := mixpay.New()
	resp, err := client.CreateOneTimePayment(context.Background(), mixpay.CreateOneTimePaymentRequest{
		PayeeId:           payeeId,
		QuoteAssetId:      quoteAssetId,
		QuoteAmount:       amount,
		SettlementAssetId: settlementAssetId,
		OrderId:           orderId,
		TraceId:           traceId,
		CallbackUrl:       callbackUrl,
		ReturnTo:          returnTo,
		FailedReturnTo:    failedReturnTo,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Payment Link: %s\n", resp.PaymentLink())
}
```

## Example: Handle callback

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pandodao/mixpay-go"
)

var client = mixpay.New()

func HandleMixpayCallback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var callbackData struct {
			OrderID string `json:"orderId"`
			TraceID string `json:"traceId"`
			PayeeID string `json:"payeeId"`
		}
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&callbackData); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		// make sure order is exist

		resp, err := client.GetPaymentResult(ctx, mixpay.GetPaymentResultRequest{
			TraceId: callbackData.TraceID,
			OrderId: callbackData.OrderID,
			PayeeId: callbackData.PayeeID,
		})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		fmt.Printf("%+v\n", resp)

		// make sure order is paid
		// update order status in database

		data, _ := json.Marshal(map[string]any{"code": "SUCCESS"})
		w.Write(data)
		w.WriteHeader(http.StatusOK)
	}
}
```

## License
[MIT](https://github.com/pandodao/mixpay-go/blob/main/LICENSE)
