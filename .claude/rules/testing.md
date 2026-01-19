# Testing Guidelines

## Test Requirements

- All new features must have corresponding tests
- Bug fixes should include regression tests
- Target: meaningful coverage for critical paths

## Test File Structure

```
foo.go           # Implementation
foo_test.go      # Tests for foo.go
```

## Test Naming

```go
func TestFunctionName(t *testing.T) {...}
func TestFunctionName_SpecificCase(t *testing.T) {...}

// Table-driven tests
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"empty input", "", ""},
        {"normal case", "foo", "bar"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic
        })
    }
}
```

## Running Tests

```bash
# All tests
make test

# Specific test
go test -v -run TestName

# With coverage
go test -cover ./...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Test Categories

### Unit Tests
- Test individual functions in isolation
- Use mocks for external dependencies
- Fast execution (< 1 second per test)

### Integration Tests
- Test component interactions
- Use `t.TempDir()` for file system tests
- Clean up resources in `t.Cleanup()`

## Best Practices

1. **Arrange-Act-Assert** pattern
2. **Use `t.Helper()`** for test helper functions
3. **Avoid test interdependence**
4. **Test edge cases**: empty input, nil, boundaries
5. **Use descriptive failure messages**

```go
if got != want {
    t.Errorf("FunctionName(%q) = %q, want %q", input, got, want)
}
```

## When Tests Fail

1. Read the error message carefully
2. Check if test expectations are correct
3. Fix implementation, not test (unless test is wrong)
4. Re-run to verify fix
