# \GuildRolesAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GuildGuildIdChannelChannelIdRolesGet**](GuildRolesAPI.md#GuildGuildIdChannelChannelIdRolesGet) | **Get** /guild/{guild_id}/channel/{channel_id}/roles | List channel role permissions
[**GuildGuildIdChannelChannelIdRolesRoleIdDelete**](GuildRolesAPI.md#GuildGuildIdChannelChannelIdRolesRoleIdDelete) | **Delete** /guild/{guild_id}/channel/{channel_id}/roles/{role_id} | Remove channel role permission
[**GuildGuildIdChannelChannelIdRolesRoleIdGet**](GuildRolesAPI.md#GuildGuildIdChannelChannelIdRolesRoleIdGet) | **Get** /guild/{guild_id}/channel/{channel_id}/roles/{role_id} | Get channel role permission
[**GuildGuildIdChannelChannelIdRolesRoleIdPatch**](GuildRolesAPI.md#GuildGuildIdChannelChannelIdRolesRoleIdPatch) | **Patch** /guild/{guild_id}/channel/{channel_id}/roles/{role_id} | Update channel role permission
[**GuildGuildIdChannelChannelIdRolesRoleIdPut**](GuildRolesAPI.md#GuildGuildIdChannelChannelIdRolesRoleIdPut) | **Put** /guild/{guild_id}/channel/{channel_id}/roles/{role_id} | Set channel role permission (create or replace)
[**GuildGuildIdMemberUserIdRolesGet**](GuildRolesAPI.md#GuildGuildIdMemberUserIdRolesGet) | **Get** /guild/{guild_id}/member/{user_id}/roles | Get member roles
[**GuildGuildIdMemberUserIdRolesRoleIdDelete**](GuildRolesAPI.md#GuildGuildIdMemberUserIdRolesRoleIdDelete) | **Delete** /guild/{guild_id}/member/{user_id}/roles/{role_id} | Remove role from member
[**GuildGuildIdMemberUserIdRolesRoleIdPut**](GuildRolesAPI.md#GuildGuildIdMemberUserIdRolesRoleIdPut) | **Put** /guild/{guild_id}/member/{user_id}/roles/{role_id} | Assign role to member
[**GuildGuildIdRolesGet**](GuildRolesAPI.md#GuildGuildIdRolesGet) | **Get** /guild/{guild_id}/roles | Get guild roles
[**GuildGuildIdRolesPost**](GuildRolesAPI.md#GuildGuildIdRolesPost) | **Post** /guild/{guild_id}/roles | Create guild role
[**GuildGuildIdRolesRoleIdDelete**](GuildRolesAPI.md#GuildGuildIdRolesRoleIdDelete) | **Delete** /guild/{guild_id}/roles/{role_id} | Delete guild role
[**GuildGuildIdRolesRoleIdPatch**](GuildRolesAPI.md#GuildGuildIdRolesRoleIdPatch) | **Patch** /guild/{guild_id}/roles/{role_id} | Update guild role



## GuildGuildIdChannelChannelIdRolesGet

> []GuildChannelRolePermission GuildGuildIdChannelChannelIdRolesGet(ctx, guildId, channelId).Execute()

List channel role permissions

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
	resp, r, err := apiClient.GuildRolesAPI.GuildGuildIdChannelChannelIdRolesGet(context.Background(), guildId, channelId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildRolesAPI.GuildGuildIdChannelChannelIdRolesGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdChannelChannelIdRolesGet`: []GuildChannelRolePermission
	fmt.Fprintf(os.Stdout, "Response from `GuildRolesAPI.GuildGuildIdChannelChannelIdRolesGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 
**channelId** | **int32** | Channel ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdChannelChannelIdRolesGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**[]GuildChannelRolePermission**](GuildChannelRolePermission.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GuildGuildIdChannelChannelIdRolesRoleIdDelete

> string GuildGuildIdChannelChannelIdRolesRoleIdDelete(ctx, guildId, channelId, roleId).Execute()

Remove channel role permission

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
	roleId := int32(56) // int32 | Role ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildRolesAPI.GuildGuildIdChannelChannelIdRolesRoleIdDelete(context.Background(), guildId, channelId, roleId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildRolesAPI.GuildGuildIdChannelChannelIdRolesRoleIdDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdChannelChannelIdRolesRoleIdDelete`: string
	fmt.Fprintf(os.Stdout, "Response from `GuildRolesAPI.GuildGuildIdChannelChannelIdRolesRoleIdDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 
**channelId** | **int32** | Channel ID | 
**roleId** | **int32** | Role ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdChannelChannelIdRolesRoleIdDeleteRequest struct via the builder pattern


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


## GuildGuildIdChannelChannelIdRolesRoleIdGet

> GuildChannelRolePermission GuildGuildIdChannelChannelIdRolesRoleIdGet(ctx, guildId, channelId, roleId).Execute()

Get channel role permission

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
	roleId := int32(56) // int32 | Role ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildRolesAPI.GuildGuildIdChannelChannelIdRolesRoleIdGet(context.Background(), guildId, channelId, roleId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildRolesAPI.GuildGuildIdChannelChannelIdRolesRoleIdGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdChannelChannelIdRolesRoleIdGet`: GuildChannelRolePermission
	fmt.Fprintf(os.Stdout, "Response from `GuildRolesAPI.GuildGuildIdChannelChannelIdRolesRoleIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 
**channelId** | **int32** | Channel ID | 
**roleId** | **int32** | Role ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdChannelChannelIdRolesRoleIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

[**GuildChannelRolePermission**](GuildChannelRolePermission.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GuildGuildIdChannelChannelIdRolesRoleIdPatch

> string GuildGuildIdChannelChannelIdRolesRoleIdPatch(ctx, guildId, channelId, roleId).GuildChannelRolePermissionRequest(guildChannelRolePermissionRequest).Execute()

Update channel role permission

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
	roleId := int32(56) // int32 | Role ID
	guildChannelRolePermissionRequest := *openapiclient.NewGuildChannelRolePermissionRequest() // GuildChannelRolePermissionRequest | Permission mask

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildRolesAPI.GuildGuildIdChannelChannelIdRolesRoleIdPatch(context.Background(), guildId, channelId, roleId).GuildChannelRolePermissionRequest(guildChannelRolePermissionRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildRolesAPI.GuildGuildIdChannelChannelIdRolesRoleIdPatch``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdChannelChannelIdRolesRoleIdPatch`: string
	fmt.Fprintf(os.Stdout, "Response from `GuildRolesAPI.GuildGuildIdChannelChannelIdRolesRoleIdPatch`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 
**channelId** | **int32** | Channel ID | 
**roleId** | **int32** | Role ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdChannelChannelIdRolesRoleIdPatchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **guildChannelRolePermissionRequest** | [**GuildChannelRolePermissionRequest**](GuildChannelRolePermissionRequest.md) | Permission mask | 

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


## GuildGuildIdChannelChannelIdRolesRoleIdPut

> string GuildGuildIdChannelChannelIdRolesRoleIdPut(ctx, guildId, channelId, roleId).GuildChannelRolePermissionRequest(guildChannelRolePermissionRequest).Execute()

Set channel role permission (create or replace)

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
	roleId := int32(56) // int32 | Role ID
	guildChannelRolePermissionRequest := *openapiclient.NewGuildChannelRolePermissionRequest() // GuildChannelRolePermissionRequest | Permission mask

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildRolesAPI.GuildGuildIdChannelChannelIdRolesRoleIdPut(context.Background(), guildId, channelId, roleId).GuildChannelRolePermissionRequest(guildChannelRolePermissionRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildRolesAPI.GuildGuildIdChannelChannelIdRolesRoleIdPut``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdChannelChannelIdRolesRoleIdPut`: string
	fmt.Fprintf(os.Stdout, "Response from `GuildRolesAPI.GuildGuildIdChannelChannelIdRolesRoleIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 
**channelId** | **int32** | Channel ID | 
**roleId** | **int32** | Role ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdChannelChannelIdRolesRoleIdPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **guildChannelRolePermissionRequest** | [**GuildChannelRolePermissionRequest**](GuildChannelRolePermissionRequest.md) | Permission mask | 

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


## GuildGuildIdMemberUserIdRolesGet

> []DtoRole GuildGuildIdMemberUserIdRolesGet(ctx, guildId, userId).Execute()

Get member roles

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
	userId := int32(56) // int32 | User ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildRolesAPI.GuildGuildIdMemberUserIdRolesGet(context.Background(), guildId, userId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildRolesAPI.GuildGuildIdMemberUserIdRolesGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdMemberUserIdRolesGet`: []DtoRole
	fmt.Fprintf(os.Stdout, "Response from `GuildRolesAPI.GuildGuildIdMemberUserIdRolesGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 
**userId** | **int32** | User ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdMemberUserIdRolesGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

[**[]DtoRole**](DtoRole.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GuildGuildIdMemberUserIdRolesRoleIdDelete

> string GuildGuildIdMemberUserIdRolesRoleIdDelete(ctx, guildId, userId, roleId).Execute()

Remove role from member

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
	userId := int32(56) // int32 | User ID
	roleId := int32(56) // int32 | Role ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildRolesAPI.GuildGuildIdMemberUserIdRolesRoleIdDelete(context.Background(), guildId, userId, roleId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildRolesAPI.GuildGuildIdMemberUserIdRolesRoleIdDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdMemberUserIdRolesRoleIdDelete`: string
	fmt.Fprintf(os.Stdout, "Response from `GuildRolesAPI.GuildGuildIdMemberUserIdRolesRoleIdDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 
**userId** | **int32** | User ID | 
**roleId** | **int32** | Role ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdMemberUserIdRolesRoleIdDeleteRequest struct via the builder pattern


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


## GuildGuildIdMemberUserIdRolesRoleIdPut

> string GuildGuildIdMemberUserIdRolesRoleIdPut(ctx, guildId, userId, roleId).Execute()

Assign role to member

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
	userId := int32(56) // int32 | User ID
	roleId := int32(56) // int32 | Role ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildRolesAPI.GuildGuildIdMemberUserIdRolesRoleIdPut(context.Background(), guildId, userId, roleId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildRolesAPI.GuildGuildIdMemberUserIdRolesRoleIdPut``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdMemberUserIdRolesRoleIdPut`: string
	fmt.Fprintf(os.Stdout, "Response from `GuildRolesAPI.GuildGuildIdMemberUserIdRolesRoleIdPut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 
**userId** | **int32** | User ID | 
**roleId** | **int32** | Role ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdMemberUserIdRolesRoleIdPutRequest struct via the builder pattern


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


## GuildGuildIdRolesGet

> []DtoRole GuildGuildIdRolesGet(ctx, guildId).Execute()

Get guild roles

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
	resp, r, err := apiClient.GuildRolesAPI.GuildGuildIdRolesGet(context.Background(), guildId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildRolesAPI.GuildGuildIdRolesGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdRolesGet`: []DtoRole
	fmt.Fprintf(os.Stdout, "Response from `GuildRolesAPI.GuildGuildIdRolesGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdRolesGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]DtoRole**](DtoRole.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GuildGuildIdRolesPost

> DtoRole GuildGuildIdRolesPost(ctx, guildId).GuildCreateGuildRoleRequest(guildCreateGuildRoleRequest).Execute()

Create guild role

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
	guildCreateGuildRoleRequest := *openapiclient.NewGuildCreateGuildRoleRequest() // GuildCreateGuildRoleRequest | Role data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildRolesAPI.GuildGuildIdRolesPost(context.Background(), guildId).GuildCreateGuildRoleRequest(guildCreateGuildRoleRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildRolesAPI.GuildGuildIdRolesPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdRolesPost`: DtoRole
	fmt.Fprintf(os.Stdout, "Response from `GuildRolesAPI.GuildGuildIdRolesPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdRolesPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **guildCreateGuildRoleRequest** | [**GuildCreateGuildRoleRequest**](GuildCreateGuildRoleRequest.md) | Role data | 

### Return type

[**DtoRole**](DtoRole.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GuildGuildIdRolesRoleIdDelete

> string GuildGuildIdRolesRoleIdDelete(ctx, guildId, roleId).Execute()

Delete guild role

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
	roleId := int32(56) // int32 | Role ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildRolesAPI.GuildGuildIdRolesRoleIdDelete(context.Background(), guildId, roleId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildRolesAPI.GuildGuildIdRolesRoleIdDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdRolesRoleIdDelete`: string
	fmt.Fprintf(os.Stdout, "Response from `GuildRolesAPI.GuildGuildIdRolesRoleIdDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 
**roleId** | **int32** | Role ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdRolesRoleIdDeleteRequest struct via the builder pattern


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


## GuildGuildIdRolesRoleIdPatch

> DtoRole GuildGuildIdRolesRoleIdPatch(ctx, guildId, roleId).GuildPatchGuildRoleRequest(guildPatchGuildRoleRequest).Execute()

Update guild role

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
	roleId := int32(56) // int32 | Role ID
	guildPatchGuildRoleRequest := *openapiclient.NewGuildPatchGuildRoleRequest() // GuildPatchGuildRoleRequest | Role changes

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildRolesAPI.GuildGuildIdRolesRoleIdPatch(context.Background(), guildId, roleId).GuildPatchGuildRoleRequest(guildPatchGuildRoleRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildRolesAPI.GuildGuildIdRolesRoleIdPatch``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdRolesRoleIdPatch`: DtoRole
	fmt.Fprintf(os.Stdout, "Response from `GuildRolesAPI.GuildGuildIdRolesRoleIdPatch`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild ID | 
**roleId** | **int32** | Role ID | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdRolesRoleIdPatchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **guildPatchGuildRoleRequest** | [**GuildPatchGuildRoleRequest**](GuildPatchGuildRoleRequest.md) | Role changes | 

### Return type

[**DtoRole**](DtoRole.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

