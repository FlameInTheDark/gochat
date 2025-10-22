# \WebhookAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**WebhookAttachmentsFinalizePost**](WebhookAPI.md#WebhookAttachmentsFinalizePost) | **Post** /webhook/attachments/finalize | Finalize attachment metadata
[**WebhookSfuHeartbeatPost**](WebhookAPI.md#WebhookSfuHeartbeatPost) | **Post** /webhook/sfu/heartbeat | SFU heartbeat



## WebhookAttachmentsFinalizePost

> WebhookAttachmentsFinalizePost(ctx).XWebhookToken(xWebhookToken).AttachmentsFinalizeRequest(attachmentsFinalizeRequest).Execute()

Finalize attachment metadata



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
	xWebhookToken := "xWebhookToken_example" // string | JWT token
	attachmentsFinalizeRequest := *openapiclient.NewAttachmentsFinalizeRequest() // AttachmentsFinalizeRequest | Finalize payload

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WebhookAPI.WebhookAttachmentsFinalizePost(context.Background()).XWebhookToken(xWebhookToken).AttachmentsFinalizeRequest(attachmentsFinalizeRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WebhookAPI.WebhookAttachmentsFinalizePost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiWebhookAttachmentsFinalizePostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xWebhookToken** | **string** | JWT token | 
 **attachmentsFinalizeRequest** | [**AttachmentsFinalizeRequest**](AttachmentsFinalizeRequest.md) | Finalize payload | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## WebhookSfuHeartbeatPost

> WebhookSfuHeartbeatPost(ctx).XWebhookToken(xWebhookToken).SfuHeartbeatRequest(sfuHeartbeatRequest).Execute()

SFU heartbeat



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
	xWebhookToken := "xWebhookToken_example" // string | JWT token
	sfuHeartbeatRequest := *openapiclient.NewSfuHeartbeatRequest() // SfuHeartbeatRequest | Heartbeat payload

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.WebhookAPI.WebhookSfuHeartbeatPost(context.Background()).XWebhookToken(xWebhookToken).SfuHeartbeatRequest(sfuHeartbeatRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WebhookAPI.WebhookSfuHeartbeatPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiWebhookSfuHeartbeatPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xWebhookToken** | **string** | JWT token | 
 **sfuHeartbeatRequest** | [**SfuHeartbeatRequest**](SfuHeartbeatRequest.md) | Heartbeat payload | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

