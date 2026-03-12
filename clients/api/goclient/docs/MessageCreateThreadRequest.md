# MessageCreateThreadRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Attachments** | Pointer to **[]int32** | IDs of attached files uploaded to the parent channel before thread creation. | [optional] 
**Content** | Pointer to **string** | First thread message content. | [optional] 
**Embeds** | Pointer to [**[]EmbedEmbed**](EmbedEmbed.md) | Manual embeds for the first thread message. | [optional] 
**Mentions** | Pointer to **[]int32** | IDs of mentioned users. | [optional] 
**Name** | Pointer to **string** | Optional explicit thread name. | [optional] 
**Nonce** | Pointer to **string** | Optional client correlation token for the starter message event. | [optional] 

## Methods

### NewMessageCreateThreadRequest

`func NewMessageCreateThreadRequest() *MessageCreateThreadRequest`

NewMessageCreateThreadRequest instantiates a new MessageCreateThreadRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMessageCreateThreadRequestWithDefaults

`func NewMessageCreateThreadRequestWithDefaults() *MessageCreateThreadRequest`

NewMessageCreateThreadRequestWithDefaults instantiates a new MessageCreateThreadRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAttachments

`func (o *MessageCreateThreadRequest) GetAttachments() []int32`

GetAttachments returns the Attachments field if non-nil, zero value otherwise.

### GetAttachmentsOk

`func (o *MessageCreateThreadRequest) GetAttachmentsOk() (*[]int32, bool)`

GetAttachmentsOk returns a tuple with the Attachments field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttachments

`func (o *MessageCreateThreadRequest) SetAttachments(v []int32)`

SetAttachments sets Attachments field to given value.

### HasAttachments

`func (o *MessageCreateThreadRequest) HasAttachments() bool`

HasAttachments returns a boolean if a field has been set.

### GetContent

`func (o *MessageCreateThreadRequest) GetContent() string`

GetContent returns the Content field if non-nil, zero value otherwise.

### GetContentOk

`func (o *MessageCreateThreadRequest) GetContentOk() (*string, bool)`

GetContentOk returns a tuple with the Content field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContent

`func (o *MessageCreateThreadRequest) SetContent(v string)`

SetContent sets Content field to given value.

### HasContent

`func (o *MessageCreateThreadRequest) HasContent() bool`

HasContent returns a boolean if a field has been set.

### GetEmbeds

`func (o *MessageCreateThreadRequest) GetEmbeds() []EmbedEmbed`

GetEmbeds returns the Embeds field if non-nil, zero value otherwise.

### GetEmbedsOk

`func (o *MessageCreateThreadRequest) GetEmbedsOk() (*[]EmbedEmbed, bool)`

GetEmbedsOk returns a tuple with the Embeds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmbeds

`func (o *MessageCreateThreadRequest) SetEmbeds(v []EmbedEmbed)`

SetEmbeds sets Embeds field to given value.

### HasEmbeds

`func (o *MessageCreateThreadRequest) HasEmbeds() bool`

HasEmbeds returns a boolean if a field has been set.

### GetMentions

`func (o *MessageCreateThreadRequest) GetMentions() []int32`

GetMentions returns the Mentions field if non-nil, zero value otherwise.

### GetMentionsOk

`func (o *MessageCreateThreadRequest) GetMentionsOk() (*[]int32, bool)`

GetMentionsOk returns a tuple with the Mentions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMentions

`func (o *MessageCreateThreadRequest) SetMentions(v []int32)`

SetMentions sets Mentions field to given value.

### HasMentions

`func (o *MessageCreateThreadRequest) HasMentions() bool`

HasMentions returns a boolean if a field has been set.

### GetName

`func (o *MessageCreateThreadRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *MessageCreateThreadRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *MessageCreateThreadRequest) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *MessageCreateThreadRequest) HasName() bool`

HasName returns a boolean if a field has been set.

### GetNonce

`func (o *MessageCreateThreadRequest) GetNonce() string`

GetNonce returns the Nonce field if non-nil, zero value otherwise.

### GetNonceOk

`func (o *MessageCreateThreadRequest) GetNonceOk() (*string, bool)`

GetNonceOk returns a tuple with the Nonce field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNonce

`func (o *MessageCreateThreadRequest) SetNonce(v string)`

SetNonce sets Nonce field to given value.

### HasNonce

`func (o *MessageCreateThreadRequest) HasNonce() bool`

HasNonce returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


