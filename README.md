# schreder - write tests and generate docs!

Simple test framework for testing RESTful API. Should be used against running dev instance of the API.

It allows to write tests in declarative way, then declaration can be used to generate documentation in various formats.

## Installation

```
go get -u github.com/testmeifyoucan/schreder
```

## Usage

Please check the test: https://github.com/testmeifyoucan/schreder/blob/master/schreder_test.go

And example: https://github.com/testmeifyoucan/schreder/tree/master/example

## Advantages of such framework

- API tests and documentation with examples from the same box.
- Tests prove that documentation is correct. You can run tests first in order to ensure that you can publish your documentation.
- General purpose tool, works with already existing RESTful API. No need to add or change any line of code in your project, just use `schreder` as external tool.

## Drawbacks

- Documentation covers **tests**, not actual code unfortunately. If tests don't follow the actual code, then documentation may miss something. You need to control test coverage yourself and manually ensure that tests cover all required cases. Or you can think something out and send us a Pull requests :).
- Swagger supports one declaration of request for each HTTP return code (1 declaration for code 200, one for 404 and so on). But what if you have different test cases and all of them produce the same 200 response code? Currently, only first test is used in such situation.
- It's difficult to define all properties of the swagger (like validators, formats) and make the code of the tests readable at the same time. Currently many things provided by swagger are ignored for sake of simplicity of the tests

## Supported documentation formats

- Swagger 2.0
- RAML 0.8
