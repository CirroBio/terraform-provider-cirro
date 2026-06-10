package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/cirro-bio/terraform-provider-cirro/internal/provider"
)

// testAccProtoV6ProviderFactories is used in acceptance tests.
// Set CIRRO_BASE_URL, CIRRO_CLIENT_ID, and CIRRO_CLIENT_SECRET before running.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"cirro": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func TestProvider_instantiates(t *testing.T) {
	_ = testAccProtoV6ProviderFactories
}
