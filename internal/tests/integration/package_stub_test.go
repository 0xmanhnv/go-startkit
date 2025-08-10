//go:build !integration

package integration

import "testing"

func TestIntegrationStub(t *testing.T) {
	t.Skip("integration tests require -tags=integration")
}
