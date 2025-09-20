# GuildPatchGuildChannelRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | Channel name. | [optional] 
**ParentId** | Pointer to **int32** | Parent channel ID. A Parent channel can only be a category channel. | [optional] 
**Private** | Pointer to **bool** | Whether the channel is private. Private channels can only be seen by users with roles assigned to this channel. | [optional] [default to false]
**Topic** | Pointer to **string** | Channel topic. | [optional] 

## Methods

### NewGuildPatchGuildChannelRequest

`func NewGuildPatchGuildChannelRequest() *GuildPatchGuildChannelRequest`

NewGuildPatchGuildChannelRequest instantiates a new GuildPatchGuildChannelRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGuildPatchGuildChannelRequestWithDefaults

`func NewGuildPatchGuildChannelRequestWithDefaults() *GuildPatchGuildChannelRequest`

NewGuildPatchGuildChannelRequestWithDefaults instantiates a new GuildPatchGuildChannelRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *GuildPatchGuildChannelRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *GuildPatchGuildChannelRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *GuildPatchGuildChannelRequest) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *GuildPatchGuildChannelRequest) HasName() bool`

HasName returns a boolean if a field has been set.

### GetParentId

`func (o *GuildPatchGuildChannelRequest) GetParentId() int32`

GetParentId returns the ParentId field if non-nil, zero value otherwise.

### GetParentIdOk

`func (o *GuildPatchGuildChannelRequest) GetParentIdOk() (*int32, bool)`

GetParentIdOk returns a tuple with the ParentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentId

`func (o *GuildPatchGuildChannelRequest) SetParentId(v int32)`

SetParentId sets ParentId field to given value.

### HasParentId

`func (o *GuildPatchGuildChannelRequest) HasParentId() bool`

HasParentId returns a boolean if a field has been set.

### GetPrivate

`func (o *GuildPatchGuildChannelRequest) GetPrivate() bool`

GetPrivate returns the Private field if non-nil, zero value otherwise.

### GetPrivateOk

`func (o *GuildPatchGuildChannelRequest) GetPrivateOk() (*bool, bool)`

GetPrivateOk returns a tuple with the Private field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrivate

`func (o *GuildPatchGuildChannelRequest) SetPrivate(v bool)`

SetPrivate sets Private field to given value.

### HasPrivate

`func (o *GuildPatchGuildChannelRequest) HasPrivate() bool`

HasPrivate returns a boolean if a field has been set.

### GetTopic

`func (o *GuildPatchGuildChannelRequest) GetTopic() string`

GetTopic returns the Topic field if non-nil, zero value otherwise.

### GetTopicOk

`func (o *GuildPatchGuildChannelRequest) GetTopicOk() (*string, bool)`

GetTopicOk returns a tuple with the Topic field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTopic

`func (o *GuildPatchGuildChannelRequest) SetTopic(v string)`

SetTopic sets Topic field to given value.

### HasTopic

`func (o *GuildPatchGuildChannelRequest) HasTopic() bool`

HasTopic returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


