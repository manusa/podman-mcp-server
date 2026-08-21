package output_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/manusa/podman-mcp-server/pkg/output"
)

type OutputSuite struct {
	suite.Suite
}

func TestOutput(t *testing.T) {
	suite.Run(t, new(OutputSuite))
}

func (s *OutputSuite) TestNewReturnsJSONOutputForJSONFormat() {
	out := output.New("json")
	s.IsType(&output.JSONOutput{}, out)
}

func (s *OutputSuite) TestNewReturnsTextOutputForTextFormat() {
	out := output.New("text")
	s.IsType(&output.TextOutput{}, out)
}

func (s *OutputSuite) TestNewReturnsTextOutputForUnknownFormat() {
	out := output.New("yaml")
	s.IsType(&output.TextOutput{}, out)
}

func (s *OutputSuite) TestNewReturnsTextOutputForEmptyFormat() {
	out := output.New("")
	s.IsType(&output.TextOutput{}, out)
}

func (s *OutputSuite) TestJSONOutputFormatPassesThroughString() {
	out := &output.JSONOutput{}

	result, err := out.Format("already a string")

	s.Run("no error", func() {
		s.NoError(err)
	})
	s.Run("returns string as-is", func() {
		s.Equal("already a string", result)
	})
}

func (s *OutputSuite) TestJSONOutputFormatMarshalsStruct() {
	out := &output.JSONOutput{}

	result, err := out.Format(map[string]string{"key": "value"})

	s.Run("no error", func() {
		s.NoError(err)
	})
	s.Run("returns valid JSON", func() {
		s.JSONEq(`{"key":"value"}`, result)
	})
}

func (s *OutputSuite) TestJSONOutputFormatMarshalsSlice() {
	out := &output.JSONOutput{}

	result, err := out.Format([]string{"a", "b"})

	s.Run("no error", func() {
		s.NoError(err)
	})
	s.Run("returns valid JSON array", func() {
		s.JSONEq(`["a","b"]`, result)
	})
}

func (s *OutputSuite) TestJSONOutputFormatReturnsErrorForUnmarshalableData() {
	out := &output.JSONOutput{}

	// channels cannot be marshaled to JSON
	_, err := out.Format(make(chan int))

	s.Error(err)
}

func (s *OutputSuite) TestTextOutputFormatPassesThroughString() {
	out := &output.TextOutput{}

	result, err := out.Format("plain text output")

	s.Run("no error", func() {
		s.NoError(err)
	})
	s.Run("returns string as-is", func() {
		s.Equal("plain text output", result)
	})
}

func (s *OutputSuite) TestTextOutputFormatConvertsNonString() {
	out := &output.TextOutput{}

	result, err := out.Format(42)

	s.Run("no error", func() {
		s.NoError(err)
	})
	s.Run("returns string representation", func() {
		s.Equal("42", result)
	})
}
