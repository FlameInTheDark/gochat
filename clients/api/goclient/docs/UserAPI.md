# \UserAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**UserMeChannelsGroupPost**](UserAPI.md#UserMeChannelsGroupPost) | **Post** /user/me/channels/group | Create group DM channel
[**UserMeChannelsPost**](UserAPI.md#UserMeChannelsPost) | **Post** /user/me/channels | Create DM channel
[**UserMeGuildsGet**](UserAPI.md#UserMeGuildsGet) | **Get** /user/me/guilds | Get user guilds
[**UserMeGuildsGuildIdDelete**](UserAPI.md#UserMeGuildsGuildIdDelete) | **Delete** /user/me/guilds/{guild_id} | Leave guild
[**UserMeGuildsGuildIdMemberGet**](UserAPI.md#UserMeGuildsGuildIdMemberGet) | **Get** /user/me/guilds/{guild_id}/member | Get user guild member
[**UserMePatch**](UserAPI.md#UserMePatch) | **Patch** /user/me | Get user
[**UserMeSettingsGet**](UserAPI.md#UserMeSettingsGet) | **Get** /user/me/settings | Get current user settings (optional version gating)
[**UserMeSettingsPost**](UserAPI.md#UserMeSettingsPost) | **Post** /user/me/settings | Update current user settings (replaces and bumps version)
[**UserUserIdGet**](UserAPI.md#UserUserIdGet) | **Get** /user/{user_id} | Get user



## UserMeChannelsGroupPost

> DtoChannel UserMeChannelsGroupPost(ctx).UserCreateDMManyRequest(userCreateDMManyRequest).Execute()

Create group DM channel

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
	userCreateDMManyRequest := *openapiclient.NewUserCreateDMManyRequest() // UserCreateDMManyRequest | Group DM data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UserAPI.UserMeChannelsGroupPost(context.Background()).UserCreateDMManyRequest(userCreateDMManyRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAPI.UserMeChannelsGroupPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UserMeChannelsGroupPost`: DtoChannel
	fmt.Fprintf(os.Stdout, "Response from `UserAPI.UserMeChannelsGroupPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUserMeChannelsGroupPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userCreateDMManyRequest** | [**UserCreateDMManyRequest**](UserCreateDMManyRequest.md) | Group DM data | 

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


## UserMeChannelsPost

> DtoChannel UserMeChannelsPost(ctx).UserCreateDMRequest(userCreateDMRequest).Execute()

Create DM channel

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
	userCreateDMRequest := *openapiclient.NewUserCreateDMRequest() // UserCreateDMRequest | Recipient data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UserAPI.UserMeChannelsPost(context.Background()).UserCreateDMRequest(userCreateDMRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAPI.UserMeChannelsPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UserMeChannelsPost`: DtoChannel
	fmt.Fprintf(os.Stdout, "Response from `UserAPI.UserMeChannelsPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUserMeChannelsPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userCreateDMRequest** | [**UserCreateDMRequest**](UserCreateDMRequest.md) | Recipient data | 

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


## UserMeGuildsGet

> []DtoGuild UserMeGuildsGet(ctx).Execute()

Get user guilds

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
	resp, r, err := apiClient.UserAPI.UserMeGuildsGet(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAPI.UserMeGuildsGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UserMeGuildsGet`: []DtoGuild
	fmt.Fprintf(os.Stdout, "Response from `UserAPI.UserMeGuildsGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiUserMeGuildsGetRequest struct via the builder pattern


### Return type

[**[]DtoGuild**](DtoGuild.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UserMeGuildsGuildIdDelete

> string UserMeGuildsGuildIdDelete(ctx, guildId).Execute()

Leave guild

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
	guildId := "2230469276416868352" // string | Guild id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UserAPI.UserMeGuildsGuildIdDelete(context.Background(), guildId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAPI.UserMeGuildsGuildIdDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UserMeGuildsGuildIdDelete`: string
	fmt.Fprintf(os.Stdout, "Response from `UserAPI.UserMeGuildsGuildIdDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **string** | Guild id | 

### Other Parameters

Other parameters are passed through a pointer to a apiUserMeGuildsGuildIdDeleteRequest struct via the builder pattern


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


## UserMeGuildsGuildIdMemberGet

> DtoMember UserMeGuildsGuildIdMemberGet(ctx, guildId).Execute()

Get user guild member

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
	resp, r, err := apiClient.UserAPI.UserMeGuildsGuildIdMemberGet(context.Background(), guildId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAPI.UserMeGuildsGuildIdMemberGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UserMeGuildsGuildIdMemberGet`: DtoMember
	fmt.Fprintf(os.Stdout, "Response from `UserAPI.UserMeGuildsGuildIdMemberGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild id | 

### Other Parameters

Other parameters are passed through a pointer to a apiUserMeGuildsGuildIdMemberGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**DtoMember**](DtoMember.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UserMePatch

> string UserMePatch(ctx).UserModifyUserRequest(userModifyUserRequest).Execute()

Get user

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
	userModifyUserRequest := *openapiclient.NewUserModifyUserRequest() // UserModifyUserRequest | Modify user data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UserAPI.UserMePatch(context.Background()).UserModifyUserRequest(userModifyUserRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAPI.UserMePatch``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UserMePatch`: string
	fmt.Fprintf(os.Stdout, "Response from `UserAPI.UserMePatch`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUserMePatchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userModifyUserRequest** | [**UserModifyUserRequest**](UserModifyUserRequest.md) | Modify user data | 

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


## UserMeSettingsGet

> UserUserSettingsResponse UserMeSettingsGet(ctx).Version(version).Execute()

Get current user settings (optional version gating)

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
	version := int32(56) // int32 | Client known version (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UserAPI.UserMeSettingsGet(context.Background()).Version(version).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAPI.UserMeSettingsGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UserMeSettingsGet`: UserUserSettingsResponse
	fmt.Fprintf(os.Stdout, "Response from `UserAPI.UserMeSettingsGet`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUserMeSettingsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **version** | **int32** | Client known version | 

### Return type

[**UserUserSettingsResponse**](UserUserSettingsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## UserMeSettingsPost

> string UserMeSettingsPost(ctx).ModelUserSettingsData(modelUserSettingsData).Execute()

Update current user settings (replaces and bumps version)

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
	modelUserSettingsData := *openapiclient.NewModelUserSettingsData() // ModelUserSettingsData | User settings

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UserAPI.UserMeSettingsPost(context.Background()).ModelUserSettingsData(modelUserSettingsData).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAPI.UserMeSettingsPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UserMeSettingsPost`: string
	fmt.Fprintf(os.Stdout, "Response from `UserAPI.UserMeSettingsPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiUserMeSettingsPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **modelUserSettingsData** | [**ModelUserSettingsData**](ModelUserSettingsData.md) | User settings | 

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


## UserUserIdGet

> DtoUser UserUserIdGet(ctx, userId).Execute()

Get user

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
	userId := "userId_example" // string | User ID or 'me'

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.UserAPI.UserUserIdGet(context.Background(), userId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `UserAPI.UserUserIdGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `UserUserIdGet`: DtoUser
	fmt.Fprintf(os.Stdout, "Response from `UserAPI.UserUserIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**userId** | **string** | User ID or &#39;me&#39; | 

### Other Parameters

Other parameters are passed through a pointer to a apiUserUserIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**DtoUser**](DtoUser.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

