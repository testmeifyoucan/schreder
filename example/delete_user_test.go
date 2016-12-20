package main

import "github.com/testmeifyoucan/schreder"

type DeleteUserTest struct{}

func (t *DeleteUserTest) Method() string      { return "DELETE" }
func (t *DeleteUserTest) Description() string { return "Test for creating new user API" }
func (t *DeleteUserTest) Path() string        { return "/users/{user_id}" }
func (t *DeleteUserTest) TestCases() []schreder.TestCase {
	return []schreder.TestCase{
		{
			Description:      "User deleted successfully",
			ExpectedHttpCode: 204,
			PathParams: schreder.ParamMap{
				"user_id": schreder.Param{Value: 3},
			},
		},
	}
}
