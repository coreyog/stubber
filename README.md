# Stubber

I need to run a report on test coverage in go projects but if a package doesn't have any tests it's excluded from the coverage output and total percentage calculation. A simple fix is to find every folder that contains Go source files but doesn't contain any tests and add a `x_test.go` with a package line and it'll include the folder in the coverage report.

```
# generate coverage file
go test ./... -coverprofile=coverage.out
# parse and cleanly display coverage
go tool cover -func=coverage.out
# alternatively generate and open HTML report
go tool cover -html=coverage.out
```

This is a stupid little project for automating a process I don't want to do by hand. Your mileage may vary.