# GuildRoleOrder

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** | Role ID. | [optional] 
**Position** | Pointer to **int32** | New role position. | [optional] 

## Methods

### NewGuildRoleOrder

`func NewGuildRoleOrder() *GuildRoleOrder`

NewGuildRoleOrder instantiates a new GuildRoleOrder object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGuildRoleOrderWithDefaults

`func NewGuildRoleOrderWithDefaults() *GuildRoleOrder`

NewGuildRoleOrderWithDefaults instantiates a new GuildRoleOrder object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *GuildRoleOrder) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *GuildRoleOrder) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *GuildRoleOrder) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *GuildRoleOrder) HasId() bool`

HasId returns a boolean if a field has been set.

### GetPosition

`func (o *GuildRoleOrder) GetPosition() int32`

GetPosition returns the Position field if non-nil, zero value otherwise.

### GetPositionOk

`func (o *GuildRoleOrder) GetPositionOk() (*int32, bool)`

GetPositionOk returns a tuple with the Position field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPosition

`func (o *GuildRoleOrder) SetPosition(v int32)`

SetPosition sets Position field to given value.

### HasPosition

`func (o *GuildRoleOrder) HasPosition() bool`

HasPosition returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


