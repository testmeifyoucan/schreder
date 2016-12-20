package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/go-openapi/spec"

	"github.com/testmeifyoucan/schreder"
)

var outFile = flag.String("out", "", "where to output swagger yaml doc. Runs only if test is successful")

func TestRunApi(t *testing.T) {

	tests := []schreder.Test{
		&GetUserTest{},
		&UpdateUserTest{},
		&CreateUserTest{},
		&DeleteUserTest{},
	}

	runner := schreder.NewRunner("http://localhost:1323", schreder.RunnerConfig{})
	runner.Run(t, tests...)

	if !t.Failed() {
		var writer io.Writer
		if outFile != nil && *outFile != "" {
			fo, err := os.Create(*outFile)
			if err != nil {
				panic(err)
			}
			// close fo on exit and check for its returned error
			defer func() {
				if err := fo.Close(); err != nil {
					panic(err)
				}
			}()

			writer = fo
		} else {
			writer = os.Stdout
		}
		generateSwaggerYAML(t, tests, writer)
	}
}

func generateSwaggerYAML(t *testing.T, tests []schreder.Test, writer io.Writer) {
	seed := spec.Swagger{}
	seed.Host = "localhost"
	seed.Produces = []string{"application/json"}
	seed.Consumes = []string{"application/json"}
	seed.Schemes = []string{"http"}
	seed.Info = &spec.Info{}
	seed.Info.Description = "Example API"
	seed.Info.Title = "Example API"
	seed.Info.Version = "0.1"
	seed.BasePath = "/"

	generator := schreder.NewSwaggerGeneratorYAML(seed)

	doc, err := generator.Generate(tests)
	if err != nil {
		t.Fatalf("could not generate doc: %s", err.Error())
	}

	fmt.Fprint(writer, string(doc))
}
