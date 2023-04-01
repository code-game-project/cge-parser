# Adapter Library

The adapter package provides helpers for interfacing with a *cge-parser* executable.

## Usage

```go
// Open example.cge.
file, _ := os.Open("example.cge")

// Parse metadata like cge_version from file.
// The returned reader is an io.MultiReader which wraps a buffer containing the
// read bytes for metadata parsing and the supplied reader, which enables choosing
// the cge-parser executable depending on the CGE version, because the data of the reader
// is not consumed.
metadata, reader, err := adapter.ParseMetadata(file)

// Execute the supplied cge-parser executable with the provided configuration
// and collect its output (objects, diagnostics, …) in the response struct.
response, errs := adapter.ParseCGE(reader, "/path/to/cge-parser", adapter.Config{
  IncludeComments: true,
  SendTokens:      true,
  NoObjects:       false,
  DisableWarnings: false,
})
```
