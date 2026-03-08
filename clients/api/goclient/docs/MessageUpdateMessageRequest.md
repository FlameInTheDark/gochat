# MessageUpdateMessageRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Content** | Pointer to **string** | Message content | [optional] 
**Embeds** | Pointer to [**[]EmbedEmbed**](EmbedEmbed.md) | Full replacement for the manual embed array. Generated embeds are managed by the embedder service. | [optional] 
**Flags** | Pointer to **int32** | Message flags bitmask. Use 4 to suppress URL embed generation and clear generated embeds. | [optional] 

## Methods

### NewMessageUpdateMessageRequest

`func NewMessageUpdateMessageRequest() *MessageUpdateMessageRequest`

NewMessageUpdateMessageRequest instantiates a new MessageUpdateMessageRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMessageUpdateMessageRequestWithDefaults

`func NewMessageUpdateMessageRequestWithDefaults() *MessageUpdateMessageRequest`

NewMessageUpdateMessageRequestWithDefaults instantiates a new MessageUpdateMessageRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContent

`func (o *MessageUpdateMessageRequest) GetContent() string`

GetContent returns the Content field if non-nil, zero value otherwise.

### GetContentOk

`func (o *MessageUpdateMessageRequest) GetContentOk() (*string, bool)`

GetContentOk returns a tuple with the Content field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContent

`func (o *MessageUpdateMessageRequest) SetContent(v string)`

SetContent sets Content field to given value.

### HasContent

`func (o *MessageUpdateMessageRequest) HasContent() bool`

HasContent returns a boolean if a field has been set.

### GetEmbeds

`func (o *MessageUpdateMessageRequest) GetEmbeds() []EmbedEmbed`

GetEmbeds returns the Embeds field if non-nil, zero value otherwise.

### GetEmbedsOk

`func (o *MessageUpdateMessageRequest) GetEmbedsOk() (*[]EmbedEmbed, bool)`

GetEmbedsOk returns a tuple with the Embeds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmbeds

`func (o *MessageUpdateMessageRequest) SetEmbeds(v []EmbedEmbed)`

SetEmbeds sets Embeds field to given value.

### HasEmbeds

`func (o *MessageUpdateMessageRequest) HasEmbeds() bool`

HasEmbeds returns a boolean if a field has been set.

### GetFlags

`func (o *MessageUpdateMessageRequest) GetFlags() int32`

GetFlags returns the Flags field if non-nil, zero value otherwise.

### GetFlagsOk

`func (o *MessageUpdateMessageRequest) GetFlagsOk() (*int32, bool)`

GetFlagsOk returns a tuple with the Flags field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlags

`func (o *MessageUpdateMessageRequest) SetFlags(v int32)`

SetFlags sets Flags field to given value.

### HasFlags

`func (o *MessageUpdateMessageRequest) HasFlags() bool`

HasFlags returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


