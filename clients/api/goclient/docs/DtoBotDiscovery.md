# DtoBotDiscovery

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BotUserId** | Pointer to **int32** |  | [optional] 
**CreatedAt** | Pointer to **int32** |  | [optional] 
**DefaultPermissions** | Pointer to **int32** |  | [optional] 
**Description** | Pointer to **string** |  | [optional] 
**InstallsCount** | Pointer to **int32** |  | [optional] 
**Tags** | Pointer to **[]string** |  | [optional] 
**UpdatedAt** | Pointer to **int32** |  | [optional] 
**User** | Pointer to [**DtoUser**](DtoUser.md) |  | [optional] 

## Methods

### NewDtoBotDiscovery

`func NewDtoBotDiscovery() *DtoBotDiscovery`

NewDtoBotDiscovery instantiates a new DtoBotDiscovery object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDtoBotDiscoveryWithDefaults

`func NewDtoBotDiscoveryWithDefaults() *DtoBotDiscovery`

NewDtoBotDiscoveryWithDefaults instantiates a new DtoBotDiscovery object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBotUserId

`func (o *DtoBotDiscovery) GetBotUserId() int32`

GetBotUserId returns the BotUserId field if non-nil, zero value otherwise.

### GetBotUserIdOk

`func (o *DtoBotDiscovery) GetBotUserIdOk() (*int32, bool)`

GetBotUserIdOk returns a tuple with the BotUserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBotUserId

`func (o *DtoBotDiscovery) SetBotUserId(v int32)`

SetBotUserId sets BotUserId field to given value.

### HasBotUserId

`func (o *DtoBotDiscovery) HasBotUserId() bool`

HasBotUserId returns a boolean if a field has been set.

### GetCreatedAt

`func (o *DtoBotDiscovery) GetCreatedAt() int32`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *DtoBotDiscovery) GetCreatedAtOk() (*int32, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *DtoBotDiscovery) SetCreatedAt(v int32)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *DtoBotDiscovery) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetDefaultPermissions

`func (o *DtoBotDiscovery) GetDefaultPermissions() int32`

GetDefaultPermissions returns the DefaultPermissions field if non-nil, zero value otherwise.

### GetDefaultPermissionsOk

`func (o *DtoBotDiscovery) GetDefaultPermissionsOk() (*int32, bool)`

GetDefaultPermissionsOk returns a tuple with the DefaultPermissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultPermissions

`func (o *DtoBotDiscovery) SetDefaultPermissions(v int32)`

SetDefaultPermissions sets DefaultPermissions field to given value.

### HasDefaultPermissions

`func (o *DtoBotDiscovery) HasDefaultPermissions() bool`

HasDefaultPermissions returns a boolean if a field has been set.

### GetDescription

`func (o *DtoBotDiscovery) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *DtoBotDiscovery) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *DtoBotDiscovery) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *DtoBotDiscovery) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetInstallsCount

`func (o *DtoBotDiscovery) GetInstallsCount() int32`

GetInstallsCount returns the InstallsCount field if non-nil, zero value otherwise.

### GetInstallsCountOk

`func (o *DtoBotDiscovery) GetInstallsCountOk() (*int32, bool)`

GetInstallsCountOk returns a tuple with the InstallsCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstallsCount

`func (o *DtoBotDiscovery) SetInstallsCount(v int32)`

SetInstallsCount sets InstallsCount field to given value.

### HasInstallsCount

`func (o *DtoBotDiscovery) HasInstallsCount() bool`

HasInstallsCount returns a boolean if a field has been set.

### GetTags

`func (o *DtoBotDiscovery) GetTags() []string`

GetTags returns the Tags field if non-nil, zero value otherwise.

### GetTagsOk

`func (o *DtoBotDiscovery) GetTagsOk() (*[]string, bool)`

GetTagsOk returns a tuple with the Tags field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTags

`func (o *DtoBotDiscovery) SetTags(v []string)`

SetTags sets Tags field to given value.

### HasTags

`func (o *DtoBotDiscovery) HasTags() bool`

HasTags returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *DtoBotDiscovery) GetUpdatedAt() int32`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *DtoBotDiscovery) GetUpdatedAtOk() (*int32, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *DtoBotDiscovery) SetUpdatedAt(v int32)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *DtoBotDiscovery) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.

### GetUser

`func (o *DtoBotDiscovery) GetUser() DtoUser`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *DtoBotDiscovery) GetUserOk() (*DtoUser, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *DtoBotDiscovery) SetUser(v DtoUser)`

SetUser sets User field to given value.

### HasUser

`func (o *DtoBotDiscovery) HasUser() bool`

HasUser returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


