# \UploadAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**UploadAttachmentsChannelIdAttachmentIdPost**](UploadAPI.md#UploadAttachmentsChannelIdAttachmentIdPost) | **Post** /upload/attachments/{channel_id}/{attachment_id} | Upload attachment
[**UploadAvatarsUserIdAvatarIdPost**](UploadAPI.md#UploadAvatarsUserIdAvatarIdPost) | **Post** /upload/avatars/{user_id}/{avatar_id} | Upload user avatar
[**UploadIconsGuildIdIconIdPost**](UploadAPI.md#UploadIconsGuildIdIconIdPost) | **Post** /upload/icons/{guild_id}/{icon_id} | Upload guild icon



## UploadAttachmentsChannelIdAttachmentIdPost

> string UploadAttachmentsChannelIdAttachmentIdPost(ctx, channelId, attachmentId).File(file).Execute()

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
	channelId := int64(789) // int64 | Channel ID
	attachmentId := int64(789) // int64 | Attachment ID
	file := []int32{int32(123)} // []int32 | Binary file to upload

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UploadAPI.UploadAttachmentsChannelIdAttachmentIdPost(context.Background(), channelId, attachmentId).File(file).Execute()
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
**channelId** | **int64** | Channel ID | 
**attachmentId** | **int64** | Attachment ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiUploadAttachmentsChannelIdAttachmentIdPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **file** | **[]int32** | Binary file to upload | 

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

> string UploadAvatarsUserIdAvatarIdPost(ctx, userId, avatarId).File(file).Execute()

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
	userId := int64(789) // int64 | User ID
	avatarId := int64(789) // int64 | Avatar ID
	file := []int32{int32(123)} // []int32 | Binary image payload

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UploadAPI.UploadAvatarsUserIdAvatarIdPost(context.Background(), userId, avatarId).File(file).Execute()
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
**userId** | **int64** | User ID | 
**avatarId** | **int64** | Avatar ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiUploadAvatarsUserIdAvatarIdPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **file** | **[]int32** | Binary image payload | 

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

> string UploadIconsGuildIdIconIdPost(ctx, guildId, iconId).File(file).Execute()

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
	guildId := int64(789) // int64 | Guild ID
	iconId := int64(789) // int64 | Icon ID
	file := []int32{int32(123)} // []int32 | Binary image payload

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UploadAPI.UploadIconsGuildIdIconIdPost(context.Background(), guildId, iconId).File(file).Execute()
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
**guildId** | **int64** | Guild ID | 
**iconId** | **int64** | Icon ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiUploadIconsGuildIdIconIdPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **file** | **[]int32** | Binary image payload | 

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

