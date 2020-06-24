package schreder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/elgris/jsondiff"
	"github.com/stretchr/testify/assert"
)

// ITestRunner is responsible for
// - receive a suite or suites
// - prepare some conditions to run suites (like setup frisby or whatsoever)
// - run tests of the suite
type ITestRunner interface {
	Run(tests []Test, t *testing.T)
}

// INameable is an interface that defines that entity can provide its name.
// Used by test runner to define name of the test.
type INameable interface {
	Name() string
}

// IHttpClient defines an interface of HTTP client that can fire HTTP requests
// and return responses.
type IHttpClient interface {
	Do(req *http.Request) (resp *http.Response, err error)
}

// IHttpClientFunc implements IHttpClient in a functional way
type IHttpClientFunc func(req *http.Request) (resp *http.Response, err error)

func (f IHttpClientFunc) Do(req *http.Request) (resp *http.Response, err error) { return f(req) }

type httpRunner struct {
	DefaultHeaders map[string]string
	BaseUrl        string
	HttpClient     IHttpClient
}

// RunnerConfig contains list of possible options that can be used to initialize
// the Runner
type RunnerConfig struct {
	DefaultHeaders map[string]string
	HttpClient     IHttpClient
}

// NewRunner creates new instance of HTTP runner
func NewRunner(baseUrl string, config RunnerConfig) *httpRunner {
	r := &httpRunner{
		DefaultHeaders: make(map[string]string),
		BaseUrl:        baseUrl,
		HttpClient:     &http.Client{},
	}

	if config.DefaultHeaders != nil {
		r.DefaultHeaders = config.DefaultHeaders
	}
	if config.HttpClient != nil {
		r.HttpClient = config.HttpClient
	}

	return r
}

func (r *httpRunner) Run(t *testing.T, tests ...Test) {
	for _, test := range tests {
		testName := extractTestName(test)
		// setup test
		if setuppable, ok := test.(Setuppable); ok {
			t.Logf("setting up test '%s'(%s)...", testName, test.Description())

			if err := setuppable.SetUp(); err != nil {
				t.Errorf("error setting up test '%s'(%s): %s",
					testName, test.Description(), err.Error())

				continue
			}
		}

		// run test
		for caseIndex, testCase := range test.TestCases() {
			t.Logf("running test '%s'(%s), case %d", testName, testCase.Description, caseIndex+1)
			r.runTest(t, testCase, test.Method(), test.Path())
		}

		// teardown test
		if teardownable, ok := test.(Teardownable); ok {
			t.Logf("tearing down test '%s'(%s)...", testName, test.Description())

			if err := teardownable.TearDown(); err != nil {
				t.Errorf("error cleaning up after a test '%s'(%s): %s",
					testName, test.Description(), err.Error())
			}
		}
	}
}

func (r *httpRunner) encode(obj interface{}) ([]byte, error) {
	// TODO: make it configurable
	return json.Marshal(obj)
}

func (r *httpRunner) runTest(t *testing.T, testCase TestCase, method, path string) {
	urlstring := r.BaseUrl + path
	url, err := testCase.Url(urlstring)
	if !assert.NoError(t, err, "could not prepare an url") {
		return
	}

	// TODO: prepare body
	var req *http.Request
	if testCase.RequestBody != nil {
		encoded, err := r.encode(testCase.RequestBody)
		if !assert.NoError(t, err, "could not encode body") {
			return
		}

		requestBody := bytes.NewBuffer(encoded)
		req, err = http.NewRequest(method, url, requestBody)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if !assert.NoError(t, err, "could not create HTTP request") {
		return
	}

	for name, value := range r.DefaultHeaders {
		req.Header.Set(name, value)
	}
	for name, param := range testCase.Headers {
		if stringValue, ok := param.Value.(string); ok {
			req.Header.Set(name, stringValue)
		} else {
			req.Header.Set(name, fmt.Sprintf("%v", param.Value))
		}
	}

	resp, err := r.HttpClient.Do(req)
	if !assert.NoError(t, err, "failed sending a request") {
		return
	}

	if !assert.NotNil(t, resp, "request to '%s' returned nil response", urlstring) {
		return
	}

	var responseBody []byte
	if resp.Body != nil {
		defer resp.Body.Close()

		responseBody, err = ioutil.ReadAll(resp.Body)
		if !assert.NoError(t, err) {
			return
		}
	}

	if !assert.Equal(t, testCase.ExpectedHttpCode, resp.StatusCode) {
		t.Logf("body received: %s", string(responseBody))

		return
	}

	// asserting headers
	if testCase.ExpectedHeaders != nil {
		for header, value := range testCase.ExpectedHeaders {
			if !assert.Equal(t, value, resp.Header.Get(header)) {
				t.Logf("body received: %s", string(responseBody))

				return
			}
		}
	}

	if testCase.AssertResponse != nil {
		testCase.AssertResponse(t, testCase.ExpectedData, responseBody)
	} else {
		AssertResponse(t, testCase.ExpectedData, responseBody)
	}
}

// AssertResponse checks that given expected object contains the same data
// as provided responseBody.
func AssertResponse(t *testing.T, expected interface{}, responseBody []byte) bool {
	if expected != nil {
		expectedData := decodeExpected(expected)
		actualData := decodeResponse(responseBody)

		diff := jsondiff.Compare(expectedData, actualData)
		if !diff.IsEqual() {
			return assert.Fail(t, string(jsondiff.Format(diff)), "request and response are not equal")
		}
		return true
	}

	return assert.Empty(t, string(responseBody), "expected empty response")
}

// decodeExpected processes expected data into representation used for comparison to actual data
// (converts object to map, does some type conversion)
func decodeExpected(data interface{}) interface{} {
	if decoded, err := objToJsonMap(data); err == nil {
		return decoded
	}

	return data
}

// decodeResponse processes response data into representation used for comparison to expected data
func decodeResponse(data []byte) interface{} {
	var dataMap map[string]interface{}
	if err := json.Unmarshal(data, &dataMap); err == nil {
		return dataMap
	}

	return string(data)
}

func extractTestName(value interface{}) string {
	if nameable, ok := value.(INameable); ok {
		return nameable.Name()
	}
	return reflect.TypeOf(value).String()
}

func objToJsonMap(obj interface{}) (map[string]interface{}, error) {
	js, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{}

	err = json.Unmarshal(js, &result)
	return result, err
}
