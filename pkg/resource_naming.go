package pkg

import (
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/golang/glog"
)

func getModuleFromPath(path string, useParentResourceAsModule bool) string {
	if useParentResourceAsModule {
		parentPath := getParentPath(path)
		parentParts := strings.Split(strings.TrimPrefix(parentPath, pathSeparator), pathSeparator)
		return strings.ToLower(parentParts[len(parentParts)-1])
	}

	parts := strings.Split(strings.TrimPrefix(path, pathSeparator), pathSeparator)

	// If the first item in parts is not a version number prefix, then
	// return as-is.
	if !versionRegex.Match([]byte(parts[0])) {
		return strings.ToLower(parts[0])
	}

	// Otherwise, we should use a versioned module.
	return strings.ToLower(parts[1]) + pathSeparator + parts[0]
}

func getParentPath(path string) string {
	parts := strings.Split(strings.TrimPrefix(path, pathSeparator), pathSeparator)
	lastPathPart := parts[len(parts)-1]
	if !strings.HasPrefix(lastPathPart, "{") && !strings.HasSuffix(lastPathPart, "}") {
		return path
	}

	// Skip the last path part which contains a path parameter.
	return pathSeparator + strings.Join(parts[0:len(parts)-1], pathSeparator)
}

// index returns the index of the element `toFind`
// in the slice `slice`. Returns -1 if not found.
func index(slice []string, toFind string) int {
	for i, s := range slice {
		if s == toFind {
			return i
		}
	}

	return -1
}

func getResourceTitleFromOperationID(originalOperationID, method string, isSeparatedByTypeSpecNamespace bool) string {
	var replaceKeywords []string

	switch method {
	case http.MethodDelete:
		replaceKeywords = append(replaceKeywords, "delete", "destroy", "remove")
	case http.MethodGet:
		replaceKeywords = append(replaceKeywords, "get", "list")
	case http.MethodPatch:
		replaceKeywords = append(replaceKeywords, "patch", "update")
	case http.MethodPost:
		replaceKeywords = append(replaceKeywords, "add", "create", "post", "put", "set")
	case http.MethodPut:
		replaceKeywords = append(replaceKeywords, "add", "create", "put", "set", "update", "replace")
	}

	result := originalOperationID

	// TypeSpec-generated operations can have an operation ID separated by the namespace
	// the operation is defined in.
	if isSeparatedByTypeSpecNamespace {
		parts := strings.Split(originalOperationID, "_")
		result = parts[len(parts)-1]
	} else if strings.Contains(originalOperationID, "_") {
		parts := strings.Split(originalOperationID, "_")
		result = parts[0]
		for _, p := range parts[1:] {
			result += ToPascalCase(p)
		}
	}

	for _, v := range replaceKeywords {
		result = strings.ReplaceAll(result, v, "")
		result = strings.ReplaceAll(result, ToPascalCase(v), "")
	}

	resourceTitle := ToPascalCase(result)

	glog.Infof("converted operation ID %s to resource title %s\n", originalOperationID, resourceTitle)

	return resourceTitle
}

func getResourceTitleFromRequestSchema(schemaName string, schemaRef *openapi3.SchemaRef) string {
	if schemaRef.Value.Title != "" {
		return ToPascalCase(schemaRef.Value.Title)
	}

	var title string

	if strings.Contains(schemaName, "_") {
		title = ToPascalCase(schemaName)
	} else {
		parts := strings.Split(schemaName, "_")
		result := parts[0]
		for _, p := range parts[1:] {
			result += ToPascalCase(p)
		}

		title = result
	}

	return title
}

// ensureIDHierarchyInRequestPath ensures that the IDs in the path
// segment following the rules:
//
// 1. parent resource id param should not be called `id`.
//
// 2. the sub-resource, if present, should use `id` if the path
// variable represents an ID.
func ensureIDHierarchyInRequestPath(path string, pathItem *openapi3.PathItem) string {
	segments := strings.Split(path, pathSeparator)
	numSegments := len(segments)

	pathParamTransformationsMap := make(map[string]string)

	updatePathParamNames := func(params openapi3.Parameters) {
		for _, param := range params {
			if param.Value.In != parameterLocationPath {
				continue
			}

			if transformedName, ok := pathParamTransformationsMap["{"+param.Value.Name+"}"]; ok {
				param.Value.Name = transformedName
			}
		}
	}

	for i, segment := range segments {
		if segment == "" || !strings.Contains(segment, "{") {
			continue
		}

		var transformedParam string

		// Satisfies rule #1.
		if i != numSegments-1 && segment == "id" {
			parentResource := segments[i-1]
			transformedParam = parentResource
			if strings.Contains(parentResource, "_") {
				transformedParam += "_id"
			} else {
				transformedParam += "Id"
			}
		} else if i == numSegments-1 && segment != "id" && strings.Contains(strings.ToLower(segment), "id") {
			// Satisfies rule #2.
			transformedParam = "id"
		}

		if transformedParam == "" {
			continue
		}

		pathParamTransformationsMap[segments[i]] = transformedParam
		segments[i] = "{" + transformedParam + "}"
	}

	if pathItem.Delete != nil && len(pathItem.Delete.Parameters) > 0 {
		updatePathParamNames(pathItem.Delete.Parameters)
	}
	if pathItem.Get != nil && len(pathItem.Get.Parameters) > 0 {
		updatePathParamNames(pathItem.Get.Parameters)
	}
	if pathItem.Patch != nil && len(pathItem.Patch.Parameters) > 0 {
		updatePathParamNames(pathItem.Patch.Parameters)
	}
	if pathItem.Put != nil && len(pathItem.Put.Parameters) > 0 {
		updatePathParamNames(pathItem.Put.Parameters)
	}
	if pathItem.Post != nil && len(pathItem.Post.Parameters) > 0 {
		updatePathParamNames(pathItem.Post.Parameters)
	}

	return strings.Join(segments, pathSeparator)
}
