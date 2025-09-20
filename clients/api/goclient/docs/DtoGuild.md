# DtoGuild

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Icon** | Pointer to **int32** | Icon ID | [optional] 
**Id** | Pointer to **int32** | Guild ID | [optional] 
**Name** | Pointer to **string** | Guild Name | [optional] 
**Owner** | Pointer to **int32** | Owner ID | [optional] 
**Permissions** | Pointer to **int32** | Default guild Permissions. Check the permissions documentation for more info. | [optional] [default to 7927905]
**Public** | Pointer to **bool** | Whether the guild is public | [optional] [default to false]

## Methods

### NewDtoGuild

`func NewDtoGuild() *DtoGuild`

NewDtoGuild instantiates a new DtoGuild object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDtoGuildWithDefaults

`func NewDtoGuildWithDefaults() *DtoGuild`

NewDtoGuildWithDefaults instantiates a new DtoGuild object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIcon

`func (o *DtoGuild) GetIcon() int32`

GetIcon returns the Icon field if non-nil, zero value otherwise.

### GetIconOk

`func (o *DtoGuild) GetIconOk() (*int32, bool)`

GetIconOk returns a tuple with the Icon field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIcon

`func (o *DtoGuild) SetIcon(v int32)`

SetIcon sets Icon field to given value.

### HasIcon

`func (o *DtoGuild) HasIcon() bool`

HasIcon returns a boolean if a field has been set.

### GetId

`func (o *DtoGuild) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DtoGuild) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DtoGuild) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *DtoGuild) HasId() bool`

HasId returns a boolean if a field has been set.

### GetName

`func (o *DtoGuild) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DtoGuild) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DtoGuild) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *DtoGuild) HasName() bool`

HasName returns a boolean if a field has been set.

### GetOwner

`func (o *DtoGuild) GetOwner() int32`

GetOwner returns the Owner field if non-nil, zero value otherwise.

### GetOwnerOk

`func (o *DtoGuild) GetOwnerOk() (*int32, bool)`

GetOwnerOk returns a tuple with the Owner field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOwner

`func (o *DtoGuild) SetOwner(v int32)`

SetOwner sets Owner field to given value.

### HasOwner

`func (o *DtoGuild) HasOwner() bool`

HasOwner returns a boolean if a field has been set.

### GetPermissions

`func (o *DtoGuild) GetPermissions() int32`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *DtoGuild) GetPermissionsOk() (*int32, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *DtoGuild) SetPermissions(v int32)`

SetPermissions sets Permissions field to given value.

### HasPermissions

`func (o *DtoGuild) HasPermissions() bool`

HasPermissions returns a boolean if a field has been set.

### GetPublic

`func (o *DtoGuild) GetPublic() bool`

GetPublic returns the Public field if non-nil, zero value otherwise.

### GetPublicOk

`func (o *DtoGuild) GetPublicOk() (*bool, bool)`

GetPublicOk returns a tuple with the Public field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublic

`func (o *DtoGuild) SetPublic(v bool)`

SetPublic sets Public field to given value.

### HasPublic

`func (o *DtoGuild) HasPublic() bool`

HasPublic returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


