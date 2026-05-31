# GuildInstalledBotResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BotUserId** | Pointer to **int32** |  | [optional] 
**CreatedAt** | Pointer to **int32** |  | [optional] 
**GrantId** | Pointer to **int32** |  | [optional] 
**GrantedPermissions** | Pointer to **int32** |  | [optional] 
**InstallerUserId** | Pointer to **int32** |  | [optional] 
**Member** | Pointer to [**DtoMember**](DtoMember.md) |  | [optional] 
**Roles** | Pointer to **[]int32** |  | [optional] 
**User** | Pointer to [**DtoUser**](DtoUser.md) |  | [optional] 

## Methods

### NewGuildInstalledBotResponse

`func NewGuildInstalledBotResponse() *GuildInstalledBotResponse`

NewGuildInstalledBotResponse instantiates a new GuildInstalledBotResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGuildInstalledBotResponseWithDefaults

`func NewGuildInstalledBotResponseWithDefaults() *GuildInstalledBotResponse`

NewGuildInstalledBotResponseWithDefaults instantiates a new GuildInstalledBotResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBotUserId

`func (o *GuildInstalledBotResponse) GetBotUserId() int32`

GetBotUserId returns the BotUserId field if non-nil, zero value otherwise.

### GetBotUserIdOk

`func (o *GuildInstalledBotResponse) GetBotUserIdOk() (*int32, bool)`

GetBotUserIdOk returns a tuple with the BotUserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBotUserId

`func (o *GuildInstalledBotResponse) SetBotUserId(v int32)`

SetBotUserId sets BotUserId field to given value.

### HasBotUserId

`func (o *GuildInstalledBotResponse) HasBotUserId() bool`

HasBotUserId returns a boolean if a field has been set.

### GetCreatedAt

`func (o *GuildInstalledBotResponse) GetCreatedAt() int32`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *GuildInstalledBotResponse) GetCreatedAtOk() (*int32, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *GuildInstalledBotResponse) SetCreatedAt(v int32)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *GuildInstalledBotResponse) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetGrantId

`func (o *GuildInstalledBotResponse) GetGrantId() int32`

GetGrantId returns the GrantId field if non-nil, zero value otherwise.

### GetGrantIdOk

`func (o *GuildInstalledBotResponse) GetGrantIdOk() (*int32, bool)`

GetGrantIdOk returns a tuple with the GrantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGrantId

`func (o *GuildInstalledBotResponse) SetGrantId(v int32)`

SetGrantId sets GrantId field to given value.

### HasGrantId

`func (o *GuildInstalledBotResponse) HasGrantId() bool`

HasGrantId returns a boolean if a field has been set.

### GetGrantedPermissions

`func (o *GuildInstalledBotResponse) GetGrantedPermissions() int32`

GetGrantedPermissions returns the GrantedPermissions field if non-nil, zero value otherwise.

### GetGrantedPermissionsOk

`func (o *GuildInstalledBotResponse) GetGrantedPermissionsOk() (*int32, bool)`

GetGrantedPermissionsOk returns a tuple with the GrantedPermissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGrantedPermissions

`func (o *GuildInstalledBotResponse) SetGrantedPermissions(v int32)`

SetGrantedPermissions sets GrantedPermissions field to given value.

### HasGrantedPermissions

`func (o *GuildInstalledBotResponse) HasGrantedPermissions() bool`

HasGrantedPermissions returns a boolean if a field has been set.

### GetInstallerUserId

`func (o *GuildInstalledBotResponse) GetInstallerUserId() int32`

GetInstallerUserId returns the InstallerUserId field if non-nil, zero value otherwise.

### GetInstallerUserIdOk

`func (o *GuildInstalledBotResponse) GetInstallerUserIdOk() (*int32, bool)`

GetInstallerUserIdOk returns a tuple with the InstallerUserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstallerUserId

`func (o *GuildInstalledBotResponse) SetInstallerUserId(v int32)`

SetInstallerUserId sets InstallerUserId field to given value.

### HasInstallerUserId

`func (o *GuildInstalledBotResponse) HasInstallerUserId() bool`

HasInstallerUserId returns a boolean if a field has been set.

### GetMember

`func (o *GuildInstalledBotResponse) GetMember() DtoMember`

GetMember returns the Member field if non-nil, zero value otherwise.

### GetMemberOk

`func (o *GuildInstalledBotResponse) GetMemberOk() (*DtoMember, bool)`

GetMemberOk returns a tuple with the Member field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMember

`func (o *GuildInstalledBotResponse) SetMember(v DtoMember)`

SetMember sets Member field to given value.

### HasMember

`func (o *GuildInstalledBotResponse) HasMember() bool`

HasMember returns a boolean if a field has been set.

### GetRoles

`func (o *GuildInstalledBotResponse) GetRoles() []int32`

GetRoles returns the Roles field if non-nil, zero value otherwise.

### GetRolesOk

`func (o *GuildInstalledBotResponse) GetRolesOk() (*[]int32, bool)`

GetRolesOk returns a tuple with the Roles field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoles

`func (o *GuildInstalledBotResponse) SetRoles(v []int32)`

SetRoles sets Roles field to given value.

### HasRoles

`func (o *GuildInstalledBotResponse) HasRoles() bool`

HasRoles returns a boolean if a field has been set.

### GetUser

`func (o *GuildInstalledBotResponse) GetUser() DtoUser`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *GuildInstalledBotResponse) GetUserOk() (*DtoUser, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *GuildInstalledBotResponse) SetUser(v DtoUser)`

SetUser sets User field to given value.

### HasUser

`func (o *GuildInstalledBotResponse) HasUser() bool`

HasUser returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


