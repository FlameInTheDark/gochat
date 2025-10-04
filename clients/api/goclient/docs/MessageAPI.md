# \MessageAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**MessageChannelChannelIdAttachmentPost**](MessageAPI.md#MessageChannelChannelIdAttachmentPost) | **Post** /message/channel/{channel_id}/attachment | Create attachment
[**MessageChannelChannelIdGet**](MessageAPI.md#MessageChannelChannelIdGet) | **Get** /message/channel/{channel_id} | Get messages
[**MessageChannelChannelIdMessageIdDelete**](MessageAPI.md#MessageChannelChannelIdMessageIdDelete) | **Delete** /message/channel/{channel_id}/{message_id} | Delete message
[**MessageChannelChannelIdMessageIdPatch**](MessageAPI.md#MessageChannelChannelIdMessageIdPatch) | **Patch** /message/channel/{channel_id}/{message_id} | Update message
[**MessageChannelChannelIdPost**](MessageAPI.md#MessageChannelChannelIdPost) | **Post** /message/channel/{channel_id} | Send message



## MessageChannelChannelIdAttachmentPost

> DtoAttachmentUpload MessageChannelChannelIdAttachmentPost(ctx, channelId).MessageUploadAttachmentRequest(messageUploadAttachmentRequest).Execute()

Create attachment

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
	channelId := int32(56) // int32 | Channel id
	messageUploadAttachmentRequest := *openapiclient.NewMessageUploadAttachmentRequest() // MessageUploadAttachmentRequest | Attachment data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.MessageAPI.MessageChannelChannelIdAttachmentPost(context.Background(), channelId).MessageUploadAttachmentRequest(messageUploadAttachmentRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MessageAPI.MessageChannelChannelIdAttachmentPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `MessageChannelChannelIdAttachmentPost`: DtoAttachmentUpload
	fmt.Fprintf(os.Stdout, "Response from `MessageAPI.MessageChannelChannelIdAttachmentPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**channelId** | **int32** | Channel id | 

### Other Parameters

Other parameters are passed through a pointer to a apiMessageChannelChannelIdAttachmentPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **messageUploadAttachmentRequest** | [**MessageUploadAttachmentRequest**](MessageUploadAttachmentRequest.md) | Attachment data | 

### Return type

[**DtoAttachmentUpload**](DtoAttachmentUpload.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## MessageChannelChannelIdGet

> []DtoMessage MessageChannelChannelIdGet(ctx, channelId).From(from).Direction(direction).Limit(limit).Execute()

Get messages

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
	channelId := int32(2230469276416868352) // int32 | Channel id
	from := int32(2230469276416868352) // int32 | Start point for messages (optional)
	direction := "before" // string | Select direction (optional)
	limit := int32(30) // int32 | Message count limit (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.MessageAPI.MessageChannelChannelIdGet(context.Background(), channelId).From(from).Direction(direction).Limit(limit).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MessageAPI.MessageChannelChannelIdGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `MessageChannelChannelIdGet`: []DtoMessage
	fmt.Fprintf(os.Stdout, "Response from `MessageAPI.MessageChannelChannelIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**channelId** | **int32** | Channel id | 

### Other Parameters

Other parameters are passed through a pointer to a apiMessageChannelChannelIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **from** | **int32** | Start point for messages | 
 **direction** | **string** | Select direction | 
 **limit** | **int32** | Message count limit | 

### Return type

[**[]DtoMessage**](DtoMessage.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## MessageChannelChannelIdMessageIdDelete

> string MessageChannelChannelIdMessageIdDelete(ctx, messageId, channelId).Execute()

Delete message

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
	messageId := int32(56) // int32 | Message id
	channelId := int32(56) // int32 | Channel id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.MessageAPI.MessageChannelChannelIdMessageIdDelete(context.Background(), messageId, channelId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MessageAPI.MessageChannelChannelIdMessageIdDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `MessageChannelChannelIdMessageIdDelete`: string
	fmt.Fprintf(os.Stdout, "Response from `MessageAPI.MessageChannelChannelIdMessageIdDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**messageId** | **int32** | Message id | 
**channelId** | **int32** | Channel id | 

### Other Parameters

Other parameters are passed through a pointer to a apiMessageChannelChannelIdMessageIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## MessageChannelChannelIdMessageIdPatch

> DtoMessage MessageChannelChannelIdMessageIdPatch(ctx, messageId, channelId).MessageUpdateMessageRequest(messageUpdateMessageRequest).Execute()

Update message

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
	messageId := int32(56) // int32 | Message id
	channelId := int32(56) // int32 | Channel id
	messageUpdateMessageRequest := *openapiclient.NewMessageUpdateMessageRequest() // MessageUpdateMessageRequest | Message data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.MessageAPI.MessageChannelChannelIdMessageIdPatch(context.Background(), messageId, channelId).MessageUpdateMessageRequest(messageUpdateMessageRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MessageAPI.MessageChannelChannelIdMessageIdPatch``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `MessageChannelChannelIdMessageIdPatch`: DtoMessage
	fmt.Fprintf(os.Stdout, "Response from `MessageAPI.MessageChannelChannelIdMessageIdPatch`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**messageId** | **int32** | Message id | 
**channelId** | **int32** | Channel id | 

### Other Parameters

Other parameters are passed through a pointer to a apiMessageChannelChannelIdMessageIdPatchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **messageUpdateMessageRequest** | [**MessageUpdateMessageRequest**](MessageUpdateMessageRequest.md) | Message data | 

### Return type

[**DtoMessage**](DtoMessage.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## MessageChannelChannelIdPost

> DtoMessage MessageChannelChannelIdPost(ctx, channelId).MessageSendMessageRequest(messageSendMessageRequest).Execute()

Send message

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
	channelId := int32(56) // int32 | Channel id
	messageSendMessageRequest := *openapiclient.NewMessageSendMessageRequest() // MessageSendMessageRequest | Message data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.MessageAPI.MessageChannelChannelIdPost(context.Background(), channelId).MessageSendMessageRequest(messageSendMessageRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `MessageAPI.MessageChannelChannelIdPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `MessageChannelChannelIdPost`: DtoMessage
	fmt.Fprintf(os.Stdout, "Response from `MessageAPI.MessageChannelChannelIdPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**channelId** | **int32** | Channel id | 

### Other Parameters

Other parameters are passed through a pointer to a apiMessageChannelChannelIdPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **messageSendMessageRequest** | [**MessageSendMessageRequest**](MessageSendMessageRequest.md) | Message data | 

### Return type

[**DtoMessage**](DtoMessage.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

