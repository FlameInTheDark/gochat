# GuildCreateGuildChannelCategoryRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | Category channel name | [optional] 
**Position** | Pointer to **int32** | Channel position in the list. Should be set as the last position in the channel list, or it will be one of the first in the list. | [optional] [default to 0]
**Private** | Pointer to **bool** | Whether the category channel is private. Private channels can only be seen by users with roles assigned to this channel. | [optional] [default to false]

## Methods

### NewGuildCreateGuildChannelCategoryRequest

`func NewGuildCreateGuildChannelCategoryRequest() *GuildCreateGuildChannelCategoryRequest`

NewGuildCreateGuildChannelCategoryRequest instantiates a new GuildCreateGuildChannelCategoryRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGuildCreateGuildChannelCategoryRequestWithDefaults

`func NewGuildCreateGuildChannelCategoryRequestWithDefaults() *GuildCreateGuildChannelCategoryRequest`

NewGuildCreateGuildChannelCategoryRequestWithDefaults instantiates a new GuildCreateGuildChannelCategoryRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *GuildCreateGuildChannelCategoryRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *GuildCreateGuildChannelCategoryRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *GuildCreateGuildChannelCategoryRequest) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *GuildCreateGuildChannelCategoryRequest) HasName() bool`

HasName returns a boolean if a field has been set.

### GetPosition

`func (o *GuildCreateGuildChannelCategoryRequest) GetPosition() int32`

GetPosition returns the Position field if non-nil, zero value otherwise.

### GetPositionOk

`func (o *GuildCreateGuildChannelCategoryRequest) GetPositionOk() (*int32, bool)`

GetPositionOk returns a tuple with the Position field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPosition

`func (o *GuildCreateGuildChannelCategoryRequest) SetPosition(v int32)`

SetPosition sets Position field to given value.

### HasPosition

`func (o *GuildCreateGuildChannelCategoryRequest) HasPosition() bool`

HasPosition returns a boolean if a field has been set.

### GetPrivate

`func (o *GuildCreateGuildChannelCategoryRequest) GetPrivate() bool`

GetPrivate returns the Private field if non-nil, zero value otherwise.

### GetPrivateOk

`func (o *GuildCreateGuildChannelCategoryRequest) GetPrivateOk() (*bool, bool)`

GetPrivateOk returns a tuple with the Private field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrivate

`func (o *GuildCreateGuildChannelCategoryRequest) SetPrivate(v bool)`

SetPrivate sets Private field to given value.

### HasPrivate

`func (o *GuildCreateGuildChannelCategoryRequest) HasPrivate() bool`

HasPrivate returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


