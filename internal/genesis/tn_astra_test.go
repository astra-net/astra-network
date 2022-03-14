package genesis

import "testing"

func TestTNAstraAccounts(t *testing.T) {
	testDeployAccounts(t, TNAstraAccounts)
}

func TestTNFoundationalAccounts(t *testing.T) {
	testDeployAccounts(t, TNFoundationalAccounts)
}
