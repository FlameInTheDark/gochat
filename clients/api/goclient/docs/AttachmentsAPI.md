# \AttachmentsAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AttachmentsChannelIdAttachmentIdPost**](AttachmentsAPI.md#AttachmentsChannelIdAttachmentIdPost) | **Post** /attachments/{channel_id}/{attachment_id} | Upload attachment



## AttachmentsChannelIdAttachmentIdPost

> string AttachmentsChannelIdAttachmentIdPost(ctx, channelId, attachmentId).RequestBody(requestBody).Execute()

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
	resp, r, err := apiClient.AttachmentsAPI.AttachmentsChannelIdAttachmentIdPost(context.Background(), channelId, attachmentId).RequestBody(requestBody).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AttachmentsAPI.AttachmentsChannelIdAttachmentIdPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AttachmentsChannelIdAttachmentIdPost`: string
	fmt.Fprintf(os.Stdout, "Response from `AttachmentsAPI.AttachmentsChannelIdAttachmentIdPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**channelId** | **int32** | Channel ID | 
**attachmentId** | **int32** | Attachment ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiAttachmentsChannelIdAttachmentIdPostRequest struct via the builder pattern


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

