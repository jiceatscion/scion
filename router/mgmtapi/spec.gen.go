// Package mgmtapi provides primitives to interact with the openapi HTTP API.
//
// Code generated by unknown module path version unknown version DO NOT EDIT.
package mgmtapi

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
