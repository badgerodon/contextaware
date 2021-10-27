package contextaware

//go:generate sh -c "go run ./internal/generate-wrap >wrap_gen.go"

// WrapIO wraps an existing type by wrapping any currently supported interfaces with their corresponding context-aware
// variants. For example if `in` is an `io.Reader`, it will become both an `io.Reader` and a `contextaware.Reader`.
// Since handling all the possible variations of interfaces can't be done using reflection, we instead test a large
// number of cases. This results in a very large multiplication of types, so we only support a limited number of
// interfaces:
//
//   io.Closer
//   io.Reader
//   io.ReaderAt
//   io.ReaderFrom
//   io.Seeker
//   io.Writer
//   io.WriterAt
//   io.WriterTo
//
func WrapIO(in interface{}) (out interface{}) {
	return wrapIO(in)
}
