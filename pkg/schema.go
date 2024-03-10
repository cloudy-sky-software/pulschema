// Copyright 2022, Cloudy Sky Software.

package pkg

import (
	"bytes"
	"encoding/json"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/pulumi/pulumi/pkg/v3/codegen"
	pschema "github.com/pulumi/pulumi/pkg/v3/codegen/schema"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
)

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

// ProviderMetadata represents metadata used by a provider.
type ProviderMetadata struct {
	// ResourceToOperationMap identifies the endpoint that will
	// handle the CRUD for a given Pulumi resource type token.
	ResourceCRUDMap map[string]*CRUDOperationsMap `json:"crudMap"`

	// AutoNameMap is a map of resource type token and the name
	// property that can be auto-named by the provider.
	AutoNameMap map[string]string `json:"autoNameMap"`

	// SDKToAPINameMap is a map of a property's name in the Pulumi
	// schema to its actual API name. Can be nil.
	SDKToAPINameMap map[string]string `json:"sdkToApiNameMap"`
	// APIToSDKNameMap is the inverse of SDKToAPINameMap.
	APIToSDKNameMap map[string]string `json:"apiToSdkNameMap"`

	// PathParamNameMap is a map of a path param's original name to
	// its Pulumi schema name. Can be nil.
	PathParamNameMap map[string]string `json:"pathParamNameMap"`
}

type resourceContext struct {
	mod               string
	pkg               *pschema.PackageSpec
	resourceName      string
	openapiComponents openapi3.Components
	visitedTypes      codegen.StringSet
	sdkToAPINameMap   map[string]string
	apiToSDKNameMap   map[string]string
	pathParamMap      map[string]string
}

func rawMessage(v interface{}) pschema.RawMessage {
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)
	contract.Assertf(err == nil, "failed to serialize message to JSON")
	return out.Bytes()
}
