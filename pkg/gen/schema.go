package gen

import (
	"bytes"
	"encoding/json"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/pulumi/pulumi/pkg/v3/codegen"
	pschema "github.com/pulumi/pulumi/pkg/v3/codegen/schema"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
)

type OperationMap map[string]string

// CRUDOperationsMap identifies the endpoints to perform
// create, read, update and delete (CRUD) operations.
type CRUDOperationsMap struct {
	// C represents the POST (create) endpoint.
	C *string `json:"c,omitempty"`
	// R represents the GET (read) endpoint.
	R *string `json:"r,omitempty"`
	// U represents the PATCH endpoint.
	U *string `json:"u,omitempty"`
	// D represents the DELETE endpoint.
	D *string `json:"d,omitempty"`

	// P represents the PUT (overwrite/update) endpoint.
	P *string `json:"p,omitempty"`
}

type ProviderMetadata struct {
	// ResourceToOperationMap identifies the endpoint that will
	// handle the CRUD for a given Pulumi resource type token.
	ResourceCRUDMap map[string]*CRUDOperationsMap `json:"crudMap"`
}

type resourceContext struct {
	mod               string
	pkg               *pschema.PackageSpec
	resourceName      string
	openapiComponents openapi3.Components
	visitedTypes      codegen.StringSet
}

func rawMessage(v interface{}) pschema.RawMessage {
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	contract.Assert(err == nil)
	return out.Bytes()
}
