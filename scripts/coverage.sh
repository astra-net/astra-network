go test ./... -coverprofile=/tmp/coverage.out;
grep -v "astra-net/AstraNetwork/core" /tmp/coverage.out > /tmp/coverage1.out
go tool cover -func=/tmp/coverage1.out
