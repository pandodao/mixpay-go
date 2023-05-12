package mixpay

import (
	"context"
	"testing"
)

func TestClient_ListSettlementAssets(t *testing.T) {
	ctx := context.Background()
	client := New()

	assets, err := client.ListSettlementAssets(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, asset := range assets {
		t.Log(asset.Symbol)
		t.Log(asset.AssetId)
	}
}
