# GuildChannelOrder

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **int32** | Channel ID. | [optional] 
**Position** | Pointer to **int32** | New channel position. | [optional] 

## Methods

### NewGuildChannelOrder

`func NewGuildChannelOrder() *GuildChannelOrder`

NewGuildChannelOrder instantiates a new GuildChannelOrder object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGuildChannelOrderWithDefaults

`func NewGuildChannelOrderWithDefaults() *GuildChannelOrder`

NewGuildChannelOrderWithDefaults instantiates a new GuildChannelOrder object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *GuildChannelOrder) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *GuildChannelOrder) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *GuildChannelOrder) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *GuildChannelOrder) HasId() bool`

HasId returns a boolean if a field has been set.

### GetPosition

`func (o *GuildChannelOrder) GetPosition() int32`

GetPosition returns the Position field if non-nil, zero value otherwise.

### GetPositionOk

`func (o *GuildChannelOrder) GetPositionOk() (*int32, bool)`

GetPositionOk returns a tuple with the Position field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPosition

`func (o *GuildChannelOrder) SetPosition(v int32)`

SetPosition sets Position field to given value.

### HasPosition

`func (o *GuildChannelOrder) HasPosition() bool`

HasPosition returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


