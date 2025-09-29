# \GuildInvitesAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GuildInvitesAcceptInviteCodePost**](GuildInvitesAPI.md#GuildInvitesAcceptInviteCodePost) | **Post** /guild/invites/accept/{invite_code} | Accept invite and join guild
[**GuildInvitesGuildIdGet**](GuildInvitesAPI.md#GuildInvitesGuildIdGet) | **Get** /guild/invites/{guild_id} | List active invites for guild
[**GuildInvitesGuildIdInviteIdDelete**](GuildInvitesAPI.md#GuildInvitesGuildIdInviteIdDelete) | **Delete** /guild/invites/{guild_id}/{invite_id} | Delete an invite by id
[**GuildInvitesGuildIdPost**](GuildInvitesAPI.md#GuildInvitesGuildIdPost) | **Post** /guild/invites/{guild_id} | Create a new invite
[**GuildInvitesReceiveInviteCodeGet**](GuildInvitesAPI.md#GuildInvitesReceiveInviteCodeGet) | **Get** /guild/invites/receive/{invite_code} | Get invite info by code



## GuildInvitesAcceptInviteCodePost

> DtoGuild GuildInvitesAcceptInviteCodePost(ctx, inviteCode).Execute()

Accept invite and join guild

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
	inviteCode := "PWBJ124G" // string | Invite code

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildInvitesAPI.GuildInvitesAcceptInviteCodePost(context.Background(), inviteCode).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildInvitesAPI.GuildInvitesAcceptInviteCodePost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildInvitesAcceptInviteCodePost`: DtoGuild
	fmt.Fprintf(os.Stdout, "Response from `GuildInvitesAPI.GuildInvitesAcceptInviteCodePost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**inviteCode** | **string** | Invite code | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildInvitesAcceptInviteCodePostRequest struct via the builder pattern


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


## GuildInvitesGuildIdGet

> []DtoGuildInvite GuildInvitesGuildIdGet(ctx, guildId).Execute()

List active invites for guild

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
	resp, r, err := apiClient.GuildInvitesAPI.GuildInvitesGuildIdGet(context.Background(), guildId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildInvitesAPI.GuildInvitesGuildIdGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildInvitesGuildIdGet`: []DtoGuildInvite
	fmt.Fprintf(os.Stdout, "Response from `GuildInvitesAPI.GuildInvitesGuildIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild id | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildInvitesGuildIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]DtoGuildInvite**](DtoGuildInvite.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GuildInvitesGuildIdInviteIdDelete

> string GuildInvitesGuildIdInviteIdDelete(ctx, guildId, inviteId).Execute()

Delete an invite by id

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
	inviteId := int32(2230469276416868352) // int32 | Invite id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildInvitesAPI.GuildInvitesGuildIdInviteIdDelete(context.Background(), guildId, inviteId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildInvitesAPI.GuildInvitesGuildIdInviteIdDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildInvitesGuildIdInviteIdDelete`: string
	fmt.Fprintf(os.Stdout, "Response from `GuildInvitesAPI.GuildInvitesGuildIdInviteIdDelete`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild id | 
**inviteId** | **int32** | Invite id | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildInvitesGuildIdInviteIdDeleteRequest struct via the builder pattern


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


## GuildInvitesGuildIdPost

> DtoGuildInvite GuildInvitesGuildIdPost(ctx, guildId).GuildCreateInviteRequest(guildCreateInviteRequest).Execute()

Create a new invite

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
	guildCreateInviteRequest := *openapiclient.NewGuildCreateInviteRequest() // GuildCreateInviteRequest | Invite options

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildInvitesAPI.GuildInvitesGuildIdPost(context.Background(), guildId).GuildCreateInviteRequest(guildCreateInviteRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildInvitesAPI.GuildInvitesGuildIdPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildInvitesGuildIdPost`: DtoGuildInvite
	fmt.Fprintf(os.Stdout, "Response from `GuildInvitesAPI.GuildInvitesGuildIdPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild id | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildInvitesGuildIdPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **guildCreateInviteRequest** | [**GuildCreateInviteRequest**](GuildCreateInviteRequest.md) | Invite options | 

### Return type

[**DtoGuildInvite**](DtoGuildInvite.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GuildInvitesReceiveInviteCodeGet

> DtoInvitePreview GuildInvitesReceiveInviteCodeGet(ctx, inviteCode).Execute()

Get invite info by code

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
	inviteCode := "PWBJ124G" // string | Invite code

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildInvitesAPI.GuildInvitesReceiveInviteCodeGet(context.Background(), inviteCode).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildInvitesAPI.GuildInvitesReceiveInviteCodeGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildInvitesReceiveInviteCodeGet`: DtoInvitePreview
	fmt.Fprintf(os.Stdout, "Response from `GuildInvitesAPI.GuildInvitesReceiveInviteCodeGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**inviteCode** | **string** | Invite code | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildInvitesReceiveInviteCodeGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**DtoInvitePreview**](DtoInvitePreview.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

