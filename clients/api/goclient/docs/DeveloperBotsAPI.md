# \DeveloperBotsAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DeveloperBotsAuthorizePreviewGet**](DeveloperBotsAPI.md#DeveloperBotsAuthorizePreviewGet) | **Get** /developer/bots/authorize/preview | Preview bot authorization
[**DeveloperBotsBotIdAvatarPost**](DeveloperBotsAPI.md#DeveloperBotsBotIdAvatarPost) | **Post** /developer/bots/{bot_id}/avatar | Create bot avatar upload metadata
[**DeveloperBotsBotIdBannerPost**](DeveloperBotsAPI.md#DeveloperBotsBotIdBannerPost) | **Post** /developer/bots/{bot_id}/banner | Create bot banner upload metadata
[**DeveloperBotsBotIdDelete**](DeveloperBotsAPI.md#DeveloperBotsBotIdDelete) | **Delete** /developer/bots/{bot_id} | Delete an owned bot
[**DeveloperBotsBotIdGet**](DeveloperBotsAPI.md#DeveloperBotsBotIdGet) | **Get** /developer/bots/{bot_id} | Get an owned bot
[**DeveloperBotsBotIdGrantsGet**](DeveloperBotsAPI.md#DeveloperBotsBotIdGrantsGet) | **Get** /developer/bots/{bot_id}/grants | List bot install grants
[**DeveloperBotsBotIdGrantsGrantIdDelete**](DeveloperBotsAPI.md#DeveloperBotsBotIdGrantsGrantIdDelete) | **Delete** /developer/bots/{bot_id}/grants/{grant_id} | Revoke a bot install grant
[**DeveloperBotsBotIdGrantsPost**](DeveloperBotsAPI.md#DeveloperBotsBotIdGrantsPost) | **Post** /developer/bots/{bot_id}/grants | Create a bot install grant
[**DeveloperBotsBotIdPatch**](DeveloperBotsAPI.md#DeveloperBotsBotIdPatch) | **Patch** /developer/bots/{bot_id} | Update an owned bot
[**DeveloperBotsBotIdTokensGet**](DeveloperBotsAPI.md#DeveloperBotsBotIdTokensGet) | **Get** /developer/bots/{bot_id}/tokens | List bot runtime tokens
[**DeveloperBotsBotIdTokensPost**](DeveloperBotsAPI.md#DeveloperBotsBotIdTokensPost) | **Post** /developer/bots/{bot_id}/tokens | Create a bot runtime token
[**DeveloperBotsBotIdTokensTokenIdDelete**](DeveloperBotsAPI.md#DeveloperBotsBotIdTokensTokenIdDelete) | **Delete** /developer/bots/{bot_id}/tokens/{token_id} | Revoke a bot runtime token
[**DeveloperBotsGet**](DeveloperBotsAPI.md#DeveloperBotsGet) | **Get** /developer/bots | List owned bots
[**DeveloperBotsPost**](DeveloperBotsAPI.md#DeveloperBotsPost) | **Post** /developer/bots | Create a bot
[**DeveloperBotsPublicGet**](DeveloperBotsAPI.md#DeveloperBotsPublicGet) | **Get** /developer/bots/public | Search public bots



## DeveloperBotsAuthorizePreviewGet

> DeveloperBotAuthorizationPreview DeveloperBotsAuthorizePreviewGet(ctx).GrantToken(grantToken).BotUserId(botUserId).Permissions(permissions).Execute()

Preview bot authorization

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
	grantToken := "grantToken_example" // string | Install grant token (optional)
	botUserId := int32(56) // int32 | Public bot user id (optional)
	permissions := int32(56) // int32 | Requested permission bitmask (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeveloperBotsAPI.DeveloperBotsAuthorizePreviewGet(context.Background()).GrantToken(grantToken).BotUserId(botUserId).Permissions(permissions).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeveloperBotsAPI.DeveloperBotsAuthorizePreviewGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeveloperBotsAuthorizePreviewGet`: DeveloperBotAuthorizationPreview
	fmt.Fprintf(os.Stdout, "Response from `DeveloperBotsAPI.DeveloperBotsAuthorizePreviewGet`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeveloperBotsAuthorizePreviewGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **grantToken** | **string** | Install grant token | 
 **botUserId** | **int32** | Public bot user id | 
 **permissions** | **int32** | Requested permission bitmask | 

### Return type

[**DeveloperBotAuthorizationPreview**](DeveloperBotAuthorizationPreview.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeveloperBotsBotIdAvatarPost

> DtoAvatarUpload DeveloperBotsBotIdAvatarPost(ctx, botId).Request(request).Execute()

Create bot avatar upload metadata

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
	botId := int32(56) // int32 | Bot user id
	request := *openapiclient.NewDeveloperCreateBotAvatarRequest() // DeveloperCreateBotAvatarRequest | Avatar upload request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeveloperBotsAPI.DeveloperBotsBotIdAvatarPost(context.Background(), botId).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeveloperBotsAPI.DeveloperBotsBotIdAvatarPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeveloperBotsBotIdAvatarPost`: DtoAvatarUpload
	fmt.Fprintf(os.Stdout, "Response from `DeveloperBotsAPI.DeveloperBotsBotIdAvatarPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**botId** | **int32** | Bot user id | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeveloperBotsBotIdAvatarPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **request** | [**DeveloperCreateBotAvatarRequest**](DeveloperCreateBotAvatarRequest.md) | Avatar upload request | 

### Return type

[**DtoAvatarUpload**](DtoAvatarUpload.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeveloperBotsBotIdBannerPost

> DtoBannerUpload DeveloperBotsBotIdBannerPost(ctx, botId).Request(request).Execute()

Create bot banner upload metadata

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
	botId := int32(56) // int32 | Bot user id
	request := *openapiclient.NewDeveloperCreateBotBannerRequest() // DeveloperCreateBotBannerRequest | Banner upload request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeveloperBotsAPI.DeveloperBotsBotIdBannerPost(context.Background(), botId).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeveloperBotsAPI.DeveloperBotsBotIdBannerPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeveloperBotsBotIdBannerPost`: DtoBannerUpload
	fmt.Fprintf(os.Stdout, "Response from `DeveloperBotsAPI.DeveloperBotsBotIdBannerPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**botId** | **int32** | Bot user id | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeveloperBotsBotIdBannerPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **request** | [**DeveloperCreateBotBannerRequest**](DeveloperCreateBotBannerRequest.md) | Banner upload request | 

### Return type

[**DtoBannerUpload**](DtoBannerUpload.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeveloperBotsBotIdDelete

> DeveloperBotsBotIdDelete(ctx, botId).Execute()

Delete an owned bot

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
	botId := int32(56) // int32 | Bot user id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeveloperBotsAPI.DeveloperBotsBotIdDelete(context.Background(), botId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeveloperBotsAPI.DeveloperBotsBotIdDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**botId** | **int32** | Bot user id | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeveloperBotsBotIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeveloperBotsBotIdGet

> DeveloperBotResponse DeveloperBotsBotIdGet(ctx, botId).Execute()

Get an owned bot

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
	botId := int32(56) // int32 | Bot user id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeveloperBotsAPI.DeveloperBotsBotIdGet(context.Background(), botId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeveloperBotsAPI.DeveloperBotsBotIdGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeveloperBotsBotIdGet`: DeveloperBotResponse
	fmt.Fprintf(os.Stdout, "Response from `DeveloperBotsAPI.DeveloperBotsBotIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**botId** | **int32** | Bot user id | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeveloperBotsBotIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**DeveloperBotResponse**](DeveloperBotResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeveloperBotsBotIdGrantsGet

> []ModelBotInstallGrant DeveloperBotsBotIdGrantsGet(ctx, botId).Execute()

List bot install grants

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
	botId := int32(56) // int32 | Bot user id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeveloperBotsAPI.DeveloperBotsBotIdGrantsGet(context.Background(), botId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeveloperBotsAPI.DeveloperBotsBotIdGrantsGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeveloperBotsBotIdGrantsGet`: []ModelBotInstallGrant
	fmt.Fprintf(os.Stdout, "Response from `DeveloperBotsAPI.DeveloperBotsBotIdGrantsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**botId** | **int32** | Bot user id | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeveloperBotsBotIdGrantsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]ModelBotInstallGrant**](ModelBotInstallGrant.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeveloperBotsBotIdGrantsGrantIdDelete

> DeveloperBotsBotIdGrantsGrantIdDelete(ctx, botId, grantId).Execute()

Revoke a bot install grant

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
	botId := int32(56) // int32 | Bot user id
	grantId := int32(56) // int32 | Grant id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeveloperBotsAPI.DeveloperBotsBotIdGrantsGrantIdDelete(context.Background(), botId, grantId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeveloperBotsAPI.DeveloperBotsBotIdGrantsGrantIdDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**botId** | **int32** | Bot user id | 
**grantId** | **int32** | Grant id | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeveloperBotsBotIdGrantsGrantIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeveloperBotsBotIdGrantsPost

> DeveloperGrantCreateResponse DeveloperBotsBotIdGrantsPost(ctx, botId).Request(request).Execute()

Create a bot install grant

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
	botId := int32(56) // int32 | Bot user id
	request := *openapiclient.NewDeveloperCreateGrantRequest() // DeveloperCreateGrantRequest | Grant request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeveloperBotsAPI.DeveloperBotsBotIdGrantsPost(context.Background(), botId).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeveloperBotsAPI.DeveloperBotsBotIdGrantsPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeveloperBotsBotIdGrantsPost`: DeveloperGrantCreateResponse
	fmt.Fprintf(os.Stdout, "Response from `DeveloperBotsAPI.DeveloperBotsBotIdGrantsPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**botId** | **int32** | Bot user id | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeveloperBotsBotIdGrantsPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **request** | [**DeveloperCreateGrantRequest**](DeveloperCreateGrantRequest.md) | Grant request | 

### Return type

[**DeveloperGrantCreateResponse**](DeveloperGrantCreateResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeveloperBotsBotIdPatch

> DeveloperBotResponse DeveloperBotsBotIdPatch(ctx, botId).Request(request).Execute()

Update an owned bot

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
	botId := int32(56) // int32 | Bot user id
	request := *openapiclient.NewDeveloperUpdateBotRequest() // DeveloperUpdateBotRequest | Bot update

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeveloperBotsAPI.DeveloperBotsBotIdPatch(context.Background(), botId).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeveloperBotsAPI.DeveloperBotsBotIdPatch``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeveloperBotsBotIdPatch`: DeveloperBotResponse
	fmt.Fprintf(os.Stdout, "Response from `DeveloperBotsAPI.DeveloperBotsBotIdPatch`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**botId** | **int32** | Bot user id | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeveloperBotsBotIdPatchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **request** | [**DeveloperUpdateBotRequest**](DeveloperUpdateBotRequest.md) | Bot update | 

### Return type

[**DeveloperBotResponse**](DeveloperBotResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeveloperBotsBotIdTokensGet

> []ModelBotToken DeveloperBotsBotIdTokensGet(ctx, botId).Execute()

List bot runtime tokens

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
	botId := int32(56) // int32 | Bot user id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeveloperBotsAPI.DeveloperBotsBotIdTokensGet(context.Background(), botId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeveloperBotsAPI.DeveloperBotsBotIdTokensGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeveloperBotsBotIdTokensGet`: []ModelBotToken
	fmt.Fprintf(os.Stdout, "Response from `DeveloperBotsAPI.DeveloperBotsBotIdTokensGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**botId** | **int32** | Bot user id | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeveloperBotsBotIdTokensGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]ModelBotToken**](ModelBotToken.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeveloperBotsBotIdTokensPost

> DeveloperTokenCreateResponse DeveloperBotsBotIdTokensPost(ctx, botId).Request(request).Execute()

Create a bot runtime token

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
	botId := int32(56) // int32 | Bot user id
	request := *openapiclient.NewDeveloperCreateTokenRequest() // DeveloperCreateTokenRequest | Token request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeveloperBotsAPI.DeveloperBotsBotIdTokensPost(context.Background(), botId).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeveloperBotsAPI.DeveloperBotsBotIdTokensPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeveloperBotsBotIdTokensPost`: DeveloperTokenCreateResponse
	fmt.Fprintf(os.Stdout, "Response from `DeveloperBotsAPI.DeveloperBotsBotIdTokensPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**botId** | **int32** | Bot user id | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeveloperBotsBotIdTokensPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **request** | [**DeveloperCreateTokenRequest**](DeveloperCreateTokenRequest.md) | Token request | 

### Return type

[**DeveloperTokenCreateResponse**](DeveloperTokenCreateResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeveloperBotsBotIdTokensTokenIdDelete

> DeveloperBotsBotIdTokensTokenIdDelete(ctx, botId, tokenId).Execute()

Revoke a bot runtime token

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
	botId := int32(56) // int32 | Bot user id
	tokenId := int32(56) // int32 | Token id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.DeveloperBotsAPI.DeveloperBotsBotIdTokensTokenIdDelete(context.Background(), botId, tokenId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeveloperBotsAPI.DeveloperBotsBotIdTokensTokenIdDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**botId** | **int32** | Bot user id | 
**tokenId** | **int32** | Token id | 

### Other Parameters

Other parameters are passed through a pointer to a apiDeveloperBotsBotIdTokensTokenIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeveloperBotsGet

> []DeveloperBotResponse DeveloperBotsGet(ctx).Execute()

List owned bots

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
	resp, r, err := apiClient.DeveloperBotsAPI.DeveloperBotsGet(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeveloperBotsAPI.DeveloperBotsGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeveloperBotsGet`: []DeveloperBotResponse
	fmt.Fprintf(os.Stdout, "Response from `DeveloperBotsAPI.DeveloperBotsGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiDeveloperBotsGetRequest struct via the builder pattern


### Return type

[**[]DeveloperBotResponse**](DeveloperBotResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeveloperBotsPost

> DeveloperBotResponse DeveloperBotsPost(ctx).Request(request).Execute()

Create a bot

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
	request := *openapiclient.NewDeveloperCreateBotRequest() // DeveloperCreateBotRequest | Bot configuration

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeveloperBotsAPI.DeveloperBotsPost(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeveloperBotsAPI.DeveloperBotsPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeveloperBotsPost`: DeveloperBotResponse
	fmt.Fprintf(os.Stdout, "Response from `DeveloperBotsAPI.DeveloperBotsPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeveloperBotsPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**DeveloperCreateBotRequest**](DeveloperCreateBotRequest.md) | Bot configuration | 

### Return type

[**DeveloperBotResponse**](DeveloperBotResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## DeveloperBotsPublicGet

> []DeveloperBotResponse DeveloperBotsPublicGet(ctx).Query(query).Limit(limit).Offset(offset).Execute()

Search public bots

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
	query := "query_example" // string | Search query (optional)
	limit := int32(56) // int32 | Page size (optional)
	offset := int32(56) // int32 | Offset (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.DeveloperBotsAPI.DeveloperBotsPublicGet(context.Background()).Query(query).Limit(limit).Offset(offset).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `DeveloperBotsAPI.DeveloperBotsPublicGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `DeveloperBotsPublicGet`: []DeveloperBotResponse
	fmt.Fprintf(os.Stdout, "Response from `DeveloperBotsAPI.DeveloperBotsPublicGet`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiDeveloperBotsPublicGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **query** | **string** | Search query | 
 **limit** | **int32** | Page size | 
 **offset** | **int32** | Offset | 

### Return type

[**[]DeveloperBotResponse**](DeveloperBotResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

