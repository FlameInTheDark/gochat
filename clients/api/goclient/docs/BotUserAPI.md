# \BotUserAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**BotApiV1UserMeGet**](BotUserAPI.md#BotApiV1UserMeGet) | **Get** /bot/api/v1/user/me | Get current bot account



## BotApiV1UserMeGet

> UserMeResponse BotApiV1UserMeGet(ctx).Execute()

Get current bot account

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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.BotUserAPI.BotApiV1UserMeGet(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `BotUserAPI.BotApiV1UserMeGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BotApiV1UserMeGet`: UserMeResponse
	fmt.Fprintf(os.Stdout, "Response from `BotUserAPI.BotApiV1UserMeGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiBotApiV1UserMeGetRequest struct via the builder pattern


### Return type

[**UserMeResponse**](UserMeResponse.md)

### Authorization

[BotToken](../README.md#BotToken)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

