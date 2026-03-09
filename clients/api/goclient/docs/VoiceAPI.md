# \VoiceAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**VoiceRegionsGet**](VoiceAPI.md#VoiceRegionsGet) | **Get** /voice/regions | List available voice regions



## VoiceRegionsGet

> VoiceVoiceRegionsResponse VoiceRegionsGet(ctx).Execute()

List available voice regions

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
	resp, r, err := apiClient.VoiceAPI.VoiceRegionsGet(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `VoiceAPI.VoiceRegionsGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `VoiceRegionsGet`: VoiceVoiceRegionsResponse
	fmt.Fprintf(os.Stdout, "Response from `VoiceAPI.VoiceRegionsGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiVoiceRegionsGetRequest struct via the builder pattern


### Return type

[**VoiceVoiceRegionsResponse**](VoiceVoiceRegionsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

