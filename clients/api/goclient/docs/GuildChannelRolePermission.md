# GuildChannelRolePermission

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Accept** | Pointer to **int32** | Allowed permission bits mask | [optional] 
**Deny** | Pointer to **int32** | Denied permission bits mask | [optional] 
**RoleId** | Pointer to **int32** | Role ID | [optional] 

## Methods

### NewGuildChannelRolePermission

`func NewGuildChannelRolePermission() *GuildChannelRolePermission`

NewGuildChannelRolePermission instantiates a new GuildChannelRolePermission object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGuildChannelRolePermissionWithDefaults

`func NewGuildChannelRolePermissionWithDefaults() *GuildChannelRolePermission`

NewGuildChannelRolePermissionWithDefaults instantiates a new GuildChannelRolePermission object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAccept

`func (o *GuildChannelRolePermission) GetAccept() int32`

GetAccept returns the Accept field if non-nil, zero value otherwise.

### GetAcceptOk

`func (o *GuildChannelRolePermission) GetAcceptOk() (*int32, bool)`

GetAcceptOk returns a tuple with the Accept field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccept

`func (o *GuildChannelRolePermission) SetAccept(v int32)`

SetAccept sets Accept field to given value.

### HasAccept

`func (o *GuildChannelRolePermission) HasAccept() bool`

HasAccept returns a boolean if a field has been set.

### GetDeny

`func (o *GuildChannelRolePermission) GetDeny() int32`

GetDeny returns the Deny field if non-nil, zero value otherwise.

### GetDenyOk

`func (o *GuildChannelRolePermission) GetDenyOk() (*int32, bool)`

GetDenyOk returns a tuple with the Deny field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeny

`func (o *GuildChannelRolePermission) SetDeny(v int32)`

SetDeny sets Deny field to given value.

### HasDeny

`func (o *GuildChannelRolePermission) HasDeny() bool`

HasDeny returns a boolean if a field has been set.

### GetRoleId

`func (o *GuildChannelRolePermission) GetRoleId() int32`

GetRoleId returns the RoleId field if non-nil, zero value otherwise.

### GetRoleIdOk

`func (o *GuildChannelRolePermission) GetRoleIdOk() (*int32, bool)`

GetRoleIdOk returns a tuple with the RoleId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoleId

`func (o *GuildChannelRolePermission) SetRoleId(v int32)`

SetRoleId sets RoleId field to given value.

### HasRoleId

`func (o *GuildChannelRolePermission) HasRoleId() bool`

HasRoleId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


