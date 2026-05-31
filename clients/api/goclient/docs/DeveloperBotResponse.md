# DeveloperBotResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BotUserId** | Pointer to **int32** |  | [optional] 
**CreatedAt** | Pointer to **int32** |  | [optional] 
**DefaultPermissions** | Pointer to **int32** |  | [optional] 
**Description** | Pointer to **string** |  | [optional] 
**Disabled** | Pointer to **bool** |  | [optional] 
**OwnerUserId** | Pointer to **int32** |  | [optional] 
**Public** | Pointer to **bool** |  | [optional] 
**Tags** | Pointer to **[]string** |  | [optional] 
**UpdatedAt** | Pointer to **int32** |  | [optional] 
**User** | Pointer to [**DtoUser**](DtoUser.md) |  | [optional] 

## Methods

### NewDeveloperBotResponse

`func NewDeveloperBotResponse() *DeveloperBotResponse`

NewDeveloperBotResponse instantiates a new DeveloperBotResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeveloperBotResponseWithDefaults

`func NewDeveloperBotResponseWithDefaults() *DeveloperBotResponse`

NewDeveloperBotResponseWithDefaults instantiates a new DeveloperBotResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBotUserId

`func (o *DeveloperBotResponse) GetBotUserId() int32`

GetBotUserId returns the BotUserId field if non-nil, zero value otherwise.

### GetBotUserIdOk

`func (o *DeveloperBotResponse) GetBotUserIdOk() (*int32, bool)`

GetBotUserIdOk returns a tuple with the BotUserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBotUserId

`func (o *DeveloperBotResponse) SetBotUserId(v int32)`

SetBotUserId sets BotUserId field to given value.

### HasBotUserId

`func (o *DeveloperBotResponse) HasBotUserId() bool`

HasBotUserId returns a boolean if a field has been set.

### GetCreatedAt

`func (o *DeveloperBotResponse) GetCreatedAt() int32`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *DeveloperBotResponse) GetCreatedAtOk() (*int32, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *DeveloperBotResponse) SetCreatedAt(v int32)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *DeveloperBotResponse) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetDefaultPermissions

`func (o *DeveloperBotResponse) GetDefaultPermissions() int32`

GetDefaultPermissions returns the DefaultPermissions field if non-nil, zero value otherwise.

### GetDefaultPermissionsOk

`func (o *DeveloperBotResponse) GetDefaultPermissionsOk() (*int32, bool)`

GetDefaultPermissionsOk returns a tuple with the DefaultPermissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultPermissions

`func (o *DeveloperBotResponse) SetDefaultPermissions(v int32)`

SetDefaultPermissions sets DefaultPermissions field to given value.

### HasDefaultPermissions

`func (o *DeveloperBotResponse) HasDefaultPermissions() bool`

HasDefaultPermissions returns a boolean if a field has been set.

### GetDescription

`func (o *DeveloperBotResponse) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *DeveloperBotResponse) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *DeveloperBotResponse) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *DeveloperBotResponse) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetDisabled

`func (o *DeveloperBotResponse) GetDisabled() bool`

GetDisabled returns the Disabled field if non-nil, zero value otherwise.

### GetDisabledOk

`func (o *DeveloperBotResponse) GetDisabledOk() (*bool, bool)`

GetDisabledOk returns a tuple with the Disabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisabled

`func (o *DeveloperBotResponse) SetDisabled(v bool)`

SetDisabled sets Disabled field to given value.

### HasDisabled

`func (o *DeveloperBotResponse) HasDisabled() bool`

HasDisabled returns a boolean if a field has been set.

### GetOwnerUserId

`func (o *DeveloperBotResponse) GetOwnerUserId() int32`

GetOwnerUserId returns the OwnerUserId field if non-nil, zero value otherwise.

### GetOwnerUserIdOk

`func (o *DeveloperBotResponse) GetOwnerUserIdOk() (*int32, bool)`

GetOwnerUserIdOk returns a tuple with the OwnerUserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOwnerUserId

`func (o *DeveloperBotResponse) SetOwnerUserId(v int32)`

SetOwnerUserId sets OwnerUserId field to given value.

### HasOwnerUserId

`func (o *DeveloperBotResponse) HasOwnerUserId() bool`

HasOwnerUserId returns a boolean if a field has been set.

### GetPublic

`func (o *DeveloperBotResponse) GetPublic() bool`

GetPublic returns the Public field if non-nil, zero value otherwise.

### GetPublicOk

`func (o *DeveloperBotResponse) GetPublicOk() (*bool, bool)`

GetPublicOk returns a tuple with the Public field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublic

`func (o *DeveloperBotResponse) SetPublic(v bool)`

SetPublic sets Public field to given value.

### HasPublic

`func (o *DeveloperBotResponse) HasPublic() bool`

HasPublic returns a boolean if a field has been set.

### GetTags

`func (o *DeveloperBotResponse) GetTags() []string`

GetTags returns the Tags field if non-nil, zero value otherwise.

### GetTagsOk

`func (o *DeveloperBotResponse) GetTagsOk() (*[]string, bool)`

GetTagsOk returns a tuple with the Tags field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTags

`func (o *DeveloperBotResponse) SetTags(v []string)`

SetTags sets Tags field to given value.

### HasTags

`func (o *DeveloperBotResponse) HasTags() bool`

HasTags returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *DeveloperBotResponse) GetUpdatedAt() int32`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *DeveloperBotResponse) GetUpdatedAtOk() (*int32, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *DeveloperBotResponse) SetUpdatedAt(v int32)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *DeveloperBotResponse) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.

### GetUser

`func (o *DeveloperBotResponse) GetUser() DtoUser`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *DeveloperBotResponse) GetUserOk() (*DtoUser, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *DeveloperBotResponse) SetUser(v DtoUser)`

SetUser sets User field to given value.

### HasUser

`func (o *DeveloperBotResponse) HasUser() bool`

HasUser returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


