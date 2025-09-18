# \SearchAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**SearchGuildIdMessagesPost**](SearchAPI.md#SearchGuildIdMessagesPost) | **Post** /search/{guild_id}/messages | Search messages



## SearchGuildIdMessagesPost

> []SearchMessageSearchResponse SearchGuildIdMessagesPost(ctx, guildId).SearchMessageSearchRequest(searchMessageSearchRequest).Execute()

Search messages

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/FlameInTheDark/gochat/clients/api/goclient"
)

func main() {
	guildId := int32(56) // int32 | Channel id
	searchMessageSearchRequest := *openapiclient.NewSearchMessageSearchRequest() // SearchMessageSearchRequest | Search request data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SearchAPI.SearchGuildIdMessagesPost(context.Background(), guildId).SearchMessageSearchRequest(searchMessageSearchRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SearchAPI.SearchGuildIdMessagesPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SearchGuildIdMessagesPost`: []SearchMessageSearchResponse
	fmt.Fprintf(os.Stdout, "Response from `SearchAPI.SearchGuildIdMessagesPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Channel id | 

### Other Parameters

Other parameters are passed through a pointer to a apiSearchGuildIdMessagesPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **searchMessageSearchRequest** | [**SearchMessageSearchRequest**](SearchMessageSearchRequest.md) | Search request data | 

### Return type

[**[]SearchMessageSearchResponse**](SearchMessageSearchResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

