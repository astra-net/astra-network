package genesis

import "testing"

func TestLocalTestAccounts(t *testing.T) {
	for name, accounts := range map[string][]DeployAccount{
		"AstraV0":      LocalAstraAccounts,
		"AstraV1":      LocalAstraAccountsV1,
		"AstraV2":      LocalAstraAccountsV2,
		"FoundationalV0": LocalFnAccounts,
		"FoundationalV1": LocalFnAccountsV1,
		"FoundationalV2": LocalFnAccountsV2,
	} {
		t.Run(name, func(t *testing.T) { testDeployAccounts(t, accounts) })
	}
}
