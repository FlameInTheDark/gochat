# \UploadAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**UploadAttachmentsChannelIdAttachmentIdPost**](UploadAPI.md#UploadAttachmentsChannelIdAttachmentIdPost) | **Post** /upload/attachments/{channel_id}/{attachment_id} | Upload attachment
[**UploadAvatarsUserIdAvatarIdPost**](UploadAPI.md#UploadAvatarsUserIdAvatarIdPost) | **Post** /upload/avatars/{user_id}/{avatar_id} | Upload user avatar
[**UploadIconsGuildIdIconIdPost**](UploadAPI.md#UploadIconsGuildIdIconIdPost) | **Post** /upload/icons/{guild_id}/{icon_id} | Upload guild icon



## UploadAttachmentsChannelIdAttachmentIdPost

> string UploadAttachmentsChannelIdAttachmentIdPost(ctx, channelId, attachmentId).RequestBody(requestBody).Execute()

Upload attachment



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
	channelId := int32(56) // int32 | Channel ID
	attachmentId := int32(56) // int32 | Attachment ID
	requestBody := []int32{int32(123)} // []int32 | Binary file to upload

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UploadAPI.UploadAttachmentsChannelIdAttachmentIdPost(context.Background(), channelId, attachmentId).RequestBody(requestBody).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UploadAPI.UploadAttachmentsChannelIdAttachmentIdPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UploadAttachmentsChannelIdAttachmentIdPost`: string
	fmt.Fprintf(os.Stdout, "Response from `UploadAPI.UploadAttachmentsChannelIdAttachmentIdPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**channelId** | **int32** | Channel ID | 
**attachmentId** | **int32** | Attachment ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiUploadAttachmentsChannelIdAttachmentIdPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **requestBody** | **[]int32** | Binary file to upload | 

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


## UploadAvatarsUserIdAvatarIdPost

> string UploadAvatarsUserIdAvatarIdPost(ctx, userId, avatarId).RequestBody(requestBody).Execute()

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
	resp, r, err := apiClient.UploadAPI.UploadAvatarsUserIdAvatarIdPost(context.Background(), userId, avatarId).RequestBody(requestBody).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UploadAPI.UploadAvatarsUserIdAvatarIdPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UploadAvatarsUserIdAvatarIdPost`: string
	fmt.Fprintf(os.Stdout, "Response from `UploadAPI.UploadAvatarsUserIdAvatarIdPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**userId** | **int32** | User ID | 
**avatarId** | **int32** | Avatar ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiUploadAvatarsUserIdAvatarIdPostRequest struct via the builder pattern


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


## UploadIconsGuildIdIconIdPost

> string UploadIconsGuildIdIconIdPost(ctx, guildId, iconId).RequestBody(requestBody).Execute()

Upload guild icon



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
	guildId := int32(56) // int32 | Guild ID
	iconId := int32(56) // int32 | Icon ID
	requestBody := []int32{int32(123)} // []int32 | Binary image payload

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UploadAPI.UploadIconsGuildIdIconIdPost(context.Background(), guildId, iconId).RequestBody(requestBody).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UploadAPI.UploadIconsGuildIdIconIdPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UploadIconsGuildIdIconIdPost`: string
	fmt.Fprintf(os.Stdout, "Response from `UploadAPI.UploadIconsGuildIdIconIdPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 
**iconId** | **int32** | Icon ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiUploadIconsGuildIdIconIdPostRequest struct via the builder pattern


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

