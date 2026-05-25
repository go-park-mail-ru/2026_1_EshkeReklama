package dto

import "eshkere/internal/service"

type DictionaryItemResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type BlockTypeDictionaryItemResponse struct {
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Platforms   []string `json:"platforms"`
}

type GeoTreeNodeResponse struct {
	Code     string                 `json:"code"`
	Name     string                 `json:"name"`
	Children []*GeoTreeNodeResponse `json:"children"`
}

func ToDictionaryItemResponses(items []service.DictionaryItem) []DictionaryItemResponse {
	out := make([]DictionaryItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, DictionaryItemResponse{
			Code: item.Code,
			Name: item.Name,
		})
	}
	return out
}

func ToBlockTypeDictionaryItemResponses(items []service.BlockTypeDictionaryItem) []BlockTypeDictionaryItemResponse {
	out := make([]BlockTypeDictionaryItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, BlockTypeDictionaryItemResponse{
			Code:        item.Code,
			Name:        item.Name,
			Description: item.Description,
			Platforms:   item.Platforms,
		})
	}
	return out
}

func ToGeoTreeNodeResponses(nodes []*service.GeoTreeNode) []*GeoTreeNodeResponse {
	out := make([]*GeoTreeNodeResponse, 0, len(nodes))
	for _, node := range nodes {
		out = append(out, &GeoTreeNodeResponse{
			Code:     node.Code,
			Name:     node.Name,
			Children: ToGeoTreeNodeResponses(node.Children),
		})
	}
	return out
}
