package schemas

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

//go:embed *.schema.json
var embedFS embed.FS

var schemas map[string]*jsonschema.Schema

var schemaUrls = []string{
	"params.schema.json",
	"action.schema.json",
	"list.schema.json",
	"detail.schema.json",
	"manifest.schema.json",
}

func init() {
	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft7

	for _, url := range schemaUrls {
		schema, err := embedFS.Open(url)
		if err != nil {
			panic(err)
		}

		if err := compiler.AddResource(url, schema); err != nil {
			panic(err)
		}
	}

	schemas = make(map[string]*jsonschema.Schema)
	for _, url := range schemaUrls {
		schema, err := compiler.Compile(url)
		if err != nil {
			panic(err)
		}

		schemas[url] = schema
	}
}

func formatValidationError(ve *jsonschema.ValidationError) string {
	leaf := ve
	for len(leaf.Causes) > 0 {
		leaf = leaf.Causes[0]
	}
	return fmt.Sprintf("%s is not valid: %s", leaf.InstanceLocation, leaf.Message)
}

func validateSchema(schema string, input []byte) error {
	var v interface{}
	if err := json.Unmarshal(input, &v); err != nil {
		return err
	}

	if err := schemas[schema].Validate(v); err != nil {
		if ve, ok := err.(*jsonschema.ValidationError); ok {
			return fmt.Errorf("%s", formatValidationError(ve))
		}
		return err
	}

	return nil
}

func ValidateDetail(input []byte) error {
	return validateSchema("detail.schema.json", input)
}

func ValidateList(input []byte) error {
	return validateSchema("list.schema.json", input)
}

func ValidateManifest(input []byte) error {
	return validateSchema("manifest.schema.json", input)
}
