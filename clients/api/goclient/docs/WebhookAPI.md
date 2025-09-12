# \WebhookAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**WebhookStorageEventsPost**](WebhookAPI.md#WebhookStorageEventsPost) | **Post** /webhook/storage/events | Storage event



## WebhookStorageEventsPost

> string WebhookStorageEventsPost(ctx).WebhookS3Event(webhookS3Event).Execute()

Storage event

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
	webhookS3Event := *openapiclient.NewWebhookS3Event() // WebhookS3Event | S3 event

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.WebhookAPI.WebhookStorageEventsPost(context.Background()).WebhookS3Event(webhookS3Event).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `WebhookAPI.WebhookStorageEventsPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `WebhookStorageEventsPost`: string
	fmt.Fprintf(os.Stdout, "Response from `WebhookAPI.WebhookStorageEventsPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiWebhookStorageEventsPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **webhookS3Event** | [**WebhookS3Event**](WebhookS3Event.md) | S3 event | 

### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

