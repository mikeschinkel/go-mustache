# GO-MUSTACHE DEVELOPMENT GUIDE

## Build & Test Commands
- Build: `go build -v ./...`
- Run all tests: `go test -v ./...`
- Run single test: `go test -v ./test -run TestName`
- Run specific spec test: `go test -v ./test -run TestSpec/comments`
- Skip spec tests: `SKIP_SPECS=yes go test -v ./...`

## Code Style Guidelines
- **Project Structure**: Core code in root, tests in `test/` directory
- **Naming**: PascalCase for exported types/funcs, camelCase for internal
- **Error Handling**: Return errors, no panics in production code
- **Types**: Interface-based design with `Node` as core abstraction
- **Imports**: Standard Go organization (stdlib first, third-party after blank line)
- **Documentation**: Godoc-style comments for all exported symbols
- **Testing**: Example-driven tests, spec conformance testing
- **Pattern**: Each node type has its own file with consistent naming (`*_node.go`)

## Project Features
- Functional options pattern for configuration
- Full Mustache spec conformance
- Support for struct field tags, methods, and pointer fields