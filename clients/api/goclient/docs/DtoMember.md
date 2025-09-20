# DtoMember

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Avatar** | Pointer to **int32** | Avatar ID | [optional] 
**JoinAt** | Pointer to **string** | Join date | [optional] 
**Roles** | Pointer to **[]int32** | List of assigned role IDs | [optional] 
**User** | Pointer to [**DtoUser**](DtoUser.md) |  | [optional] 
**Username** | Pointer to **string** | Username in this guild | [optional] 

## Methods

### NewDtoMember

`func NewDtoMember() *DtoMember`

NewDtoMember instantiates a new DtoMember object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDtoMemberWithDefaults

`func NewDtoMemberWithDefaults() *DtoMember`

NewDtoMemberWithDefaults instantiates a new DtoMember object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAvatar

`func (o *DtoMember) GetAvatar() int32`

GetAvatar returns the Avatar field if non-nil, zero value otherwise.

### GetAvatarOk

`func (o *DtoMember) GetAvatarOk() (*int32, bool)`

GetAvatarOk returns a tuple with the Avatar field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAvatar

`func (o *DtoMember) SetAvatar(v int32)`

SetAvatar sets Avatar field to given value.

### HasAvatar

`func (o *DtoMember) HasAvatar() bool`

HasAvatar returns a boolean if a field has been set.

### GetJoinAt

`func (o *DtoMember) GetJoinAt() string`

GetJoinAt returns the JoinAt field if non-nil, zero value otherwise.

### GetJoinAtOk

`func (o *DtoMember) GetJoinAtOk() (*string, bool)`

GetJoinAtOk returns a tuple with the JoinAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetJoinAt

`func (o *DtoMember) SetJoinAt(v string)`

SetJoinAt sets JoinAt field to given value.

### HasJoinAt

`func (o *DtoMember) HasJoinAt() bool`

HasJoinAt returns a boolean if a field has been set.

### GetRoles

`func (o *DtoMember) GetRoles() []int32`

GetRoles returns the Roles field if non-nil, zero value otherwise.

### GetRolesOk

`func (o *DtoMember) GetRolesOk() (*[]int32, bool)`

GetRolesOk returns a tuple with the Roles field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoles

`func (o *DtoMember) SetRoles(v []int32)`

SetRoles sets Roles field to given value.

### HasRoles

`func (o *DtoMember) HasRoles() bool`

HasRoles returns a boolean if a field has been set.

### GetUser

`func (o *DtoMember) GetUser() DtoUser`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *DtoMember) GetUserOk() (*DtoUser, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *DtoMember) SetUser(v DtoUser)`

SetUser sets User field to given value.

### HasUser

`func (o *DtoMember) HasUser() bool`

HasUser returns a boolean if a field has been set.

### GetUsername

`func (o *DtoMember) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *DtoMember) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *DtoMember) SetUsername(v string)`

SetUsername sets Username field to given value.

### HasUsername

`func (o *DtoMember) HasUsername() bool`

HasUsername returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


