run_e2e_tests:
	go test -v ./...

run_test:
	go test ./... -failfast -p 1 -v -count=1 -run TestStaleOperators