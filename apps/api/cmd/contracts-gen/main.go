package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/invopop/jsonschema"

	"github.com/novarod/polina/apps/api/internal/domain/mission"
)

func main() {
	out := flag.String("out", "../../packages/contracts/schema/contract.schema.json", "output path for the generated schema")
	flag.Parse()

	data, err := generate()
	if err != nil {
		fmt.Fprintln(os.Stderr, "contracts-gen:", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o750); err != nil {
		fmt.Fprintln(os.Stderr, "contracts-gen:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*out, data, 0o600); err != nil {
		fmt.Fprintln(os.Stderr, "contracts-gen:", err)
		os.Exit(1)
	}
}

func generate() ([]byte, error) {
	r := &jsonschema.Reflector{
		Anonymous:      true,
		ExpandedStruct: true,
		Mapper: func(t reflect.Type) *jsonschema.Schema {
			if t == reflect.TypeOf(json.RawMessage{}) {
				return &jsonschema.Schema{}
			}
			return nil
		},
	}
	s := r.Reflect(&mission.Contract{})
	s.ID = "https://github.com/novarod/polina/packages/contracts/schema/contract.schema.json"
	s.Title = "Contract"

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return nil, err
	}
	data = bytes.ReplaceAll(data, []byte(`"$defs"`), []byte(`"definitions"`))
	data = bytes.ReplaceAll(data, []byte(`"#/$defs/`), []byte(`"#/definitions/`))
	return append(data, '\n'), nil
}
