# MessageUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Content** | Pointer to **string** |  | [optional] 
**Embeds** | Pointer to [**[]EmbedEmbed**](EmbedEmbed.md) |  | [optional] 
**Flags** | Pointer to **int32** |  | [optional] 

## Methods

### NewMessageUpdateRequest

`func NewMessageUpdateRequest() *MessageUpdateRequest`

NewMessageUpdateRequest instantiates a new MessageUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMessageUpdateRequestWithDefaults

`func NewMessageUpdateRequestWithDefaults() *MessageUpdateRequest`

NewMessageUpdateRequestWithDefaults instantiates a new MessageUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContent

`func (o *MessageUpdateRequest) GetContent() string`

GetContent returns the Content field if non-nil, zero value otherwise.

### GetContentOk

`func (o *MessageUpdateRequest) GetContentOk() (*string, bool)`

GetContentOk returns a tuple with the Content field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContent

`func (o *MessageUpdateRequest) SetContent(v string)`

SetContent sets Content field to given value.

### HasContent

`func (o *MessageUpdateRequest) HasContent() bool`

HasContent returns a boolean if a field has been set.

### GetEmbeds

`func (o *MessageUpdateRequest) GetEmbeds() []EmbedEmbed`

GetEmbeds returns the Embeds field if non-nil, zero value otherwise.

### GetEmbedsOk

`func (o *MessageUpdateRequest) GetEmbedsOk() (*[]EmbedEmbed, bool)`

GetEmbedsOk returns a tuple with the Embeds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmbeds

`func (o *MessageUpdateRequest) SetEmbeds(v []EmbedEmbed)`

SetEmbeds sets Embeds field to given value.

### HasEmbeds

`func (o *MessageUpdateRequest) HasEmbeds() bool`

HasEmbeds returns a boolean if a field has been set.

### GetFlags

`func (o *MessageUpdateRequest) GetFlags() int32`

GetFlags returns the Flags field if non-nil, zero value otherwise.

### GetFlagsOk

`func (o *MessageUpdateRequest) GetFlagsOk() (*int32, bool)`

GetFlagsOk returns a tuple with the Flags field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFlags

`func (o *MessageUpdateRequest) SetFlags(v int32)`

SetFlags sets Flags field to given value.

### HasFlags

`func (o *MessageUpdateRequest) HasFlags() bool`

HasFlags returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


