# MessageSendMessageRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Attachments** | Pointer to **[]int32** | IDs of attached files | [optional] 
**Content** | Pointer to **string** | Message content | [optional] 
**Embeds** | Pointer to [**[]EmbedEmbed**](EmbedEmbed.md) | Manual embeds supplied by the client. These are stored separately from generated URL embeds. | [optional] 
**EnforceNonce** | Pointer to **bool** | When true, deduplicates sends with the same nonce in the same channel for a short window. | [optional] 
**Mentions** | Pointer to **[]int32** | IDs of mentioned users | [optional] 
**Nonce** | Pointer to **string** | Optional client correlation token echoed back to the author. | [optional] 
**Reference** | Pointer to **int32** | Referenced message ID in the same channel. When set, the new message is stored as type 1 (Reply). | [optional] 

## Methods

### NewMessageSendMessageRequest

`func NewMessageSendMessageRequest() *MessageSendMessageRequest`

NewMessageSendMessageRequest instantiates a new MessageSendMessageRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMessageSendMessageRequestWithDefaults

`func NewMessageSendMessageRequestWithDefaults() *MessageSendMessageRequest`

NewMessageSendMessageRequestWithDefaults instantiates a new MessageSendMessageRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAttachments

`func (o *MessageSendMessageRequest) GetAttachments() []int32`

GetAttachments returns the Attachments field if non-nil, zero value otherwise.

### GetAttachmentsOk

`func (o *MessageSendMessageRequest) GetAttachmentsOk() (*[]int32, bool)`

GetAttachmentsOk returns a tuple with the Attachments field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttachments

`func (o *MessageSendMessageRequest) SetAttachments(v []int32)`

SetAttachments sets Attachments field to given value.

### HasAttachments

`func (o *MessageSendMessageRequest) HasAttachments() bool`

HasAttachments returns a boolean if a field has been set.

### GetContent

`func (o *MessageSendMessageRequest) GetContent() string`

GetContent returns the Content field if non-nil, zero value otherwise.

### GetContentOk

`func (o *MessageSendMessageRequest) GetContentOk() (*string, bool)`

GetContentOk returns a tuple with the Content field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContent

`func (o *MessageSendMessageRequest) SetContent(v string)`

SetContent sets Content field to given value.

### HasContent

`func (o *MessageSendMessageRequest) HasContent() bool`

HasContent returns a boolean if a field has been set.

### GetEmbeds

`func (o *MessageSendMessageRequest) GetEmbeds() []EmbedEmbed`

GetEmbeds returns the Embeds field if non-nil, zero value otherwise.

### GetEmbedsOk

`func (o *MessageSendMessageRequest) GetEmbedsOk() (*[]EmbedEmbed, bool)`

GetEmbedsOk returns a tuple with the Embeds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmbeds

`func (o *MessageSendMessageRequest) SetEmbeds(v []EmbedEmbed)`

SetEmbeds sets Embeds field to given value.

### HasEmbeds

`func (o *MessageSendMessageRequest) HasEmbeds() bool`

HasEmbeds returns a boolean if a field has been set.

### GetEnforceNonce

`func (o *MessageSendMessageRequest) GetEnforceNonce() bool`

GetEnforceNonce returns the EnforceNonce field if non-nil, zero value otherwise.

### GetEnforceNonceOk

`func (o *MessageSendMessageRequest) GetEnforceNonceOk() (*bool, bool)`

GetEnforceNonceOk returns a tuple with the EnforceNonce field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnforceNonce

`func (o *MessageSendMessageRequest) SetEnforceNonce(v bool)`

SetEnforceNonce sets EnforceNonce field to given value.

### HasEnforceNonce

`func (o *MessageSendMessageRequest) HasEnforceNonce() bool`

HasEnforceNonce returns a boolean if a field has been set.

### GetMentions

`func (o *MessageSendMessageRequest) GetMentions() []int32`

GetMentions returns the Mentions field if non-nil, zero value otherwise.

### GetMentionsOk

`func (o *MessageSendMessageRequest) GetMentionsOk() (*[]int32, bool)`

GetMentionsOk returns a tuple with the Mentions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMentions

`func (o *MessageSendMessageRequest) SetMentions(v []int32)`

SetMentions sets Mentions field to given value.

### HasMentions

`func (o *MessageSendMessageRequest) HasMentions() bool`

HasMentions returns a boolean if a field has been set.

### GetNonce

`func (o *MessageSendMessageRequest) GetNonce() string`

GetNonce returns the Nonce field if non-nil, zero value otherwise.

### GetNonceOk

`func (o *MessageSendMessageRequest) GetNonceOk() (*string, bool)`

GetNonceOk returns a tuple with the Nonce field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNonce

`func (o *MessageSendMessageRequest) SetNonce(v string)`

SetNonce sets Nonce field to given value.

### HasNonce

`func (o *MessageSendMessageRequest) HasNonce() bool`

HasNonce returns a boolean if a field has been set.

### GetReference

`func (o *MessageSendMessageRequest) GetReference() int32`

GetReference returns the Reference field if non-nil, zero value otherwise.

### GetReferenceOk

`func (o *MessageSendMessageRequest) GetReferenceOk() (*int32, bool)`

GetReferenceOk returns a tuple with the Reference field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReference

`func (o *MessageSendMessageRequest) SetReference(v int32)`

SetReference sets Reference field to given value.

### HasReference

`func (o *MessageSendMessageRequest) HasReference() bool`

HasReference returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


