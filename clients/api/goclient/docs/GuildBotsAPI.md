# \GuildBotsAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GuildBotsAuthorizeGuildsGet**](GuildBotsAPI.md#GuildBotsAuthorizeGuildsGet) | **Get** /guild/bots/authorize-guilds | List guilds available for bot authorization
[**GuildGuildIdBotsBotIdDelete**](GuildBotsAPI.md#GuildGuildIdBotsBotIdDelete) | **Delete** /guild/{guild_id}/bots/{bot_id} | Remove a bot from a guild
[**GuildGuildIdBotsGet**](GuildBotsAPI.md#GuildGuildIdBotsGet) | **Get** /guild/{guild_id}/bots | List installed guild bots
[**GuildGuildIdBotsPost**](GuildBotsAPI.md#GuildGuildIdBotsPost) | **Post** /guild/{guild_id}/bots | Install a bot into a guild



## GuildBotsAuthorizeGuildsGet

> []DtoGuild GuildBotsAuthorizeGuildsGet(ctx).Execute()

List guilds available for bot authorization

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
	resp, r, err := apiClient.GuildBotsAPI.GuildBotsAuthorizeGuildsGet(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildBotsAPI.GuildBotsAuthorizeGuildsGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildBotsAuthorizeGuildsGet`: []DtoGuild
	fmt.Fprintf(os.Stdout, "Response from `GuildBotsAPI.GuildBotsAuthorizeGuildsGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiGuildBotsAuthorizeGuildsGetRequest struct via the builder pattern


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


## GuildGuildIdBotsBotIdDelete

> GuildGuildIdBotsBotIdDelete(ctx, guildId, botId).Execute()

Remove a bot from a guild

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
	botId := int32(56) // int32 | Bot user id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.GuildBotsAPI.GuildGuildIdBotsBotIdDelete(context.Background(), guildId, botId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildBotsAPI.GuildGuildIdBotsBotIdDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild id | 
**botId** | **int32** | Bot user id | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdBotsBotIdDeleteRequest struct via the builder pattern


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


## GuildGuildIdBotsGet

> []GuildInstalledBotResponse GuildGuildIdBotsGet(ctx, guildId).Execute()

List installed guild bots

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
	resp, r, err := apiClient.GuildBotsAPI.GuildGuildIdBotsGet(context.Background(), guildId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildBotsAPI.GuildGuildIdBotsGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdBotsGet`: []GuildInstalledBotResponse
	fmt.Fprintf(os.Stdout, "Response from `GuildBotsAPI.GuildGuildIdBotsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild id | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdBotsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]GuildInstalledBotResponse**](GuildInstalledBotResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GuildGuildIdBotsPost

> GuildInstalledBotResponse GuildGuildIdBotsPost(ctx, guildId).Request(request).Execute()

Install a bot into a guild

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
	request := *openapiclient.NewGuildInstallBotRequest() // GuildInstallBotRequest | Install request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.GuildBotsAPI.GuildGuildIdBotsPost(context.Background(), guildId).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `GuildBotsAPI.GuildGuildIdBotsPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `GuildGuildIdBotsPost`: GuildInstalledBotResponse
	fmt.Fprintf(os.Stdout, "Response from `GuildBotsAPI.GuildGuildIdBotsPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild id | 

### Other Parameters

Other parameters are passed through a pointer to a apiGuildGuildIdBotsPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **request** | [**GuildInstallBotRequest**](GuildInstallBotRequest.md) | Install request | 

### Return type

[**GuildInstalledBotResponse**](GuildInstalledBotResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

