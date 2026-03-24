# \EmojiAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**EmojiEmojiIdGet**](EmojiAPI.md#EmojiEmojiIdGet) | **Get** /emoji/{emoji_id} | Redirect to public emoji asset
[**InfoEmojiEmojiIdGet**](EmojiAPI.md#InfoEmojiEmojiIdGet) | **Get** /info/emoji/{emoji_id} | Get emoji info



## EmojiEmojiIdGet

> EmojiEmojiIdGet(ctx, emojiId).Size(size).Execute()

Redirect to public emoji asset

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
	emojiId := "emojiId_example" // string | Emoji filename ending in .webp
	size := int32(56) // int32 | Preferred rendered size (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.EmojiAPI.EmojiEmojiIdGet(context.Background(), emojiId).Size(size).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EmojiAPI.EmojiEmojiIdGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**emojiId** | **string** | Emoji filename ending in .webp | 

### Other Parameters

Other parameters are passed through a pointer to a apiEmojiEmojiIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **size** | **int32** | Preferred rendered size | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: text/plain

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## InfoEmojiEmojiIdGet

> DtoEmojiInfo InfoEmojiEmojiIdGet(ctx, emojiId).Execute()

Get emoji info

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
	emojiId := int32(56) // int32 | Emoji ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.EmojiAPI.InfoEmojiEmojiIdGet(context.Background(), emojiId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `EmojiAPI.InfoEmojiEmojiIdGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `InfoEmojiEmojiIdGet`: DtoEmojiInfo
	fmt.Fprintf(os.Stdout, "Response from `EmojiAPI.InfoEmojiEmojiIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**emojiId** | **int32** | Emoji ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiInfoEmojiEmojiIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**DtoEmojiInfo**](DtoEmojiInfo.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

