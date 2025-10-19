# \AvatarsAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AvatarsUserIdAvatarIdPost**](AvatarsAPI.md#AvatarsUserIdAvatarIdPost) | **Post** /avatars/{user_id}/{avatar_id} | Upload user avatar



## AvatarsUserIdAvatarIdPost

> string AvatarsUserIdAvatarIdPost(ctx, userId, avatarId).RequestBody(requestBody).Execute()

Upload user avatar



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
	userId := int32(56) // int32 | User ID
	avatarId := int32(56) // int32 | Avatar ID
	requestBody := []int32{int32(123)} // []int32 | Binary image payload

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AvatarsAPI.AvatarsUserIdAvatarIdPost(context.Background(), userId, avatarId).RequestBody(requestBody).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AvatarsAPI.AvatarsUserIdAvatarIdPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AvatarsUserIdAvatarIdPost`: string
	fmt.Fprintf(os.Stdout, "Response from `AvatarsAPI.AvatarsUserIdAvatarIdPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**userId** | **int32** | User ID | 
**avatarId** | **int32** | Avatar ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiAvatarsUserIdAvatarIdPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **requestBody** | **[]int32** | Binary image payload | 

### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/octet-stream
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

