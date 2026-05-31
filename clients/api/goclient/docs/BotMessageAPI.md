# \BotMessageAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**BotApiV1MessageChannelChannelIdGet**](BotMessageAPI.md#BotApiV1MessageChannelChannelIdGet) | **Get** /bot/api/v1/message/channel/{channel_id} | List messages visible to the bot
[**BotApiV1MessageChannelChannelIdMessageIdAckPost**](BotMessageAPI.md#BotApiV1MessageChannelChannelIdMessageIdAckPost) | **Post** /bot/api/v1/message/channel/{channel_id}/{message_id}/ack | Mark a channel read as the bot
[**BotApiV1MessageChannelChannelIdMessageIdDelete**](BotMessageAPI.md#BotApiV1MessageChannelChannelIdMessageIdDelete) | **Delete** /bot/api/v1/message/channel/{channel_id}/{message_id} | Delete a message as the bot
[**BotApiV1MessageChannelChannelIdMessageIdPatch**](BotMessageAPI.md#BotApiV1MessageChannelChannelIdMessageIdPatch) | **Patch** /bot/api/v1/message/channel/{channel_id}/{message_id} | Edit a message as the bot
[**BotApiV1MessageChannelChannelIdMessageIdReactionsReactionNameDelete**](BotMessageAPI.md#BotApiV1MessageChannelChannelIdMessageIdReactionsReactionNameDelete) | **Delete** /bot/api/v1/message/channel/{channel_id}/{message_id}/reactions/{reaction_name} | Remove the bot reaction from a message
[**BotApiV1MessageChannelChannelIdMessageIdReactionsReactionNameGet**](BotMessageAPI.md#BotApiV1MessageChannelChannelIdMessageIdReactionsReactionNameGet) | **Get** /bot/api/v1/message/channel/{channel_id}/{message_id}/reactions/{reaction_name} | List users who reacted with a specific reaction
[**BotApiV1MessageChannelChannelIdMessageIdReactionsReactionNamePut**](BotMessageAPI.md#BotApiV1MessageChannelChannelIdMessageIdReactionsReactionNamePut) | **Put** /bot/api/v1/message/channel/{channel_id}/{message_id}/reactions/{reaction_name} | Add the bot reaction to a message
[**BotApiV1MessageChannelChannelIdPost**](BotMessageAPI.md#BotApiV1MessageChannelChannelIdPost) | **Post** /bot/api/v1/message/channel/{channel_id} | Send a message as the bot
[**BotApiV1MessageChannelChannelIdTypingPost**](BotMessageAPI.md#BotApiV1MessageChannelChannelIdTypingPost) | **Post** /bot/api/v1/message/channel/{channel_id}/typing | Send a typing indicator as the bot



## BotApiV1MessageChannelChannelIdGet

> []DtoMessage BotApiV1MessageChannelChannelIdGet(ctx, channelId).From(from).Limit(limit).Direction(direction).Execute()

List messages visible to the bot

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
	channelId := int32(56) // int32 | Channel id
	from := int32(56) // int32 | Message id cursor (optional)
	limit := int32(56) // int32 | Page size (optional)
	direction := "direction_example" // string | before, after, or around (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.BotMessageAPI.BotApiV1MessageChannelChannelIdGet(context.Background(), channelId).From(from).Limit(limit).Direction(direction).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `BotMessageAPI.BotApiV1MessageChannelChannelIdGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BotApiV1MessageChannelChannelIdGet`: []DtoMessage
	fmt.Fprintf(os.Stdout, "Response from `BotMessageAPI.BotApiV1MessageChannelChannelIdGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**channelId** | **int32** | Channel id | 

### Other Parameters

Other parameters are passed through a pointer to a apiBotApiV1MessageChannelChannelIdGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **from** | **int32** | Message id cursor | 
 **limit** | **int32** | Page size | 
 **direction** | **string** | before, after, or around | 

### Return type

[**[]DtoMessage**](DtoMessage.md)

### Authorization

[BotToken](../README.md#BotToken)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BotApiV1MessageChannelChannelIdMessageIdAckPost

> BotApiV1MessageChannelChannelIdMessageIdAckPost(ctx, channelId, messageId).Execute()

Mark a channel read as the bot

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
	channelId := int32(56) // int32 | Channel id
	messageId := int32(56) // int32 | Message id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.BotMessageAPI.BotApiV1MessageChannelChannelIdMessageIdAckPost(context.Background(), channelId, messageId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `BotMessageAPI.BotApiV1MessageChannelChannelIdMessageIdAckPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**channelId** | **int32** | Channel id | 
**messageId** | **int32** | Message id | 

### Other Parameters

Other parameters are passed through a pointer to a apiBotApiV1MessageChannelChannelIdMessageIdAckPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

 (empty response body)

### Authorization

[BotToken](../README.md#BotToken)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BotApiV1MessageChannelChannelIdMessageIdDelete

> BotApiV1MessageChannelChannelIdMessageIdDelete(ctx, channelId, messageId).Execute()

Delete a message as the bot

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
	channelId := int32(56) // int32 | Channel id
	messageId := int32(56) // int32 | Message id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.BotMessageAPI.BotApiV1MessageChannelChannelIdMessageIdDelete(context.Background(), channelId, messageId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `BotMessageAPI.BotApiV1MessageChannelChannelIdMessageIdDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**channelId** | **int32** | Channel id | 
**messageId** | **int32** | Message id | 

### Other Parameters

Other parameters are passed through a pointer to a apiBotApiV1MessageChannelChannelIdMessageIdDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



### Return type

 (empty response body)

### Authorization

[BotToken](../README.md#BotToken)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BotApiV1MessageChannelChannelIdMessageIdPatch

> DtoMessage BotApiV1MessageChannelChannelIdMessageIdPatch(ctx, channelId, messageId).Request(request).Execute()

Edit a message as the bot

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
	channelId := int32(56) // int32 | Channel id
	messageId := int32(56) // int32 | Message id
	request := *openapiclient.NewMessageUpdateRequest() // MessageUpdateRequest | Message update

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.BotMessageAPI.BotApiV1MessageChannelChannelIdMessageIdPatch(context.Background(), channelId, messageId).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `BotMessageAPI.BotApiV1MessageChannelChannelIdMessageIdPatch``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BotApiV1MessageChannelChannelIdMessageIdPatch`: DtoMessage
	fmt.Fprintf(os.Stdout, "Response from `BotMessageAPI.BotApiV1MessageChannelChannelIdMessageIdPatch`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**channelId** | **int32** | Channel id | 
**messageId** | **int32** | Message id | 

### Other Parameters

Other parameters are passed through a pointer to a apiBotApiV1MessageChannelChannelIdMessageIdPatchRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **request** | [**MessageUpdateRequest**](MessageUpdateRequest.md) | Message update | 

### Return type

[**DtoMessage**](DtoMessage.md)

### Authorization

[BotToken](../README.md#BotToken)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BotApiV1MessageChannelChannelIdMessageIdReactionsReactionNameDelete

> BotApiV1MessageChannelChannelIdMessageIdReactionsReactionNameDelete(ctx, channelId, messageId, reactionName).Execute()

Remove the bot reaction from a message

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
	channelId := int32(56) // int32 | Channel id
	messageId := int32(56) // int32 | Message id
	reactionName := "reactionName_example" // string | Reaction name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.BotMessageAPI.BotApiV1MessageChannelChannelIdMessageIdReactionsReactionNameDelete(context.Background(), channelId, messageId, reactionName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `BotMessageAPI.BotApiV1MessageChannelChannelIdMessageIdReactionsReactionNameDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**channelId** | **int32** | Channel id | 
**messageId** | **int32** | Message id | 
**reactionName** | **string** | Reaction name | 

### Other Parameters

Other parameters are passed through a pointer to a apiBotApiV1MessageChannelChannelIdMessageIdReactionsReactionNameDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

 (empty response body)

### Authorization

[BotToken](../README.md#BotToken)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BotApiV1MessageChannelChannelIdMessageIdReactionsReactionNameGet

> DtoMessageReactionUsersPage BotApiV1MessageChannelChannelIdMessageIdReactionsReactionNameGet(ctx, channelId, messageId, reactionName).After(after).Limit(limit).Execute()

List users who reacted with a specific reaction

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
	channelId := int32(56) // int32 | Channel id
	messageId := int32(56) // int32 | Message id
	reactionName := "reactionName_example" // string | Reaction name
	after := int32(56) // int32 | Reaction ID cursor (optional)
	limit := int32(56) // int32 | Page size (optional)

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.BotMessageAPI.BotApiV1MessageChannelChannelIdMessageIdReactionsReactionNameGet(context.Background(), channelId, messageId, reactionName).After(after).Limit(limit).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `BotMessageAPI.BotApiV1MessageChannelChannelIdMessageIdReactionsReactionNameGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BotApiV1MessageChannelChannelIdMessageIdReactionsReactionNameGet`: DtoMessageReactionUsersPage
	fmt.Fprintf(os.Stdout, "Response from `BotMessageAPI.BotApiV1MessageChannelChannelIdMessageIdReactionsReactionNameGet`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**channelId** | **int32** | Channel id | 
**messageId** | **int32** | Message id | 
**reactionName** | **string** | Reaction name | 

### Other Parameters

Other parameters are passed through a pointer to a apiBotApiV1MessageChannelChannelIdMessageIdReactionsReactionNameGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------



 **after** | **int32** | Reaction ID cursor | 
 **limit** | **int32** | Page size | 

### Return type

[**DtoMessageReactionUsersPage**](DtoMessageReactionUsersPage.md)

### Authorization

[BotToken](../README.md#BotToken)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BotApiV1MessageChannelChannelIdMessageIdReactionsReactionNamePut

> DtoMessageReaction BotApiV1MessageChannelChannelIdMessageIdReactionsReactionNamePut(ctx, channelId, messageId, reactionName).Execute()

Add the bot reaction to a message

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
	channelId := int32(56) // int32 | Channel id
	messageId := int32(56) // int32 | Message id
	reactionName := "reactionName_example" // string | Reaction name

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.BotMessageAPI.BotApiV1MessageChannelChannelIdMessageIdReactionsReactionNamePut(context.Background(), channelId, messageId, reactionName).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `BotMessageAPI.BotApiV1MessageChannelChannelIdMessageIdReactionsReactionNamePut``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BotApiV1MessageChannelChannelIdMessageIdReactionsReactionNamePut`: DtoMessageReaction
	fmt.Fprintf(os.Stdout, "Response from `BotMessageAPI.BotApiV1MessageChannelChannelIdMessageIdReactionsReactionNamePut`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**channelId** | **int32** | Channel id | 
**messageId** | **int32** | Message id | 
**reactionName** | **string** | Reaction name | 

### Other Parameters

Other parameters are passed through a pointer to a apiBotApiV1MessageChannelChannelIdMessageIdReactionsReactionNamePutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------




### Return type

[**DtoMessageReaction**](DtoMessageReaction.md)

### Authorization

[BotToken](../README.md#BotToken)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BotApiV1MessageChannelChannelIdPost

> DtoMessage BotApiV1MessageChannelChannelIdPost(ctx, channelId).Request(request).Execute()

Send a message as the bot

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
	channelId := int32(56) // int32 | Channel id
	request := *openapiclient.NewMessageSendRequest() // MessageSendRequest | Message body

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.BotMessageAPI.BotApiV1MessageChannelChannelIdPost(context.Background(), channelId).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `BotMessageAPI.BotApiV1MessageChannelChannelIdPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `BotApiV1MessageChannelChannelIdPost`: DtoMessage
	fmt.Fprintf(os.Stdout, "Response from `BotMessageAPI.BotApiV1MessageChannelChannelIdPost`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**channelId** | **int32** | Channel id | 

### Other Parameters

Other parameters are passed through a pointer to a apiBotApiV1MessageChannelChannelIdPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **request** | [**MessageSendRequest**](MessageSendRequest.md) | Message body | 

### Return type

[**DtoMessage**](DtoMessage.md)

### Authorization

[BotToken](../README.md#BotToken)

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## BotApiV1MessageChannelChannelIdTypingPost

> BotApiV1MessageChannelChannelIdTypingPost(ctx, channelId).Execute()

Send a typing indicator as the bot

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
	channelId := int32(56) // int32 | Channel id

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	r, err := apiClient.BotMessageAPI.BotApiV1MessageChannelChannelIdTypingPost(context.Background(), channelId).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `BotMessageAPI.BotApiV1MessageChannelChannelIdTypingPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**channelId** | **int32** | Channel id | 

### Other Parameters

Other parameters are passed through a pointer to a apiBotApiV1MessageChannelChannelIdTypingPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

 (empty response body)

### Authorization

[BotToken](../README.md#BotToken)

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

