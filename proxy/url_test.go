package proxy

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type URLTestSuite struct {
	suite.Suite
}

func (s *URLTestSuite) TestAddress() {
	tests := []struct {
		u        *URL
		expected string
	}{
		{
			MustParseURL("http://example.com/"),
			"http://example.com",
		},
	}
	for _, tt := range tests {
		s.Assert().Equal(tt.expected, tt.u.String())
	}

}

func TestURLTestSuite(t *testing.T) {
	suite.Run(t, new(URLTestSuite))
}
