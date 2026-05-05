package docs

import (
	"strings"
	"testing"

	"github.com/swaggo/swag"
)

func TestSwaggerInfoRegistered(t *testing.T) {
	spec := swag.GetSwagger(SwaggerInfo.InstanceName())
	if spec == nil {
		t.Fatal("expected swagger spec to be registered")
	}

	doc := spec.ReadDoc()
	if !strings.Contains(doc, `"swagger": "2.0"`) {
		t.Fatalf("expected swagger document, got %q", doc[:min(64, len(doc))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
