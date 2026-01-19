# Coding Style Guide (Go)

## Core Principles

1. **Simplicity First**: Write simple, readable code. Avoid over-engineering.
2. **Idiomatic Go**: Follow Go conventions and `gofmt` standards.
3. **Small Functions**: Keep functions under 50 lines. Extract if longer.
4. **Small Files**: Target 200-400 lines per file, max 800 lines.

## Naming Conventions

- **Packages**: Short, lowercase, single-word names
- **Exported**: `PascalCase` for public types/functions
- **Unexported**: `camelCase` for private types/functions
- **Interfaces**: Single-method interfaces end with `-er` (e.g., `Reader`, `Writer`)
- **Receivers**: Short, 1-2 letter names (e.g., `m` for `Model`)

## Error Handling

```go
// Always handle errors explicitly
result, err := doSomething()
if err != nil {
    return fmt.Errorf("context: %w", err)
}

// Use sentinel errors for expected conditions
var ErrNotFound = errors.New("not found")
```

## Struct Design

- Group related fields together
- Add comments for non-obvious fields
- Use pointer receivers for methods that modify state

## File Organization

```
package main

import (
    "standard-library"

    "external-packages"
)

// Constants
const (...)

// Types
type Foo struct {...}

// Methods
func (f *Foo) Method() {...}

// Functions
func helperFunction() {...}
```

## Quality Checklist

Before completing any change, verify:

- [ ] `go fmt` applied
- [ ] `go vet` passes
- [ ] No magic numbers (use named constants)
- [ ] Error messages are descriptive
- [ ] No debug statements left (fmt.Println, log.Println for debugging)
- [ ] Functions have single responsibility
- [ ] Comments explain "why", not "what"
