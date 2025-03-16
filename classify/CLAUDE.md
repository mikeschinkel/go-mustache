# GO-MUSTACHE CLASSIFY DEVELOPMENT GUIDE

## Build & Test Commands
- Build: `go build -v ./...`
- Run all tests: `go test -v ./...`
- Run specific test: `go test -v ./test -run TestName`
- Run tests in a specific file: `go test -v ./test/a_test.go`
- Format code: `go fmt ./...`
- Update dependencies: `go mod tidy`
- Check package documentation: `godoc -http=:6060`

## Module Management
- Module path: `github.com/alexkappa/mustache/classify`
- Add dependency: `go get github.com/example/package`
- Update dependency: `go get -u github.com/example/package`
- Local development: Using `replace` directive in go.mod

## Code Style Guidelines
- **Naming**: PascalCase for exported symbols, camelCase for internal symbols
- **Types**: Bitmap-based implementation for Segments (low 4 bits = SegmentType, high 4 bits = TagType)
- **Error Handling**: Return errors with context, use sentinel errors in errors.go
- **Control Flow**: Use goto for early returns and clean error handling
- **Imports**: Standard library first, third-party after blank line
- **Documentation**: Godoc-style comments for all exported functions, types, and constants
- **File Organization**: Related types in their own files (line.go, line_type.go, etc.)

## Project Structure
- Core functionality in `template_classifier.go`
- Types defined in dedicated files (segment.go, tag.go, etc.)
- Tests in the `/test` directory organized by feature
- Lines are classified as EmptyLine, InlineLine, StandaloneLine, TextLine, or WhitespaceLine
- Segments represent parts of a line (TextContent, Whitespace, BeginTag, EndTag, etc.)
- Tag types include sections, partials, comments, variables, and more

## Common Operations
- Create a new classifier: `tc := classify.NewTemplateClassifier(template)`
- Classify a template: `lines, err := tc.Classify()`
- Check line type: `lineType := lines[i].Type`
- Extract segments: `segments := lines[i].Segments`

## Error Handling
- Error pattern: Use sentinel errors from errors.go with ErrArg() for context
- Example: `errors.Join(ErrInvalidTagOpeningType, ErrArg(TagOpenTypeErrArg, ott.String()))`