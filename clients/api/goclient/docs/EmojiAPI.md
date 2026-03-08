# \EmojiAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**EmojiEmojiIdGet**](EmojiAPI.md#EmojiEmojiIdGet) | **Get** /emoji/{emoji_id} | Redirect to public emoji asset



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

