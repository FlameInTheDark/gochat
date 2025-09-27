# GuildCreateGuildChannelRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | Channel name | [optional] 
**ParentId** | Pointer to **int32** | Parent channel ID. A Parent channel can only be a category channel. | [optional] 
**Position** | Pointer to **int32** | Channel position in the list. Should be set as the last position in the channel list, or it will be one of the first in the list. | [optional] [default to 0]
**Private** | Pointer to **bool** | Whether the channel is private. Private channels can only be seen by users with roles assigned to this channel. | [optional] [default to false]
**Type** | Pointer to **int32** | Channel type | [optional] 

## Methods

### NewGuildCreateGuildChannelRequest

`func NewGuildCreateGuildChannelRequest() *GuildCreateGuildChannelRequest`

NewGuildCreateGuildChannelRequest instantiates a new GuildCreateGuildChannelRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGuildCreateGuildChannelRequestWithDefaults

`func NewGuildCreateGuildChannelRequestWithDefaults() *GuildCreateGuildChannelRequest`

NewGuildCreateGuildChannelRequestWithDefaults instantiates a new GuildCreateGuildChannelRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *GuildCreateGuildChannelRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *GuildCreateGuildChannelRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *GuildCreateGuildChannelRequest) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *GuildCreateGuildChannelRequest) HasName() bool`

HasName returns a boolean if a field has been set.

### GetParentId

`func (o *GuildCreateGuildChannelRequest) GetParentId() int32`

GetParentId returns the ParentId field if non-nil, zero value otherwise.

### GetParentIdOk

`func (o *GuildCreateGuildChannelRequest) GetParentIdOk() (*int32, bool)`

GetParentIdOk returns a tuple with the ParentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentId

`func (o *GuildCreateGuildChannelRequest) SetParentId(v int32)`

SetParentId sets ParentId field to given value.

### HasParentId

`func (o *GuildCreateGuildChannelRequest) HasParentId() bool`

HasParentId returns a boolean if a field has been set.

### GetPosition

`func (o *GuildCreateGuildChannelRequest) GetPosition() int32`

GetPosition returns the Position field if non-nil, zero value otherwise.

### GetPositionOk

`func (o *GuildCreateGuildChannelRequest) GetPositionOk() (*int32, bool)`

GetPositionOk returns a tuple with the Position field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPosition

`func (o *GuildCreateGuildChannelRequest) SetPosition(v int32)`

SetPosition sets Position field to given value.

### HasPosition

`func (o *GuildCreateGuildChannelRequest) HasPosition() bool`

HasPosition returns a boolean if a field has been set.

### GetPrivate

`func (o *GuildCreateGuildChannelRequest) GetPrivate() bool`

GetPrivate returns the Private field if non-nil, zero value otherwise.

### GetPrivateOk

`func (o *GuildCreateGuildChannelRequest) GetPrivateOk() (*bool, bool)`

GetPrivateOk returns a tuple with the Private field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrivate

`func (o *GuildCreateGuildChannelRequest) SetPrivate(v bool)`

SetPrivate sets Private field to given value.

### HasPrivate

`func (o *GuildCreateGuildChannelRequest) HasPrivate() bool`

HasPrivate returns a boolean if a field has been set.

### GetType

`func (o *GuildCreateGuildChannelRequest) GetType() int32`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *GuildCreateGuildChannelRequest) GetTypeOk() (*int32, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *GuildCreateGuildChannelRequest) SetType(v int32)`

SetType sets Type field to given value.

### HasType

`func (o *GuildCreateGuildChannelRequest) HasType() bool`

HasType returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


