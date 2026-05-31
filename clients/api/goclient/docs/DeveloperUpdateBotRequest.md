# DeveloperUpdateBotRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Avatar** | Pointer to **int32** |  | [optional] 
**Banner** | Pointer to **int32** |  | [optional] 
**BannerColor** | Pointer to **int32** |  | [optional] 
**Bio** | Pointer to **string** |  | [optional] 
**DefaultPermissions** | Pointer to **int32** |  | [optional] 
**Description** | Pointer to **string** |  | [optional] 
**Disabled** | Pointer to **bool** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**PanelColor** | Pointer to **int32** |  | [optional] 
**Public** | Pointer to **bool** |  | [optional] 
**Tags** | Pointer to **[]string** |  | [optional] 

## Methods

### NewDeveloperUpdateBotRequest

`func NewDeveloperUpdateBotRequest() *DeveloperUpdateBotRequest`

NewDeveloperUpdateBotRequest instantiates a new DeveloperUpdateBotRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeveloperUpdateBotRequestWithDefaults

`func NewDeveloperUpdateBotRequestWithDefaults() *DeveloperUpdateBotRequest`

NewDeveloperUpdateBotRequestWithDefaults instantiates a new DeveloperUpdateBotRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAvatar

`func (o *DeveloperUpdateBotRequest) GetAvatar() int32`

GetAvatar returns the Avatar field if non-nil, zero value otherwise.

### GetAvatarOk

`func (o *DeveloperUpdateBotRequest) GetAvatarOk() (*int32, bool)`

GetAvatarOk returns a tuple with the Avatar field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAvatar

`func (o *DeveloperUpdateBotRequest) SetAvatar(v int32)`

SetAvatar sets Avatar field to given value.

### HasAvatar

`func (o *DeveloperUpdateBotRequest) HasAvatar() bool`

HasAvatar returns a boolean if a field has been set.

### GetBanner

`func (o *DeveloperUpdateBotRequest) GetBanner() int32`

GetBanner returns the Banner field if non-nil, zero value otherwise.

### GetBannerOk

`func (o *DeveloperUpdateBotRequest) GetBannerOk() (*int32, bool)`

GetBannerOk returns a tuple with the Banner field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBanner

`func (o *DeveloperUpdateBotRequest) SetBanner(v int32)`

SetBanner sets Banner field to given value.

### HasBanner

`func (o *DeveloperUpdateBotRequest) HasBanner() bool`

HasBanner returns a boolean if a field has been set.

### GetBannerColor

`func (o *DeveloperUpdateBotRequest) GetBannerColor() int32`

GetBannerColor returns the BannerColor field if non-nil, zero value otherwise.

### GetBannerColorOk

`func (o *DeveloperUpdateBotRequest) GetBannerColorOk() (*int32, bool)`

GetBannerColorOk returns a tuple with the BannerColor field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBannerColor

`func (o *DeveloperUpdateBotRequest) SetBannerColor(v int32)`

SetBannerColor sets BannerColor field to given value.

### HasBannerColor

`func (o *DeveloperUpdateBotRequest) HasBannerColor() bool`

HasBannerColor returns a boolean if a field has been set.

### GetBio

`func (o *DeveloperUpdateBotRequest) GetBio() string`

GetBio returns the Bio field if non-nil, zero value otherwise.

### GetBioOk

`func (o *DeveloperUpdateBotRequest) GetBioOk() (*string, bool)`

GetBioOk returns a tuple with the Bio field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBio

`func (o *DeveloperUpdateBotRequest) SetBio(v string)`

SetBio sets Bio field to given value.

### HasBio

`func (o *DeveloperUpdateBotRequest) HasBio() bool`

HasBio returns a boolean if a field has been set.

### GetDefaultPermissions

`func (o *DeveloperUpdateBotRequest) GetDefaultPermissions() int32`

GetDefaultPermissions returns the DefaultPermissions field if non-nil, zero value otherwise.

### GetDefaultPermissionsOk

`func (o *DeveloperUpdateBotRequest) GetDefaultPermissionsOk() (*int32, bool)`

GetDefaultPermissionsOk returns a tuple with the DefaultPermissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultPermissions

`func (o *DeveloperUpdateBotRequest) SetDefaultPermissions(v int32)`

SetDefaultPermissions sets DefaultPermissions field to given value.

### HasDefaultPermissions

`func (o *DeveloperUpdateBotRequest) HasDefaultPermissions() bool`

HasDefaultPermissions returns a boolean if a field has been set.

### GetDescription

`func (o *DeveloperUpdateBotRequest) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *DeveloperUpdateBotRequest) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *DeveloperUpdateBotRequest) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *DeveloperUpdateBotRequest) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetDisabled

`func (o *DeveloperUpdateBotRequest) GetDisabled() bool`

GetDisabled returns the Disabled field if non-nil, zero value otherwise.

### GetDisabledOk

`func (o *DeveloperUpdateBotRequest) GetDisabledOk() (*bool, bool)`

GetDisabledOk returns a tuple with the Disabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDisabled

`func (o *DeveloperUpdateBotRequest) SetDisabled(v bool)`

SetDisabled sets Disabled field to given value.

### HasDisabled

`func (o *DeveloperUpdateBotRequest) HasDisabled() bool`

HasDisabled returns a boolean if a field has been set.

### GetName

`func (o *DeveloperUpdateBotRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DeveloperUpdateBotRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DeveloperUpdateBotRequest) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *DeveloperUpdateBotRequest) HasName() bool`

HasName returns a boolean if a field has been set.

### GetPanelColor

`func (o *DeveloperUpdateBotRequest) GetPanelColor() int32`

GetPanelColor returns the PanelColor field if non-nil, zero value otherwise.

### GetPanelColorOk

`func (o *DeveloperUpdateBotRequest) GetPanelColorOk() (*int32, bool)`

GetPanelColorOk returns a tuple with the PanelColor field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPanelColor

`func (o *DeveloperUpdateBotRequest) SetPanelColor(v int32)`

SetPanelColor sets PanelColor field to given value.

### HasPanelColor

`func (o *DeveloperUpdateBotRequest) HasPanelColor() bool`

HasPanelColor returns a boolean if a field has been set.

### GetPublic

`func (o *DeveloperUpdateBotRequest) GetPublic() bool`

GetPublic returns the Public field if non-nil, zero value otherwise.

### GetPublicOk

`func (o *DeveloperUpdateBotRequest) GetPublicOk() (*bool, bool)`

GetPublicOk returns a tuple with the Public field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublic

`func (o *DeveloperUpdateBotRequest) SetPublic(v bool)`

SetPublic sets Public field to given value.

### HasPublic

`func (o *DeveloperUpdateBotRequest) HasPublic() bool`

HasPublic returns a boolean if a field has been set.

### GetTags

`func (o *DeveloperUpdateBotRequest) GetTags() []string`

GetTags returns the Tags field if non-nil, zero value otherwise.

### GetTagsOk

`func (o *DeveloperUpdateBotRequest) GetTagsOk() (*[]string, bool)`

GetTagsOk returns a tuple with the Tags field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTags

`func (o *DeveloperUpdateBotRequest) SetTags(v []string)`

SetTags sets Tags field to given value.

### HasTags

`func (o *DeveloperUpdateBotRequest) HasTags() bool`

HasTags returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


