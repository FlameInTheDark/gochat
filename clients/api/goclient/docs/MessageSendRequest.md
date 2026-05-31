# MessageSendRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Attachments** | Pointer to **[]int32** |  | [optional] 
**Content** | Pointer to **string** |  | [optional] 
**Embeds** | Pointer to [**[]EmbedEmbed**](EmbedEmbed.md) |  | [optional] 
**Reference** | Pointer to **int32** |  | [optional] 

## Methods

### NewMessageSendRequest

`func NewMessageSendRequest() *MessageSendRequest`

NewMessageSendRequest instantiates a new MessageSendRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMessageSendRequestWithDefaults

`func NewMessageSendRequestWithDefaults() *MessageSendRequest`

NewMessageSendRequestWithDefaults instantiates a new MessageSendRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAttachments

`func (o *MessageSendRequest) GetAttachments() []int32`

GetAttachments returns the Attachments field if non-nil, zero value otherwise.

### GetAttachmentsOk

`func (o *MessageSendRequest) GetAttachmentsOk() (*[]int32, bool)`

GetAttachmentsOk returns a tuple with the Attachments field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttachments

`func (o *MessageSendRequest) SetAttachments(v []int32)`

SetAttachments sets Attachments field to given value.

### HasAttachments

`func (o *MessageSendRequest) HasAttachments() bool`

HasAttachments returns a boolean if a field has been set.

### GetContent

`func (o *MessageSendRequest) GetContent() string`

GetContent returns the Content field if non-nil, zero value otherwise.

### GetContentOk

`func (o *MessageSendRequest) GetContentOk() (*string, bool)`

GetContentOk returns a tuple with the Content field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContent

`func (o *MessageSendRequest) SetContent(v string)`

SetContent sets Content field to given value.

### HasContent

`func (o *MessageSendRequest) HasContent() bool`

HasContent returns a boolean if a field has been set.

### GetEmbeds

`func (o *MessageSendRequest) GetEmbeds() []EmbedEmbed`

GetEmbeds returns the Embeds field if non-nil, zero value otherwise.

### GetEmbedsOk

`func (o *MessageSendRequest) GetEmbedsOk() (*[]EmbedEmbed, bool)`

GetEmbedsOk returns a tuple with the Embeds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmbeds

`func (o *MessageSendRequest) SetEmbeds(v []EmbedEmbed)`

SetEmbeds sets Embeds field to given value.

### HasEmbeds

`func (o *MessageSendRequest) HasEmbeds() bool`

HasEmbeds returns a boolean if a field has been set.

### GetReference

`func (o *MessageSendRequest) GetReference() int32`

GetReference returns the Reference field if non-nil, zero value otherwise.

### GetReferenceOk

`func (o *MessageSendRequest) GetReferenceOk() (*int32, bool)`

GetReferenceOk returns a tuple with the Reference field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReference

`func (o *MessageSendRequest) SetReference(v int32)`

SetReference sets Reference field to given value.

### HasReference

`func (o *MessageSendRequest) HasReference() bool`

HasReference returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


