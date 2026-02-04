---
trigger: model_decision
description: Go Language Coding Guidelines
---

# Coding Guidelines

- These guidelines should be strictly followed. 
- If you're not sure how to apply a rule, ask for clarification.
- If a rule is not applicable, propose an alternative approach and ask for feedback. 

## Go Principles

- **Idiomatic Code**: 
  - Follow [Effective Go](https://golang.org/doc/effective_go).
  - Prioritize readability.
  - Prefer standard library over third-party packages.
  - Do not export symbols that are not used beyond the initial package.
  - Add concise and meaningful comments to all public symbols.

## Project Structure

### Source Files

- All source files should be located in the `internal` directory.
- Use `go mod` to manage dependencies.
- Ask where to put `main` package. 

## Testing Guidelines

- **Isolation**:
  - Use `t.TempDir()` for all filesystem-related tests to ensure isolation and automatic cleanup.
  - Use `defer` to clean up modified environment variables.

- **Assertions**:
  - Use standard `if got != want { t.Errorf(...) }` patterns.
  - Check for existence/non-existence of files using `os.Stat` and `os.IsNotExist`.
  - Validate both success paths and error conditions.

- **Coverage**:
  - Test constructors (`New...`).
  - Test serialization/deserialization cycles.

- **Build & Run**: 
  - When testing the application locally, use `go run`.
  - When checking for compilation errors, use `go build`.
    - Never persist the built binary to VCS (e.g., git).
    - Delete the binary after finished work.
  - When working with a Go submodule (i.e., a separate `go.mod` in a subdirectory), use `go -C <subdir>` to run Go in that subdirectory.
  - The application is not meant to be installed globally.
    - Never run `go install` on the project (i.e., no `go install .`).
    - However, you can `go install <some needed app>` if you need a specific app locally.
    - Always ask for permission before installing anything.

## Architecture
 
- **Separation of Concerns**:
- **Single Responsibility Packages**: Packages should be scoped to a single domain or technical concern. Avoid "god packages" that mix unrelated functionalities.
- **Presentation vs. Business Logic**: The User Interface layer must focus solely on rendering state and capturing user input. It should not perform business rules, data persistence, or external communication directly.
- **Configuration Independence**: Configuration logic (path resolution, environment parsing) must be isolated. Components should have their configuration injected rather than deriving it themselves.
- **Explicit Dependencies**: Components should define their dependencies via their constructor signatures. Avoid hidden dependencies on global state.
- **Event-Driven Decoupling**: Use events or callbacks for communication between layers. Low-level components should remain unaware of the larger application context.
- **Logging**: Prefer structured logging (`slog`) over standard logging (`log`).
- **Dependency Injection**:.

## Development Workflow

- **Atomic Commits**:
  - Work in small, logical steps (e.g., "Step 1: Create the data model," "Step 2: Implement the API endpoint").
  - Commit changes after every step.
  - Aim to keep functions and files focused.