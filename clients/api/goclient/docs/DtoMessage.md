# DtoMessage

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Attachments** | Pointer to [**[]DtoAttachment**](DtoAttachment.md) |  | [optional] 
**Author** | Pointer to [**DtoUser**](DtoUser.md) |  | [optional] 
**ChannelId** | Pointer to **int32** | Channel id the message was sent to | [optional] 
**Content** | Pointer to **string** |  | [optional] 
**Id** | Pointer to **int32** | Message ID | [optional] 
**Type** | Pointer to **int32** |  | [optional] 
**UpdatedAt** | Pointer to **string** | Timestamp of the last message edit | [optional] 

## Methods

### NewDtoMessage

`func NewDtoMessage() *DtoMessage`

NewDtoMessage instantiates a new DtoMessage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDtoMessageWithDefaults

`func NewDtoMessageWithDefaults() *DtoMessage`

NewDtoMessageWithDefaults instantiates a new DtoMessage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAttachments

`func (o *DtoMessage) GetAttachments() []DtoAttachment`

GetAttachments returns the Attachments field if non-nil, zero value otherwise.

### GetAttachmentsOk

`func (o *DtoMessage) GetAttachmentsOk() (*[]DtoAttachment, bool)`

GetAttachmentsOk returns a tuple with the Attachments field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttachments

`func (o *DtoMessage) SetAttachments(v []DtoAttachment)`

SetAttachments sets Attachments field to given value.

### HasAttachments

`func (o *DtoMessage) HasAttachments() bool`

HasAttachments returns a boolean if a field has been set.

### GetAuthor

`func (o *DtoMessage) GetAuthor() DtoUser`

GetAuthor returns the Author field if non-nil, zero value otherwise.

### GetAuthorOk

`func (o *DtoMessage) GetAuthorOk() (*DtoUser, bool)`

GetAuthorOk returns a tuple with the Author field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthor

`func (o *DtoMessage) SetAuthor(v DtoUser)`

SetAuthor sets Author field to given value.

### HasAuthor

`func (o *DtoMessage) HasAuthor() bool`

HasAuthor returns a boolean if a field has been set.

### GetChannelId

`func (o *DtoMessage) GetChannelId() int32`

GetChannelId returns the ChannelId field if non-nil, zero value otherwise.

### GetChannelIdOk

`func (o *DtoMessage) GetChannelIdOk() (*int32, bool)`

GetChannelIdOk returns a tuple with the ChannelId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChannelId

`func (o *DtoMessage) SetChannelId(v int32)`

SetChannelId sets ChannelId field to given value.

### HasChannelId

`func (o *DtoMessage) HasChannelId() bool`

HasChannelId returns a boolean if a field has been set.

### GetContent

`func (o *DtoMessage) GetContent() string`

GetContent returns the Content field if non-nil, zero value otherwise.

### GetContentOk

`func (o *DtoMessage) GetContentOk() (*string, bool)`

GetContentOk returns a tuple with the Content field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContent

`func (o *DtoMessage) SetContent(v string)`

SetContent sets Content field to given value.

### HasContent

`func (o *DtoMessage) HasContent() bool`

HasContent returns a boolean if a field has been set.

### GetId

`func (o *DtoMessage) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DtoMessage) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DtoMessage) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *DtoMessage) HasId() bool`

HasId returns a boolean if a field has been set.

### GetType

`func (o *DtoMessage) GetType() int32`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *DtoMessage) GetTypeOk() (*int32, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *DtoMessage) SetType(v int32)`

SetType sets Type field to given value.

### HasType

`func (o *DtoMessage) HasType() bool`

HasType returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *DtoMessage) GetUpdatedAt() string`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *DtoMessage) GetUpdatedAtOk() (*string, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *DtoMessage) SetUpdatedAt(v string)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *DtoMessage) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


