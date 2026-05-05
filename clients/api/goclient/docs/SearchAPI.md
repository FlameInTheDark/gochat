# \SearchAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**SearchGuildIdMessagesPost**](SearchAPI.md#SearchGuildIdMessagesPost) | **Post** /search/{guild_id}/messages | Search messages
[**SearchGuildTagsGet**](SearchAPI.md#SearchGuildTagsGet) | **Get** /search/guild-tags | Autocomplete public guild tags
[**SearchGuildsGet**](SearchAPI.md#SearchGuildsGet) | **Get** /search/guilds | Search public guilds
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
	guildId := int32(56) // int32 | Guild id
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
**guildId** | **int32** | Guild id | 

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


## SearchGuildTagsGet

> []string SearchGuildTagsGet(ctx).Q(q).Limit(limit).Execute()

Autocomplete public guild tags



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
	q := "q_example" // string | Tag prefix (optional)
	limit := int32(56) // int32 | Maximum tags to return, capped at 16 (optional) (default to 16)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SearchAPI.SearchGuildTagsGet(context.Background()).Q(q).Limit(limit).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SearchAPI.SearchGuildTagsGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SearchGuildTagsGet`: []string
	fmt.Fprintf(os.Stdout, "Response from `SearchAPI.SearchGuildTagsGet`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSearchGuildTagsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **q** | **string** | Tag prefix | 
 **limit** | **int32** | Maximum tags to return, capped at 16 | [default to 16]

### Return type

**[]string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## SearchGuildsGet

> DtoGuildDiscoverySearchResponse SearchGuildsGet(ctx).Q(q).Tags(tags).Sort(sort).Page(page).Limit(limit).Execute()

Search public guilds



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
	q := "q_example" // string | Search text for guild name, description, and tags (optional)
	tags := "tags_example" // string | Comma-separated normalized tags (optional)
	sort := "sort_example" // string | Sort mode (optional) (default to "best_match")
	page := int32(56) // int32 | Zero-based page number (optional) (default to 0)
	limit := int32(56) // int32 | Results per page, capped at 16 (optional) (default to 16)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.SearchAPI.SearchGuildsGet(context.Background()).Q(q).Tags(tags).Sort(sort).Page(page).Limit(limit).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `SearchAPI.SearchGuildsGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `SearchGuildsGet`: DtoGuildDiscoverySearchResponse
	fmt.Fprintf(os.Stdout, "Response from `SearchAPI.SearchGuildsGet`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiSearchGuildsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **q** | **string** | Search text for guild name, description, and tags | 
 **tags** | **string** | Comma-separated normalized tags | 
 **sort** | **string** | Sort mode | [default to &quot;best_match&quot;]
 **page** | **int32** | Zero-based page number | [default to 0]
 **limit** | **int32** | Results per page, capped at 16 | [default to 16]

### Return type

[**DtoGuildDiscoverySearchResponse**](DtoGuildDiscoverySearchResponse.md)

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

