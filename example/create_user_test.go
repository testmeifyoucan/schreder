package main

import (
	"encoding/json"
	"testing"

	"github.com/testmeifyoucan/schreder"
)

type CreateUserTest struct{}

func (t *CreateUserTest) Method() string      { return "POST" }
func (t *CreateUserTest) Description() string { return "Test for creating new user API" }
func (t *CreateUserTest) Path() string        { return "/users" }
func (t *CreateUserTest) TestCases() []schreder.TestCase {
	expectedUser := User{
		Name: "New User",
	}
	return []schreder.TestCase{
		{
			Description:      "User created successfully",
			ExpectedHttpCode: 201,
			Headers: schreder.ParamMap{
				"Content-Type": schreder.Param{Value: "application/json;charset=UTF-8"},
			},

			RequestBody: User{
				Name: "New User",
			},

			AssertResponse: func(t *testing.T, expected interface{}, responseBody []byte) bool {
				responseObj := User{}
				if err := json.Unmarshal(responseBody, &responseObj); err != nil {
					t.Errorf("could not unmarshal payload into Business model: %s\nPayload: %s", err.Error(), string(responseBody))
					return false
				}
				if responseObj.ID == 0 {
					t.Errorf("User.ID must not be empty\nPayload: %s", string(responseBody))
					return false
				}

				expectedUser.ID = responseObj.ID
				return schreder.AssertResponse(t, expectedUser, responseBody)
			},

			ExpectedData: User{
				ID:   3,
				Name: "New User",
			},
		},
	}
}
