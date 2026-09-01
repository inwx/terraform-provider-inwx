package resource

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-logr/logr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/inwx/terraform-provider-inwx/inwx/internal/api"
)

// testClient returns an api.Client talking to a test server that always answers with responseBody.
func testClient(t *testing.T, responseBody string) *api.Client {
	t.Helper()

	// keep the persistent cookie jar out of the real home directory
	t.Setenv("HOME", t.TempDir())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(responseBody))
	}))
	t.Cleanup(server.Close)

	serverUrl, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("could not parse test server url: %v", err)
	}

	logger := logr.Discard()
	client, err := api.NewClient("user", "password", serverUrl, &logger, false)
	if err != nil {
		t.Fatalf("could not create api client: %v", err)
	}
	return client
}

func testDomainData(t *testing.T) *schema.ResourceData {
	t.Helper()

	data := schema.TestResourceDataRaw(t, DomainResource().Schema, map[string]interface{}{
		"name":   "example.com",
		"period": "1Y",
	})
	data.SetId("example.com")
	return data
}

func TestResourceDomainReadRemovesDeletedDomainFromState(t *testing.T) {
	client := testClient(t, `{"code":2303,"msg":"Object does not exist"}`)
	data := testDomainData(t)

	diags := resourceDomainRead(context.Background(), data, client)

	if diags.HasError() {
		t.Fatalf("expected no error for a domain that does not exist, got: %v", diags)
	}
	if len(diags) != 1 || diags[0].Severity != diag.Warning {
		t.Fatalf("expected exactly one warning, got: %v", diags)
	}
	if data.Id() != "" {
		t.Fatalf("expected the resource to be removed from state, id is still %q", data.Id())
	}
}

func TestResourceDomainReadSetsStatus(t *testing.T) {
	client := testClient(t, `{"code":1000,"resData":{"domain":"example.com","status":"EXPIRED",`+
		`"ns":["ns.inwx.de"],"period":"1Y","renewalMode":"AUTORENEW","transferLock":true,"registrant":123}}`)
	data := testDomainData(t)

	diags := resourceDomainRead(context.Background(), data, client)

	if diags.HasError() {
		t.Fatalf("expected no error, got: %v", diags)
	}
	if data.Id() != "example.com" {
		t.Fatalf("expected the resource to stay in state, got id %q", data.Id())
	}
	if status := data.Get("status").(string); status != "EXPIRED" {
		t.Fatalf("expected status 'EXPIRED' to show up as drift, got %q", status)
	}
}

func TestResourceDomainReadKeepsFailingOnOtherErrors(t *testing.T) {
	client := testClient(t, `{"code":2200,"msg":"Authentication error"}`)
	data := testDomainData(t)

	diags := resourceDomainRead(context.Background(), data, client)

	if !diags.HasError() {
		t.Fatal("expected an error for an api error other than 2303")
	}
	if data.Id() != "example.com" {
		t.Fatalf("expected the resource to stay in state on an unrelated error, got id %q", data.Id())
	}
}

func TestResourceDomainDeleteAcceptsAlreadyDeletedDomain(t *testing.T) {
	client := testClient(t, `{"code":2303,"msg":"Object does not exist"}`)
	data := testDomainData(t)

	diags := resourceDomainDelete(context.Background(), data, client)

	if diags.HasError() {
		t.Fatalf("expected no error when deleting a domain that is already gone, got: %v", diags)
	}
}

func TestResourceDomainDeleteKeepsFailingOnOtherErrors(t *testing.T) {
	client := testClient(t, `{"code":2200,"msg":"Authentication error"}`)
	data := testDomainData(t)

	diags := resourceDomainDelete(context.Background(), data, client)

	if !diags.HasError() {
		t.Fatal("expected an error for an api error other than 2303")
	}
}
