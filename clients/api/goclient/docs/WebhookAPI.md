# \WebhookAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**WebhookAttachmentsFinalizePost**](WebhookAPI.md#WebhookAttachmentsFinalizePost) | **Post** /webhook/attachments/finalize | Finalize attachment metadata
[**WebhookSfuChannelAlivePost**](WebhookAPI.md#WebhookSfuChannelAlivePost) | **Post** /webhook/sfu/channel/alive | SFU update channel TTL
[**WebhookSfuHeartbeatPost**](WebhookAPI.md#WebhookSfuHeartbeatPost) | **Post** /webhook/sfu/heartbeat | SFU heartbeat
[**WebhookSfuVoiceJoinPost**](WebhookAPI.md#WebhookSfuVoiceJoinPost) | **Post** /webhook/sfu/voice/join | SFU voice join
[**WebhookSfuVoiceLeavePost**](WebhookAPI.md#WebhookSfuVoiceLeavePost) | **Post** /webhook/sfu/voice/leave | SFU voice leave



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


## WebhookSfuChannelAlivePost

> map[string]interface{} WebhookSfuChannelAlivePost(ctx).XWebhookToken(xWebhookToken).SfuChannelAlive(sfuChannelAlive).Execute()

SFU update channel TTL



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
	sfuChannelAlive := *openapiclient.NewSfuChannelAlive() // SfuChannelAlive | Channel liveness data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WebhookAPI.WebhookSfuChannelAlivePost(context.Background()).XWebhookToken(xWebhookToken).SfuChannelAlive(sfuChannelAlive).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WebhookAPI.WebhookSfuChannelAlivePost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `WebhookSfuChannelAlivePost`: map[string]interface{}
	fmt.Fprintf(os.Stdout, "Response from `WebhookAPI.WebhookSfuChannelAlivePost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiWebhookSfuChannelAlivePostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xWebhookToken** | **string** | JWT token | 
 **sfuChannelAlive** | [**SfuChannelAlive**](SfuChannelAlive.md) | Channel liveness data | 

### Return type

**map[string]interface{}**

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


## WebhookSfuVoiceJoinPost

> map[string]interface{} WebhookSfuVoiceJoinPost(ctx).XWebhookToken(xWebhookToken).SfuChannelUserJoin(sfuChannelUserJoin).Execute()

SFU voice join



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
	sfuChannelUserJoin := *openapiclient.NewSfuChannelUserJoin() // SfuChannelUserJoin | Client join data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WebhookAPI.WebhookSfuVoiceJoinPost(context.Background()).XWebhookToken(xWebhookToken).SfuChannelUserJoin(sfuChannelUserJoin).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WebhookAPI.WebhookSfuVoiceJoinPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `WebhookSfuVoiceJoinPost`: map[string]interface{}
	fmt.Fprintf(os.Stdout, "Response from `WebhookAPI.WebhookSfuVoiceJoinPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiWebhookSfuVoiceJoinPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xWebhookToken** | **string** | JWT token | 
 **sfuChannelUserJoin** | [**SfuChannelUserJoin**](SfuChannelUserJoin.md) | Client join data | 

### Return type

**map[string]interface{}**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## WebhookSfuVoiceLeavePost

> map[string]interface{} WebhookSfuVoiceLeavePost(ctx).XWebhookToken(xWebhookToken).SfuChannelUserLeave(sfuChannelUserLeave).Execute()

SFU voice leave



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
	sfuChannelUserLeave := *openapiclient.NewSfuChannelUserLeave() // SfuChannelUserLeave | Client join data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WebhookAPI.WebhookSfuVoiceLeavePost(context.Background()).XWebhookToken(xWebhookToken).SfuChannelUserLeave(sfuChannelUserLeave).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WebhookAPI.WebhookSfuVoiceLeavePost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `WebhookSfuVoiceLeavePost`: map[string]interface{}
	fmt.Fprintf(os.Stdout, "Response from `WebhookAPI.WebhookSfuVoiceLeavePost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiWebhookSfuVoiceLeavePostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xWebhookToken** | **string** | JWT token | 
 **sfuChannelUserLeave** | [**SfuChannelUserLeave**](SfuChannelUserLeave.md) | Client join data | 

### Return type

**map[string]interface{}**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

