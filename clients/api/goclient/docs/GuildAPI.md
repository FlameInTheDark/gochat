# \GuildAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GuildGuildIdCategoryCategoryIdDelete**](GuildAPI.md#GuildGuildIdCategoryCategoryIdDelete) | **Delete** /guild/{guild_id}/category/{category_id} | Delete channel category
[**GuildGuildIdCategoryPost**](GuildAPI.md#GuildGuildIdCategoryPost) | **Post** /guild/{guild_id}/category | Create guild channel category
[**GuildGuildIdChannelChannelIdDelete**](GuildAPI.md#GuildGuildIdChannelChannelIdDelete) | **Delete** /guild/{guild_id}/channel/{channel_id} | Delete channel
[**GuildGuildIdChannelChannelIdGet**](GuildAPI.md#GuildGuildIdChannelChannelIdGet) | **Get** /guild/{guild_id}/channel/{channel_id} | Get guild channel
[**GuildGuildIdChannelGet**](GuildAPI.md#GuildGuildIdChannelGet) | **Get** /guild/{guild_id}/channel | Get guild channels
[**GuildGuildIdChannelPost**](GuildAPI.md#GuildGuildIdChannelPost) | **Post** /guild/{guild_id}/channel | Create guild channel
[**GuildGuildIdGet**](GuildAPI.md#GuildGuildIdGet) | **Get** /guild/{guild_id} | Get guild
[**GuildGuildIdPatch**](GuildAPI.md#GuildGuildIdPatch) | **Patch** /guild/{guild_id} | Update guild
[**GuildPost**](GuildAPI.md#GuildPost) | **Post** /guild | Create guild



## GuildGuildIdCategoryCategoryIdDelete

> string GuildGuildIdCategoryCategoryIdDelete(ctx, guildId, categoryId).Execute()

Delete channel category

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
	categoryId := int32(56) // int32 | Category ID (actually a channel with special type)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildAPI.GuildGuildIdCategoryCategoryIdDelete(context.Background(), guildId, categoryId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildAPI.GuildGuildIdCategoryCategoryIdDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdCategoryCategoryIdDelete`: string
	fmt.Fprintf(os.Stdout, "Response from `GuildAPI.GuildGuildIdCategoryCategoryIdDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 
**categoryId** | **int32** | Category ID (actually a channel with special type) | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdCategoryCategoryIdDeleteRequest struct via the builder pattern


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


## GuildGuildIdCategoryPost

> string GuildGuildIdCategoryPost(ctx, guildId).GuildCreateGuildChannelCategoryRequest(guildCreateGuildChannelCategoryRequest).Execute()

Create guild channel category

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
	guildCreateGuildChannelCategoryRequest := *openapiclient.NewGuildCreateGuildChannelCategoryRequest() // GuildCreateGuildChannelCategoryRequest | Create category data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildAPI.GuildGuildIdCategoryPost(context.Background(), guildId).GuildCreateGuildChannelCategoryRequest(guildCreateGuildChannelCategoryRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildAPI.GuildGuildIdCategoryPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdCategoryPost`: string
	fmt.Fprintf(os.Stdout, "Response from `GuildAPI.GuildGuildIdCategoryPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdCategoryPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **guildCreateGuildChannelCategoryRequest** | [**GuildCreateGuildChannelCategoryRequest**](GuildCreateGuildChannelCategoryRequest.md) | Create category data | 

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


## GuildGuildIdChannelChannelIdDelete

> string GuildGuildIdChannelChannelIdDelete(ctx, guildId, channelId).Execute()

Delete channel

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
	channelId := int32(56) // int32 | Channel ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildAPI.GuildGuildIdChannelChannelIdDelete(context.Background(), guildId, channelId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildAPI.GuildGuildIdChannelChannelIdDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdChannelChannelIdDelete`: string
	fmt.Fprintf(os.Stdout, "Response from `GuildAPI.GuildGuildIdChannelChannelIdDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 
**channelId** | **int32** | Channel ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdChannelChannelIdDeleteRequest struct via the builder pattern


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


## GuildGuildIdChannelChannelIdGet

> DtoChannel GuildGuildIdChannelChannelIdGet(ctx, guildId, channelId).Execute()

Get guild channel

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
	guildId := int32(56) // int32 | Guild id
	channelId := int32(56) // int32 | Channel id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildAPI.GuildGuildIdChannelChannelIdGet(context.Background(), guildId, channelId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildAPI.GuildGuildIdChannelChannelIdGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdChannelChannelIdGet`: DtoChannel
	fmt.Fprintf(os.Stdout, "Response from `GuildAPI.GuildGuildIdChannelChannelIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild id | 
**channelId** | **int32** | Channel id | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdChannelChannelIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**DtoChannel**](DtoChannel.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GuildGuildIdChannelGet

> []DtoChannel GuildGuildIdChannelGet(ctx, guildId).Execute()

Get guild channels

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
	guildId := int32(56) // int32 | Guild id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildAPI.GuildGuildIdChannelGet(context.Background(), guildId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildAPI.GuildGuildIdChannelGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdChannelGet`: []DtoChannel
	fmt.Fprintf(os.Stdout, "Response from `GuildAPI.GuildGuildIdChannelGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild id | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdChannelGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]DtoChannel**](DtoChannel.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GuildGuildIdChannelPost

> string GuildGuildIdChannelPost(ctx, guildId).GuildCreateGuildChannelRequest(guildCreateGuildChannelRequest).Execute()

Create guild channel

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
	guildCreateGuildChannelRequest := *openapiclient.NewGuildCreateGuildChannelRequest() // GuildCreateGuildChannelRequest | Create channel data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildAPI.GuildGuildIdChannelPost(context.Background(), guildId).GuildCreateGuildChannelRequest(guildCreateGuildChannelRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildAPI.GuildGuildIdChannelPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdChannelPost`: string
	fmt.Fprintf(os.Stdout, "Response from `GuildAPI.GuildGuildIdChannelPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdChannelPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **guildCreateGuildChannelRequest** | [**GuildCreateGuildChannelRequest**](GuildCreateGuildChannelRequest.md) | Create channel data | 

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


## GuildGuildIdGet

> DtoGuild GuildGuildIdGet(ctx, guildId).Execute()

Get guild

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
	guildId := int32(56) // int32 | Guild id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildAPI.GuildGuildIdGet(context.Background(), guildId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildAPI.GuildGuildIdGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdGet`: DtoGuild
	fmt.Fprintf(os.Stdout, "Response from `GuildAPI.GuildGuildIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild id | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**DtoGuild**](DtoGuild.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GuildGuildIdPatch

> DtoGuild GuildGuildIdPatch(ctx, guildId).GuildUpdateGuildRequest(guildUpdateGuildRequest).Execute()

Update guild

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
	guildUpdateGuildRequest := *openapiclient.NewGuildUpdateGuildRequest() // GuildUpdateGuildRequest | Update guild data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildAPI.GuildGuildIdPatch(context.Background(), guildId).GuildUpdateGuildRequest(guildUpdateGuildRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildAPI.GuildGuildIdPatch``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdPatch`: DtoGuild
	fmt.Fprintf(os.Stdout, "Response from `GuildAPI.GuildGuildIdPatch`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdPatchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **guildUpdateGuildRequest** | [**GuildUpdateGuildRequest**](GuildUpdateGuildRequest.md) | Update guild data | 

### Return type

[**DtoGuild**](DtoGuild.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GuildPost

> DtoGuild GuildPost(ctx).GuildCreateGuildRequest(guildCreateGuildRequest).Execute()

Create guild

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
	guildCreateGuildRequest := *openapiclient.NewGuildCreateGuildRequest() // GuildCreateGuildRequest | Guild data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildAPI.GuildPost(context.Background()).GuildCreateGuildRequest(guildCreateGuildRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildAPI.GuildPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildPost`: DtoGuild
	fmt.Fprintf(os.Stdout, "Response from `GuildAPI.GuildPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGuildPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **guildCreateGuildRequest** | [**GuildCreateGuildRequest**](GuildCreateGuildRequest.md) | Guild data | 

### Return type

[**DtoGuild**](DtoGuild.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

