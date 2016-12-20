package schreder

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/jingweno/go-sawyer/hypermedia"
)

// Test defines an interface of API test definition.
//
// Responsibility of api test is to provide some meta data about API endpoint and
// collection of test cases that describe concrete usage of that endpoint.
type Test interface {
	Method() string
	Description() string
	Path() string
	TestCases() []TestCase
}

// ISetuppable defines interface for tests that have setup logic
//
// If you need to run some preparation logic - this is the right place to do
type Setuppable interface {
	SetUp() error
}

// ITeardownable defines interface for tests that have teardown logic
//
// Good place to put some cleanup logic if you need to
type Teardownable interface {
	TearDown() error
}

// AssertResponseFunc defines function that asserts that expected object equals to
// given response body
type AssertResponseFunc func(t *testing.T, expected interface{}, responseBody []byte) bool

// TestCase provides use case of some API endpoint: input and expected output.
//
// Test case knows nothing about API endpoint itself. Path of an API endpoint and
// its description must be provided by Test.
//
// Ideally each different error that API endpoint can return should
// be described by a test case.
type TestCase struct {
	Description string

	Headers     ParamMap
	QueryParams ParamMap
	PathParams  ParamMap
	RequestBody interface{}

	ExpectedHttpCode int
	ExpectedHeaders  map[string]string
	ExpectedData     interface{}

	// AssertResponse is a custom assertion logic that can be used
	// instead of ExpectedData. If provided, it is fully responsible
	// for processing of API response payload and assertion with
	// expected data.
	AssertResponse AssertResponseFunc
}

type ParamMap map[string]Param

// Param defines a parameter that is used in headers or URL query
// of the API request
type Param struct {
	Value       interface{}
	Required    bool
	Description string
}

// Url generates full URL to API endpoint for given test case.
// urlpath must provide full URL to the endpoint with no query parameters.
func (testCase *TestCase) Url(urlpath string) (string, error) {
	// TODO: what if url template contains something, but it's not provided by test case?
	sawyerHyperlink := hypermedia.Hyperlink(urlpath)
	params := hypermedia.M{}
	for name, p := range testCase.PathParams {
		params[name] = p.Value
	}

	u, err := sawyerHyperlink.Expand(params)
	if err != nil {
		return "", err
	}
	if testCase.QueryParams != nil {
		query := url.Values{}
		for key, param := range testCase.QueryParams {
			valueStr := fmt.Sprintf("%v", param.Value)
			query.Set(key, valueStr)
		}
		u.RawQuery = query.Encode()
	}

	return u.String(), nil
}
