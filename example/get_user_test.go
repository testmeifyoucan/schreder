package main

import "github.com/testmeifyoucan/schreder"

type GetUserTest struct{}

func (t *GetUserTest) Method() string      { return "GET" }
func (t *GetUserTest) Description() string { return "Test for creating new user API" }
func (t *GetUserTest) Path() string        { return "/users/{user_id}" }
func (t *GetUserTest) TestCases() []schreder.TestCase {
	return []schreder.TestCase{
		{
			Description:      "User returned successfully",
			ExpectedHttpCode: 200,
			PathParams: schreder.ParamMap{
				"user_id": schreder.Param{Value: 2},
			},

			ExpectedData: User{
				ID:   2,
				Name: "Second User",
			},
		},

		{
			Description:      "User was not found",
			ExpectedHttpCode: 404,
			PathParams: schreder.ParamMap{
				"user_id": schreder.Param{Value: -10},
			},
		},
	}
}
