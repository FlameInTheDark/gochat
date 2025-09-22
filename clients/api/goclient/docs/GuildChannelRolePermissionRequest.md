# GuildChannelRolePermissionRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Accept** | Pointer to **int32** | Allowed permission bits mask | [optional] 
**Deny** | Pointer to **int32** | Denied permission bits mask | [optional] 

## Methods

### NewGuildChannelRolePermissionRequest

`func NewGuildChannelRolePermissionRequest() *GuildChannelRolePermissionRequest`

NewGuildChannelRolePermissionRequest instantiates a new GuildChannelRolePermissionRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGuildChannelRolePermissionRequestWithDefaults

`func NewGuildChannelRolePermissionRequestWithDefaults() *GuildChannelRolePermissionRequest`

NewGuildChannelRolePermissionRequestWithDefaults instantiates a new GuildChannelRolePermissionRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAccept

`func (o *GuildChannelRolePermissionRequest) GetAccept() int32`

GetAccept returns the Accept field if non-nil, zero value otherwise.

### GetAcceptOk

`func (o *GuildChannelRolePermissionRequest) GetAcceptOk() (*int32, bool)`

GetAcceptOk returns a tuple with the Accept field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccept

`func (o *GuildChannelRolePermissionRequest) SetAccept(v int32)`

SetAccept sets Accept field to given value.

### HasAccept

`func (o *GuildChannelRolePermissionRequest) HasAccept() bool`

HasAccept returns a boolean if a field has been set.

### GetDeny

`func (o *GuildChannelRolePermissionRequest) GetDeny() int32`

GetDeny returns the Deny field if non-nil, zero value otherwise.

### GetDenyOk

`func (o *GuildChannelRolePermissionRequest) GetDenyOk() (*int32, bool)`

GetDenyOk returns a tuple with the Deny field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeny

`func (o *GuildChannelRolePermissionRequest) SetDeny(v int32)`

SetDeny sets Deny field to given value.

### HasDeny

`func (o *GuildChannelRolePermissionRequest) HasDeny() bool`

HasDeny returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


