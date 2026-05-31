# \BotGuildAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**BotApiV1GuildGet**](BotGuildAPI.md#BotApiV1GuildGet) | **Get** /bot/api/v1/guild | List guilds where the bot is installed
[**BotApiV1GuildGuildIdChannelsGet**](BotGuildAPI.md#BotApiV1GuildGuildIdChannelsGet) | **Get** /bot/api/v1/guild/{guild_id}/channels | List channels visible to the bot in a guild



## BotApiV1GuildGet

> []GuildResponse BotApiV1GuildGet(ctx).Execute()

List guilds where the bot is installed

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
	resp, r, err := apiClient.BotGuildAPI.BotApiV1GuildGet(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `BotGuildAPI.BotApiV1GuildGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BotApiV1GuildGet`: []GuildResponse
	fmt.Fprintf(os.Stdout, "Response from `BotGuildAPI.BotApiV1GuildGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiBotApiV1GuildGetRequest struct via the builder pattern


### Return type

[**[]GuildResponse**](GuildResponse.md)

### Authorization

[BotToken](../README.md#BotToken)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BotApiV1GuildGuildIdChannelsGet

> []DtoChannel BotApiV1GuildGuildIdChannelsGet(ctx, guildId).Execute()

List channels visible to the bot in a guild

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
	resp, r, err := apiClient.BotGuildAPI.BotApiV1GuildGuildIdChannelsGet(context.Background(), guildId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `BotGuildAPI.BotApiV1GuildGuildIdChannelsGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BotApiV1GuildGuildIdChannelsGet`: []DtoChannel
	fmt.Fprintf(os.Stdout, "Response from `BotGuildAPI.BotApiV1GuildGuildIdChannelsGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**guildId** | **int32** | Guild id | 

### Other Parameters

Other parameters are passed through a pointer to a apiBotApiV1GuildGuildIdChannelsGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**[]DtoChannel**](DtoChannel.md)

### Authorization

[BotToken](../README.md#BotToken)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

