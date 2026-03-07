# \SearchAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**SearchGuildIdMessagesPost**](SearchAPI.md#SearchGuildIdMessagesPost) | **Post** /search/{guild_id}/messages | Search messages
[**SearchMessagesPost**](SearchAPI.md#SearchMessagesPost) | **Post** /search/messages | Search messages in a channel



## SearchGuildIdMessagesPost

> []SearchMessageSearchResponse SearchGuildIdMessagesPost(ctx, guildId).Request(request).Execute()

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
	guildId := int64(789) // int64 | Guild id
	request := *openapiclient.NewSearchMessageSearchRequest() // SearchMessageSearchRequest | Search request data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SearchAPI.SearchGuildIdMessagesPost(context.Background(), guildId).Request(request).Execute()
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
**guildId** | **int64** | Guild id | 

### Other Parameters

Other parameters are passed through a pointer to a apiSearchGuildIdMessagesPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **request** | [**SearchMessageSearchRequest**](SearchMessageSearchRequest.md) | Search request data | 

### Return type

[**[]SearchMessageSearchResponse**](SearchMessageSearchResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SearchMessagesPost

> []SearchMessageSearchResponse SearchMessagesPost(ctx).Request(request).Execute()

Search messages in a channel

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
	request := *openapiclient.NewSearchMessageSearchRequest() // SearchMessageSearchRequest | Search request data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SearchAPI.SearchMessagesPost(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SearchAPI.SearchMessagesPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SearchMessagesPost`: []SearchMessageSearchResponse
	fmt.Fprintf(os.Stdout, "Response from `SearchAPI.SearchMessagesPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSearchMessagesPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**SearchMessageSearchRequest**](SearchMessageSearchRequest.md) | Search request data | 

### Return type

[**[]SearchMessageSearchResponse**](SearchMessageSearchResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

