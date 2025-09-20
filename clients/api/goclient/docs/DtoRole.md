# DtoRole

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Color** | Pointer to **int32** | Role color. Will change username color. Represent RGB color in one Integer value. | [optional] 
**GuildId** | Pointer to **int32** | Guild ID | [optional] 
**Id** | Pointer to **int32** | Role ID | [optional] 
**Name** | Pointer to **string** | Role name | [optional] 
**Permissions** | Pointer to **int32** | Role permissions. Check the permissions documentation for more info. | [optional] 

## Methods

### NewDtoRole

`func NewDtoRole() *DtoRole`

NewDtoRole instantiates a new DtoRole object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDtoRoleWithDefaults

`func NewDtoRoleWithDefaults() *DtoRole`

NewDtoRoleWithDefaults instantiates a new DtoRole object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetColor

`func (o *DtoRole) GetColor() int32`

GetColor returns the Color field if non-nil, zero value otherwise.

### GetColorOk

`func (o *DtoRole) GetColorOk() (*int32, bool)`

GetColorOk returns a tuple with the Color field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetColor

`func (o *DtoRole) SetColor(v int32)`

SetColor sets Color field to given value.

### HasColor

`func (o *DtoRole) HasColor() bool`

HasColor returns a boolean if a field has been set.

### GetGuildId

`func (o *DtoRole) GetGuildId() int32`

GetGuildId returns the GuildId field if non-nil, zero value otherwise.

### GetGuildIdOk

`func (o *DtoRole) GetGuildIdOk() (*int32, bool)`

GetGuildIdOk returns a tuple with the GuildId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGuildId

`func (o *DtoRole) SetGuildId(v int32)`

SetGuildId sets GuildId field to given value.

### HasGuildId

`func (o *DtoRole) HasGuildId() bool`

HasGuildId returns a boolean if a field has been set.

### GetId

`func (o *DtoRole) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DtoRole) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DtoRole) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *DtoRole) HasId() bool`

HasId returns a boolean if a field has been set.

### GetName

`func (o *DtoRole) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DtoRole) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DtoRole) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *DtoRole) HasName() bool`

HasName returns a boolean if a field has been set.

### GetPermissions

`func (o *DtoRole) GetPermissions() int32`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *DtoRole) GetPermissionsOk() (*int32, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *DtoRole) SetPermissions(v int32)`

SetPermissions sets Permissions field to given value.

### HasPermissions

`func (o *DtoRole) HasPermissions() bool`

HasPermissions returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


