package dto

import (
	"testing"

	"eshkere/internal/service"
)

func TestPartnerDictionaryDTOConversions(t *testing.T) {
	items := ToDictionaryItemResponses([]service.DictionaryItem{{Code: "RU", Name: "Russia"}})
	if len(items) != 1 || items[0].Code != "RU" {
		t.Fatalf("unexpected dictionary items: %+v", items)
	}

	blockTypes := ToBlockTypeDictionaryItemResponses([]service.BlockTypeDictionaryItem{{
		Code: "banner", Name: "Banner", Description: "Banner ad", Platforms: []string{"web"},
	}})
	if len(blockTypes) != 1 || blockTypes[0].Platforms[0] != "web" {
		t.Fatalf("unexpected block types: %+v", blockTypes)
	}

	tree := ToGeoTreeNodeResponses([]*service.GeoTreeNode{{
		Code: "RU", Name: "Russia", Children: []*service.GeoTreeNode{{Code: "MSK", Name: "Moscow"}},
	}})
	if len(tree) != 1 || len(tree[0].Children) != 1 || tree[0].Children[0].Code != "MSK" {
		t.Fatalf("unexpected geo tree: %+v", tree)
	}
}
