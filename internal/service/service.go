package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

type SearchService interface {
	SearchStoresByPincode(pincode string, term string) (any, error)
	SearchProductsByPincode(pincode string, term string) (any, error)
}

type searchService struct {
	client *opensearchapi.Client
}

func NewSearchService(client *opensearchapi.Client) SearchService {
	return &searchService{client: client}
}

func (s *searchService) SearchStoresByPincode(pincode string, term string) (any, error) {
	// Implement the logic to search for stores by pincode
	content := strings.NewReader(`
	{
    "query": {
       "bool": {
          "filter": [
           {
             "term": {
               "pincode": ` + pincode + `
            }
        }
      ],
    "must": [
      {
      "multi_match": {
      "query": "` + term + `",
      "fields": ["name^3", "category^1", "sub_category^2", "description","address^5", "primary_cuisine", "secondary_cuisine"],
      "fuzziness": "AUTO"
    }
      }
    ]
    }
  }
}
	`)

	ctx := context.Background()
	searchResp, err := s.client.Search(
		ctx, &opensearchapi.SearchReq{
			Body:    content,
			Indices: []string{"stores-index"},
		},
	)

	if err != nil {
		fmt.Printf("Error searching: %v\n", err)
		return err, nil
	}
	fmt.Printf("Search hits: %v\n", searchResp.Hits.Total.Value)

	stores := make([]map[string]interface{}, 0)

	if searchResp.Hits.Total.Value > 0 {
		for _, hit := range searchResp.Hits.Hits {

			store := make(map[string]interface{})
			err := json.Unmarshal(hit.Source, &store)
			if err != nil {
				fmt.Printf("Error unmarshalling: %v\n", err)
				return err, nil
			}
			stores = append(stores, store)
		}

	}
	fmt.Printf("Stores: %+v\n", stores)
	return stores, nil
}

func (s *searchService) getStoreIdsByPincode(pincode string) (map[string]string, error) {
	// Implement the logic to get store ids by pincode
	content := strings.NewReader(`
	{
				"query": {
					"term": {
						"pincode": ` + pincode + `
					}
				}
	}`)
	ctx := context.Background()
	searchResp, err := s.client.Search(
		ctx, &opensearchapi.SearchReq{
			Body:    content,
			Indices: []string{"stores-index"},
		},
	)

	if err != nil {
		fmt.Printf("Error searching: %v\n", err)
		return map[string]string{}, err
	}
	fmt.Printf("Search hits: %v\n", searchResp.Hits.Total.Value)

	// storeIds := make([]string, 0)
	// storeNames := make([]string, 0)
	storeIdToName := make(map[string]string)
	if searchResp.Hits.Total.Value > 0 {
		for _, hit := range searchResp.Hits.Hits {
			store := make(map[string]interface{})
			err := json.Unmarshal(hit.Source, &store)
			if err != nil {
				fmt.Printf("Error unmarshalling: %v\n", err)
				return map[string]string{}, err
			}
			// storeIds = append(storeIds, store["id"].(string))
			// storeNames = append(storeNames, store["name"].(string))
			storeIdToName[store["id"].(string)] = store["name"].(string)
		}

	}

	return storeIdToName, nil
}

func (s *searchService) SearchProductsByPincode(pincode string, term string) (any, error) {

	storeIdToName, err := s.getStoreIdsByPincode(pincode)
	if err != nil {
		return err, nil
	}

	storeIds := make([]string, 0)
	for storeId := range storeIdToName {
		storeIds = append(storeIds, storeId)
	}
	// storeNames := make([]string, 0)

	content := strings.NewReader(`{
  "query": {
    "bool": {
       "filter": {
   "terms": {
    "store_id.keyword": [
	  "` + strings.Join(storeIds, `","`) + `"
        ]
      
    }
        
      },
      "must": [
        {
        "multi_match": {
        "query": "` + term + `",  
          "fields": ["name", "description", "category", "brand"],
        "fuzziness": "AUTO"
        }
        }
       ]
   
    }
  }
}`)
	ctx := context.Background()
	searchResp, err := s.client.Search(
		ctx, &opensearchapi.SearchReq{
			Body:    content,
			Indices: []string{"products-index"},
		},
	)

	if err != nil {
		fmt.Printf("Error searching: %v\n", err)
		return err, nil
	}

	fmt.Printf("Search hits: %v\n", searchResp.Hits.Total.Value)

	products := make([]map[string]interface{}, 0)

	if searchResp.Hits.Total.Value > 0 {
		for _, hit := range searchResp.Hits.Hits {

			product := make(map[string]interface{})

			err := json.Unmarshal(hit.Source, &product)
			if err != nil {
				fmt.Printf("Error unmarshalling: %v\n", err)
				return err, nil
			}
			product["store_name"] = storeIdToName[product["store_id"].(string)]

			products = append(products, product)
		}

	}
	fmt.Printf("Products: %+v\n", products)
	return products, nil

}
