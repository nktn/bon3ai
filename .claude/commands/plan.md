# /plan - Feature Planning

Plan the implementation of a new feature or change.

## Process

1. **Understand the request**
   - Clarify requirements
   - Identify affected components

2. **Analyze codebase**
   - Find relevant files
   - Understand existing patterns
   - Check for similar implementations

3. **Create implementation plan**
   - List files to modify/create
   - Outline changes for each file
   - Identify potential risks

4. **Output format**

```markdown
## Plan: <feature name>

### Affected Files
- `file1.go`: <brief description of changes>
- `file2.go`: <brief description of changes>

### Implementation Steps
1. <step 1>
2. <step 2>
3. <step 3>

### Testing Strategy
- <test approach>

### Risks/Considerations
- <potential issues>

### Questions
- <unresolved questions, if any>
```

## Guidelines

- Keep plans concise
- Focus on "what" and "where", not "how" in detail
- Identify dependencies between steps
- Always include testing strategy
- List unresolved questions at the end

## Usage

```
/plan Add dark mode support
/plan Implement file search feature
/plan Refactor VCS integration
```
