# MessageSendMessageRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Attachments** | Pointer to **[]int32** | IDs of attached files | [optional] 
**Content** | Pointer to **string** | Message content | [optional] 
**Mentions** | Pointer to **[]int32** | IDs of mentioned users | [optional] 

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


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


