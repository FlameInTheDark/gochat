# UserModifyUserRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Avatar** | Pointer to **int32** | Avatar ID. | [optional] 
**BannerColor** | Pointer to **int32** | Banner/header RGB int value. | [optional] 
**Bio** | Pointer to **string** | Public user bio. | [optional] 
**Name** | Pointer to **string** | User name. | [optional] 
**PanelColor** | Pointer to **int32** | User panel RGB int value. | [optional] 

## Methods

### NewUserModifyUserRequest

`func NewUserModifyUserRequest() *UserModifyUserRequest`

NewUserModifyUserRequest instantiates a new UserModifyUserRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUserModifyUserRequestWithDefaults

`func NewUserModifyUserRequestWithDefaults() *UserModifyUserRequest`

NewUserModifyUserRequestWithDefaults instantiates a new UserModifyUserRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAvatar

`func (o *UserModifyUserRequest) GetAvatar() int32`

GetAvatar returns the Avatar field if non-nil, zero value otherwise.

### GetAvatarOk

`func (o *UserModifyUserRequest) GetAvatarOk() (*int32, bool)`

GetAvatarOk returns a tuple with the Avatar field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAvatar

`func (o *UserModifyUserRequest) SetAvatar(v int32)`

SetAvatar sets Avatar field to given value.

### HasAvatar

`func (o *UserModifyUserRequest) HasAvatar() bool`

HasAvatar returns a boolean if a field has been set.

### GetBannerColor

`func (o *UserModifyUserRequest) GetBannerColor() int32`

GetBannerColor returns the BannerColor field if non-nil, zero value otherwise.

### GetBannerColorOk

`func (o *UserModifyUserRequest) GetBannerColorOk() (*int32, bool)`

GetBannerColorOk returns a tuple with the BannerColor field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBannerColor

`func (o *UserModifyUserRequest) SetBannerColor(v int32)`

SetBannerColor sets BannerColor field to given value.

### HasBannerColor

`func (o *UserModifyUserRequest) HasBannerColor() bool`

HasBannerColor returns a boolean if a field has been set.

### GetBio

`func (o *UserModifyUserRequest) GetBio() string`

GetBio returns the Bio field if non-nil, zero value otherwise.

### GetBioOk

`func (o *UserModifyUserRequest) GetBioOk() (*string, bool)`

GetBioOk returns a tuple with the Bio field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBio

`func (o *UserModifyUserRequest) SetBio(v string)`

SetBio sets Bio field to given value.

### HasBio

`func (o *UserModifyUserRequest) HasBio() bool`

HasBio returns a boolean if a field has been set.

### GetName

`func (o *UserModifyUserRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *UserModifyUserRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *UserModifyUserRequest) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *UserModifyUserRequest) HasName() bool`

HasName returns a boolean if a field has been set.

### GetPanelColor

`func (o *UserModifyUserRequest) GetPanelColor() int32`

GetPanelColor returns the PanelColor field if non-nil, zero value otherwise.

### GetPanelColorOk

`func (o *UserModifyUserRequest) GetPanelColorOk() (*int32, bool)`

GetPanelColorOk returns a tuple with the PanelColor field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPanelColor

`func (o *UserModifyUserRequest) SetPanelColor(v int32)`

SetPanelColor sets PanelColor field to given value.

### HasPanelColor

`func (o *UserModifyUserRequest) HasPanelColor() bool`

HasPanelColor returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


