package checktypes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestByAssettype(t *testing.T) {
	// Mock server returning some sample checktypes
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"checktypes": [
				{
					"assets": ["Hostname", "IP"],
					"name": "vulcan-nessus"
				},
				{
					"assets": ["IP"],
					"name": "vulcan-tls"
				}
			]
		}`))
	}))
	defer ts.Close()

	client := New(ts.URL)
	assetMap, err := client.ByAssettype(context.Background())
	if err != nil {
		t.Fatalf("ByAssettype() error = %v", err)
	}

	if got := len(assetMap); got != 2 {
		t.Errorf("expected 2 keys in returned map, got %d", got)
	}

	if got, want := len(assetMap["Hostname"]), 1; got != want {
		t.Errorf("expected Hostname to have %d checktype, got %d", want, got)
	}

	if got, want := assetMap["Hostname"][0], "vulcan-nessus"; got != want {
		t.Errorf("expected Hostname[0] to be %q, got %q", want, got)
	}

	if got, want := len(assetMap["IP"]), 2; got != want {
		t.Errorf("expected IP to have %d checktypes, got %d", want, got)
	}
}
