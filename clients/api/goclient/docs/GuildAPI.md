# \GuildAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GuildGuildIdCategoryCategoryIdDelete**](GuildAPI.md#GuildGuildIdCategoryCategoryIdDelete) | **Delete** /guild/{guild_id}/category/{category_id} | Delete channel category
[**GuildGuildIdCategoryPost**](GuildAPI.md#GuildGuildIdCategoryPost) | **Post** /guild/{guild_id}/category | Create guild channel category
[**GuildGuildIdChannelChannelIdDelete**](GuildAPI.md#GuildGuildIdChannelChannelIdDelete) | **Delete** /guild/{guild_id}/channel/{channel_id} | Delete channel
[**GuildGuildIdChannelChannelIdGet**](GuildAPI.md#GuildGuildIdChannelChannelIdGet) | **Get** /guild/{guild_id}/channel/{channel_id} | Get guild channel
[**GuildGuildIdChannelChannelIdPatch**](GuildAPI.md#GuildGuildIdChannelChannelIdPatch) | **Patch** /guild/{guild_id}/channel/{channel_id} | Change channels data
[**GuildGuildIdChannelGet**](GuildAPI.md#GuildGuildIdChannelGet) | **Get** /guild/{guild_id}/channel | Get guild channels
[**GuildGuildIdChannelOrderPatch**](GuildAPI.md#GuildGuildIdChannelOrderPatch) | **Patch** /guild/{guild_id}/channel/order | Change channels order
[**GuildGuildIdChannelPost**](GuildAPI.md#GuildGuildIdChannelPost) | **Post** /guild/{guild_id}/channel | Create guild channel
[**GuildGuildIdDelete**](GuildAPI.md#GuildGuildIdDelete) | **Delete** /guild/{guild_id} | Delete guild
[**GuildGuildIdGet**](GuildAPI.md#GuildGuildIdGet) | **Get** /guild/{guild_id} | Get guild
[**GuildGuildIdIconPost**](GuildAPI.md#GuildGuildIdIconPost) | **Post** /guild/{guild_id}/icon | Create guild icon metadata
[**GuildGuildIdIconsGet**](GuildAPI.md#GuildGuildIdIconsGet) | **Get** /guild/{guild_id}/icons | List guild icons
[**GuildGuildIdIconsIconIdDelete**](GuildAPI.md#GuildGuildIdIconsIconIdDelete) | **Delete** /guild/{guild_id}/icons/{icon_id} | Delete guild icon by ID
[**GuildGuildIdMembersGet**](GuildAPI.md#GuildGuildIdMembersGet) | **Get** /guild/{guild_id}/members | Get guild members
[**GuildGuildIdPatch**](GuildAPI.md#GuildGuildIdPatch) | **Patch** /guild/{guild_id} | Update guild
[**GuildGuildIdVoiceChannelIdJoinPost**](GuildAPI.md#GuildGuildIdVoiceChannelIdJoinPost) | **Post** /guild/{guild_id}/voice/{channel_id}/join | Join voice channel (get SFU signaling info)
[**GuildGuildIdVoiceChannelIdRegionPatch**](GuildAPI.md#GuildGuildIdVoiceChannelIdRegionPatch) | **Patch** /guild/{guild_id}/voice/{channel_id}/region | Set channel voice region
[**GuildGuildIdVoiceMovePost**](GuildAPI.md#GuildGuildIdVoiceMovePost) | **Post** /guild/{guild_id}/voice/move | Move member to voice channel
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
	guildId := int32(2230469276416868352) // int32 | Guild ID
	categoryId := int32(2230469276416868352) // int32 | Category ID (actually a channel with special type)

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
	guildId := int32(2230469276416868352) // int32 | Guild ID
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
	guildId := int32(2230469276416868352) // int32 | Guild ID
	channelId := int32(2230469276416868352) // int32 | Channel ID

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
	guildId := int32(2230469276416868352) // int32 | Guild id
	channelId := int32(2230469276416868352) // int32 | Channel id

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


## GuildGuildIdChannelChannelIdPatch

> DtoChannel GuildGuildIdChannelChannelIdPatch(ctx, guildId, channelId).GuildPatchGuildChannelRequest(guildPatchGuildChannelRequest).Execute()

Change channels data

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
	guildId := int32(2230469276416868352) // int32 | Guild ID
	channelId := int32(2230469276416868352) // int32 | Channel ID
	guildPatchGuildChannelRequest := *openapiclient.NewGuildPatchGuildChannelRequest() // GuildPatchGuildChannelRequest | Request body

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildAPI.GuildGuildIdChannelChannelIdPatch(context.Background(), guildId, channelId).GuildPatchGuildChannelRequest(guildPatchGuildChannelRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildAPI.GuildGuildIdChannelChannelIdPatch``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdChannelChannelIdPatch`: DtoChannel
	fmt.Fprintf(os.Stdout, "Response from `GuildAPI.GuildGuildIdChannelChannelIdPatch`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 
**channelId** | **int32** | Channel ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdChannelChannelIdPatchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **guildPatchGuildChannelRequest** | [**GuildPatchGuildChannelRequest**](GuildPatchGuildChannelRequest.md) | Request body | 

### Return type

[**DtoChannel**](DtoChannel.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
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
	guildId := int32(2230469276416868352) // int32 | Guild id

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


## GuildGuildIdChannelOrderPatch

> string GuildGuildIdChannelOrderPatch(ctx, guildId).GuildPatchGuildChannelOrderRequest(guildPatchGuildChannelOrderRequest).Execute()

Change channels order

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
	guildId := int32(2230469276416868352) // int32 | Guild ID
	guildPatchGuildChannelOrderRequest := *openapiclient.NewGuildPatchGuildChannelOrderRequest() // GuildPatchGuildChannelOrderRequest | Update channel order data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildAPI.GuildGuildIdChannelOrderPatch(context.Background(), guildId).GuildPatchGuildChannelOrderRequest(guildPatchGuildChannelOrderRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildAPI.GuildGuildIdChannelOrderPatch``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdChannelOrderPatch`: string
	fmt.Fprintf(os.Stdout, "Response from `GuildAPI.GuildGuildIdChannelOrderPatch`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdChannelOrderPatchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **guildPatchGuildChannelOrderRequest** | [**GuildPatchGuildChannelOrderRequest**](GuildPatchGuildChannelOrderRequest.md) | Update channel order data | 

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
	guildId := int32(2230469276416868352) // int32 | Guild ID
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


## GuildGuildIdDelete

> string GuildGuildIdDelete(ctx, guildId).Execute()

Delete guild



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
	guildId := int32(2230469276416868352) // int32 | Guild ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildAPI.GuildGuildIdDelete(context.Background(), guildId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildAPI.GuildGuildIdDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdDelete`: string
	fmt.Fprintf(os.Stdout, "Response from `GuildAPI.GuildGuildIdDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdDeleteRequest struct via the builder pattern


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
	guildId := int32(2230469276416868352) // int32 | Guild id

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


## GuildGuildIdIconPost

> DtoIconUpload GuildGuildIdIconPost(ctx, guildId).GuildCreateIconRequest(guildCreateIconRequest).Execute()

Create guild icon metadata



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
	guildCreateIconRequest := *openapiclient.NewGuildCreateIconRequest() // GuildCreateIconRequest | Icon creation request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildAPI.GuildGuildIdIconPost(context.Background(), guildId).GuildCreateIconRequest(guildCreateIconRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildAPI.GuildGuildIdIconPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdIconPost`: DtoIconUpload
	fmt.Fprintf(os.Stdout, "Response from `GuildAPI.GuildGuildIdIconPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdIconPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **guildCreateIconRequest** | [**GuildCreateIconRequest**](GuildCreateIconRequest.md) | Icon creation request | 

### Return type

[**DtoIconUpload**](DtoIconUpload.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GuildGuildIdIconsGet

> []DtoIcon GuildGuildIdIconsGet(ctx, guildId).Execute()

List guild icons



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

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildAPI.GuildGuildIdIconsGet(context.Background(), guildId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildAPI.GuildGuildIdIconsGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdIconsGet`: []DtoIcon
	fmt.Fprintf(os.Stdout, "Response from `GuildAPI.GuildGuildIdIconsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdIconsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]DtoIcon**](DtoIcon.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GuildGuildIdIconsIconIdDelete

> string GuildGuildIdIconsIconIdDelete(ctx, guildId, iconId).Execute()

Delete guild icon by ID



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
	iconId := int32(56) // int32 | Icon ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildAPI.GuildGuildIdIconsIconIdDelete(context.Background(), guildId, iconId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildAPI.GuildGuildIdIconsIconIdDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdIconsIconIdDelete`: string
	fmt.Fprintf(os.Stdout, "Response from `GuildAPI.GuildGuildIdIconsIconIdDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 
**iconId** | **int32** | Icon ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdIconsIconIdDeleteRequest struct via the builder pattern


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


## GuildGuildIdMembersGet

> []DtoMember GuildGuildIdMembersGet(ctx, guildId).Execute()

Get guild members

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
	guildId := int32(2230469276416868352) // int32 | Guild ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildAPI.GuildGuildIdMembersGet(context.Background(), guildId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildAPI.GuildGuildIdMembersGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdMembersGet`: []DtoMember
	fmt.Fprintf(os.Stdout, "Response from `GuildAPI.GuildGuildIdMembersGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdMembersGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]DtoMember**](DtoMember.md)

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
	guildId := int32(2230469276416868352) // int32 | Guild ID
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


## GuildGuildIdVoiceChannelIdJoinPost

> GuildJoinVoiceResponse GuildGuildIdVoiceChannelIdJoinPost(ctx, guildId, channelId).Execute()

Join voice channel (get SFU signaling info)



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
	resp, r, err := apiClient.GuildAPI.GuildGuildIdVoiceChannelIdJoinPost(context.Background(), guildId, channelId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildAPI.GuildGuildIdVoiceChannelIdJoinPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdVoiceChannelIdJoinPost`: GuildJoinVoiceResponse
	fmt.Fprintf(os.Stdout, "Response from `GuildAPI.GuildGuildIdVoiceChannelIdJoinPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 
**channelId** | **int32** | Channel ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdVoiceChannelIdJoinPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**GuildJoinVoiceResponse**](GuildJoinVoiceResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GuildGuildIdVoiceChannelIdRegionPatch

> GuildSetVoiceRegionResponse GuildGuildIdVoiceChannelIdRegionPatch(ctx, guildId, channelId).GuildSetVoiceRegionRequest(guildSetVoiceRegionRequest).Execute()

Set channel voice region



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
	guildSetVoiceRegionRequest := *openapiclient.NewGuildSetVoiceRegionRequest() // GuildSetVoiceRegionRequest | Region payload

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildAPI.GuildGuildIdVoiceChannelIdRegionPatch(context.Background(), guildId, channelId).GuildSetVoiceRegionRequest(guildSetVoiceRegionRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildAPI.GuildGuildIdVoiceChannelIdRegionPatch``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdVoiceChannelIdRegionPatch`: GuildSetVoiceRegionResponse
	fmt.Fprintf(os.Stdout, "Response from `GuildAPI.GuildGuildIdVoiceChannelIdRegionPatch`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 
**channelId** | **int32** | Channel ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdVoiceChannelIdRegionPatchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **guildSetVoiceRegionRequest** | [**GuildSetVoiceRegionRequest**](GuildSetVoiceRegionRequest.md) | Region payload | 

### Return type

[**GuildSetVoiceRegionResponse**](GuildSetVoiceRegionResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GuildGuildIdVoiceMovePost

> GuildMoveMemberResponse GuildGuildIdVoiceMovePost(ctx, guildId).GuildMoveMemberRequest(guildMoveMemberRequest).Execute()

Move member to voice channel



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
	guildMoveMemberRequest := *openapiclient.NewGuildMoveMemberRequest() // GuildMoveMemberRequest | Move request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildAPI.GuildGuildIdVoiceMovePost(context.Background(), guildId).GuildMoveMemberRequest(guildMoveMemberRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildAPI.GuildGuildIdVoiceMovePost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdVoiceMovePost`: GuildMoveMemberResponse
	fmt.Fprintf(os.Stdout, "Response from `GuildAPI.GuildGuildIdVoiceMovePost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdVoiceMovePostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **guildMoveMemberRequest** | [**GuildMoveMemberRequest**](GuildMoveMemberRequest.md) | Move request | 

### Return type

[**GuildMoveMemberResponse**](GuildMoveMemberResponse.md)

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

