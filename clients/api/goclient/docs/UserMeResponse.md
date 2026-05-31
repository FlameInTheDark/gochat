# UserMeResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BotUserId** | Pointer to **int32** |  | [optional] 
**DefaultPermissions** | Pointer to **int32** |  | [optional] 
**Description** | Pointer to **string** |  | [optional] 
**OwnerUserId** | Pointer to **int32** |  | [optional] 
**Public** | Pointer to **bool** |  | [optional] 
**User** | Pointer to [**DtoUser**](DtoUser.md) |  | [optional] 

## Methods

### NewUserMeResponse

`func NewUserMeResponse() *UserMeResponse`

NewUserMeResponse instantiates a new UserMeResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUserMeResponseWithDefaults

`func NewUserMeResponseWithDefaults() *UserMeResponse`

NewUserMeResponseWithDefaults instantiates a new UserMeResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBotUserId

`func (o *UserMeResponse) GetBotUserId() int32`

GetBotUserId returns the BotUserId field if non-nil, zero value otherwise.

### GetBotUserIdOk

`func (o *UserMeResponse) GetBotUserIdOk() (*int32, bool)`

GetBotUserIdOk returns a tuple with the BotUserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBotUserId

`func (o *UserMeResponse) SetBotUserId(v int32)`

SetBotUserId sets BotUserId field to given value.

### HasBotUserId

`func (o *UserMeResponse) HasBotUserId() bool`

HasBotUserId returns a boolean if a field has been set.

### GetDefaultPermissions

`func (o *UserMeResponse) GetDefaultPermissions() int32`

GetDefaultPermissions returns the DefaultPermissions field if non-nil, zero value otherwise.

### GetDefaultPermissionsOk

`func (o *UserMeResponse) GetDefaultPermissionsOk() (*int32, bool)`

GetDefaultPermissionsOk returns a tuple with the DefaultPermissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultPermissions

`func (o *UserMeResponse) SetDefaultPermissions(v int32)`

SetDefaultPermissions sets DefaultPermissions field to given value.

### HasDefaultPermissions

`func (o *UserMeResponse) HasDefaultPermissions() bool`

HasDefaultPermissions returns a boolean if a field has been set.

### GetDescription

`func (o *UserMeResponse) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *UserMeResponse) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *UserMeResponse) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *UserMeResponse) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetOwnerUserId

`func (o *UserMeResponse) GetOwnerUserId() int32`

GetOwnerUserId returns the OwnerUserId field if non-nil, zero value otherwise.

### GetOwnerUserIdOk

`func (o *UserMeResponse) GetOwnerUserIdOk() (*int32, bool)`

GetOwnerUserIdOk returns a tuple with the OwnerUserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOwnerUserId

`func (o *UserMeResponse) SetOwnerUserId(v int32)`

SetOwnerUserId sets OwnerUserId field to given value.

### HasOwnerUserId

`func (o *UserMeResponse) HasOwnerUserId() bool`

HasOwnerUserId returns a boolean if a field has been set.

### GetPublic

`func (o *UserMeResponse) GetPublic() bool`

GetPublic returns the Public field if non-nil, zero value otherwise.

### GetPublicOk

`func (o *UserMeResponse) GetPublicOk() (*bool, bool)`

GetPublicOk returns a tuple with the Public field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublic

`func (o *UserMeResponse) SetPublic(v bool)`

SetPublic sets Public field to given value.

### HasPublic

`func (o *UserMeResponse) HasPublic() bool`

HasPublic returns a boolean if a field has been set.

### GetUser

`func (o *UserMeResponse) GetUser() DtoUser`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *UserMeResponse) GetUserOk() (*DtoUser, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *UserMeResponse) SetUser(v DtoUser)`

SetUser sets User field to given value.

### HasUser

`func (o *UserMeResponse) HasUser() bool`

HasUser returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


