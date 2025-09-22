# GuildCreateGuildRoleRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Color** | Pointer to **int32** | RGB int value | [optional] 
**Name** | Pointer to **string** | Role name | [optional] 
**Permissions** | Pointer to **int32** | Permissions bitset | [optional] [default to 0]

## Methods

### NewGuildCreateGuildRoleRequest

`func NewGuildCreateGuildRoleRequest() *GuildCreateGuildRoleRequest`

NewGuildCreateGuildRoleRequest instantiates a new GuildCreateGuildRoleRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGuildCreateGuildRoleRequestWithDefaults

`func NewGuildCreateGuildRoleRequestWithDefaults() *GuildCreateGuildRoleRequest`

NewGuildCreateGuildRoleRequestWithDefaults instantiates a new GuildCreateGuildRoleRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetColor

`func (o *GuildCreateGuildRoleRequest) GetColor() int32`

GetColor returns the Color field if non-nil, zero value otherwise.

### GetColorOk

`func (o *GuildCreateGuildRoleRequest) GetColorOk() (*int32, bool)`

GetColorOk returns a tuple with the Color field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetColor

`func (o *GuildCreateGuildRoleRequest) SetColor(v int32)`

SetColor sets Color field to given value.

### HasColor

`func (o *GuildCreateGuildRoleRequest) HasColor() bool`

HasColor returns a boolean if a field has been set.

### GetName

`func (o *GuildCreateGuildRoleRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *GuildCreateGuildRoleRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *GuildCreateGuildRoleRequest) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *GuildCreateGuildRoleRequest) HasName() bool`

HasName returns a boolean if a field has been set.

### GetPermissions

`func (o *GuildCreateGuildRoleRequest) GetPermissions() int32`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *GuildCreateGuildRoleRequest) GetPermissionsOk() (*int32, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *GuildCreateGuildRoleRequest) SetPermissions(v int32)`

SetPermissions sets Permissions field to given value.

### HasPermissions

`func (o *GuildCreateGuildRoleRequest) HasPermissions() bool`

HasPermissions returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


