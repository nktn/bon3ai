# Development Guidelines

## Plan Mode

- Make the plan extremely concise. Sacrifice grammar for the sake of concision.
- At the end of each plan, give me a list of unresolved questions to answer, if any.

## Writing Tests

Always create corresponding tests when implementing new features:

1. Add tests to the `*_test.go` file that corresponds to your implementation
2. Run `make test` to verify all tests pass
3. Write tests especially for:
   - New functions and methods
   - Changes to existing function behavior
   - Edge cases and error handling

## State Machine Updates

When making changes to InputMode, always:

1. Update `.claude/rules/architecture.md` state machine diagram
2. Ensure consistency with `README.md` keybindings table
