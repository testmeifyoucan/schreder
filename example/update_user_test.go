package main

import "github.com/testmeifyoucan/schreder"

type UpdateUserTest struct{}

func (t *UpdateUserTest) Method() string      { return "PATCH" }
func (t *UpdateUserTest) Description() string { return "Test for creating new user API" }
func (t *UpdateUserTest) Path() string        { return "/users/{user_id}" }
func (t *UpdateUserTest) TestCases() []schreder.TestCase {
	return []schreder.TestCase{
		{
			Description:      "User updated successfully",
			ExpectedHttpCode: 200,
			Headers: schreder.ParamMap{
				"Content-Type": schreder.Param{Value: "application/json;charset=UTF-8"},
			},
			PathParams: schreder.ParamMap{
				"user_id": schreder.Param{Value: 1},
			},

			RequestBody: User{
				Name: "I Am Updated!",
			},

			ExpectedData: User{
				ID:   1,
				Name: "I Am Updated!",
			},
		},
	}
}
