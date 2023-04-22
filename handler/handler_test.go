package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func Test_RequestParser(t *testing.T) {
	var parseRequestTestCases = []struct {
		name       string
		input      string
		expected   *parsedRequest
		expectFail bool
	}{
		{
			name:  "Valid SET command with key, value and expiry time",
			input: `{"command":"SET foo bar EX10"}`,
			expected: &parsedRequest{
				command:       "SET",
				key:           "foo",
				value:         "bar",
				expirySeconds: time.Duration(10),
				condition:     "",
			},
		},
		{
			name:  "Valid SET command with key, value, expiry time and condition",
			input: `{"command":"SET foo bar EX10 NX"}`,
			expected: &parsedRequest{
				command:       "SET",
				key:           "foo",
				value:         "bar",
				expirySeconds: time.Duration(10),
				condition:     "NX",
			},
		},
		{
			name:  "Valid SET command with key, value and condition",
			input: `{"command":"SET foo bar NX"}`,
			expected: &parsedRequest{
				command:       "SET",
				key:           "foo",
				value:         "bar",
				expirySeconds: 0,
				condition:     "NX",
			},
		},
		{
			name:       "Invalid SET command with missing arguments",
			input:      `{"command":"SET"}`,
			expected:   nil,
			expectFail: true,
		},
		{
			name:       "Invalid SET command with invalid arguments",
			input:      `{"command":"SET foo"}`,
			expected:   nil,
			expectFail: true,
		},
		{
			name:  "Valid GET command with key",
			input: `{"command":"GET foo"}`,
			expected: &parsedRequest{
				command: "GET",
				key:     "foo",
			},
		},
		{
			name:       "Invalid GET command with missing key",
			input:      `{"command":"GET"}`,
			expected:   nil,
			expectFail: true,
		},
		{
			name:       "Invalid GET command with extra arguments",
			input:      `{"command":"GET foo bar"}`,
			expected:   nil,
			expectFail: true,
		},
		{
			name:  "Valid QPUSH command with key and values",
			input: `{"command":"QPUSH foo bar baz"}`,
			expected: &parsedRequest{
				command: "QPUSH",
				key:     "foo",
				values:  []string{"bar", "baz"},
			},
		},
		{
			name:       "Invalid QPUSH command with missing arguments",
			input:      `{"command":"QPUSH foo"}`,
			expected:   nil,
			expectFail: true,
		},
		{
			name:  "Valid QPOP command with key",
			input: `{"command":"QPOP foo"}`,
			expected: &parsedRequest{
				command: "QPOP",
				key:     "foo",
			},
		},
		{
			name:       "Invalid QPOP command with missing key",
			input:      `{"command":"QPOP"}`,
			expected:   nil,
			expectFail: true,
		},
		{
			name:       "Invalid QPOP command with extra arguments",
			input:      `{"command":"QPOP foo bar"}`,
			expected:   nil,
			expectFail: true,
		},
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	for _, test := range parseRequestTestCases {
		req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(test.input))
		c.Request = req

		res, err := ParseRequest(c)

		if err != nil && test.expectFail == false {
			t.Errorf("Expecting result got Failure")
		}

		if !reflect.DeepEqual(res, test.expected) {
			t.Errorf("Wrong Result")
		}

	}
}
