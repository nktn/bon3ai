# Git Workflow

## Commit Message Format

```
<type>: <description>

[optional body]

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `refactor` | Code refactoring (no functional change) |
| `test` | Adding or updating tests |
| `docs` | Documentation changes |
| `chore` | Maintenance tasks |
| `perf` | Performance improvements |
| `ci` | CI/CD changes |

### Examples

```
feat: add keyboard hint for ModeGoTo popup
fix: correct popup width calculation
refactor: extract wrapText helper function
test: add tests for completion filtering
```

## Branch Naming

```
feature/<short-description>
fix/<issue-description>
chore/<task-description>
```

## Pull Request Process

1. **Examine full diff**: Use `git diff main...HEAD` to review all changes
2. **Write summary**: Bullet points of key changes
3. **Include test plan**: Checklist of manual testing steps
4. **Request review**: Use `/codex` for automated code review

### PR Template

```markdown
## Summary
- <bullet points>

## Test plan
- [ ] Manual test item 1
- [ ] Manual test item 2

🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

## Feature Development Cycle

1. **Plan**: Break task into smaller steps
2. **Implement**: Write code with tests
3. **Review**: Run `/codex` for code review
4. **Fix**: Address review feedback
5. **Merge**: Create PR and merge

## Safety Rules

- NEVER force push to `main`
- NEVER use `--no-verify` without explicit reason
- NEVER commit secrets or credentials
- ALWAYS create new commits (avoid `--amend` unless requested)
