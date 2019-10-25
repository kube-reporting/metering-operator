package ini

import (
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws/awserr"
)

// OpenFile takes a path to a given ***REMOVED***le, and will open  and parse
// that ***REMOVED***le.
func OpenFile(path string) (Sections, error) {
	f, err := os.Open(path)
	if err != nil {
		return Sections{}, awserr.New(ErrCodeUnableToReadFile, "unable to open ***REMOVED***le", err)
	}
	defer f.Close()

	return Parse(f)
}

// Parse will parse the given ***REMOVED***le using the shared con***REMOVED***g
// visitor.
func Parse(f io.Reader) (Sections, error) {
	tree, err := ParseAST(f)
	if err != nil {
		return Sections{}, err
	}

	v := NewDefaultVisitor()
	if err = Walk(tree, v); err != nil {
		return Sections{}, err
	}

	return v.Sections, nil
}

// ParseBytes will parse the given bytes and return the parsed sections.
func ParseBytes(b []byte) (Sections, error) {
	tree, err := ParseASTBytes(b)
	if err != nil {
		return Sections{}, err
	}

	v := NewDefaultVisitor()
	if err = Walk(tree, v); err != nil {
		return Sections{}, err
	}

	return v.Sections, nil
}
