# DtoUser

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Avatar** | Pointer to [**DtoAvatarData**](DtoAvatarData.md) |  | [optional] 
**BannerColor** | Pointer to **int32** |  | [optional] 
**Bio** | Pointer to **string** |  | [optional] 
**Discriminator** | Pointer to **string** |  | [optional] 
**Id** | Pointer to **int32** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**PanelColor** | Pointer to **int32** |  | [optional] 

## Methods

### NewDtoUser

`func NewDtoUser() *DtoUser`

NewDtoUser instantiates a new DtoUser object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDtoUserWithDefaults

`func NewDtoUserWithDefaults() *DtoUser`

NewDtoUserWithDefaults instantiates a new DtoUser object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAvatar

`func (o *DtoUser) GetAvatar() DtoAvatarData`

GetAvatar returns the Avatar field if non-nil, zero value otherwise.

### GetAvatarOk

`func (o *DtoUser) GetAvatarOk() (*DtoAvatarData, bool)`

GetAvatarOk returns a tuple with the Avatar field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAvatar

`func (o *DtoUser) SetAvatar(v DtoAvatarData)`

SetAvatar sets Avatar field to given value.

### HasAvatar

`func (o *DtoUser) HasAvatar() bool`

HasAvatar returns a boolean if a field has been set.

### GetBannerColor

`func (o *DtoUser) GetBannerColor() int32`

GetBannerColor returns the BannerColor field if non-nil, zero value otherwise.

### GetBannerColorOk

`func (o *DtoUser) GetBannerColorOk() (*int32, bool)`

GetBannerColorOk returns a tuple with the BannerColor field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBannerColor

`func (o *DtoUser) SetBannerColor(v int32)`

SetBannerColor sets BannerColor field to given value.

### HasBannerColor

`func (o *DtoUser) HasBannerColor() bool`

HasBannerColor returns a boolean if a field has been set.

### GetBio

`func (o *DtoUser) GetBio() string`

GetBio returns the Bio field if non-nil, zero value otherwise.

### GetBioOk

`func (o *DtoUser) GetBioOk() (*string, bool)`

GetBioOk returns a tuple with the Bio field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBio

`func (o *DtoUser) SetBio(v string)`

SetBio sets Bio field to given value.

### HasBio

`func (o *DtoUser) HasBio() bool`

HasBio returns a boolean if a field has been set.

### GetDiscriminator

`func (o *DtoUser) GetDiscriminator() string`

GetDiscriminator returns the Discriminator field if non-nil, zero value otherwise.

### GetDiscriminatorOk

`func (o *DtoUser) GetDiscriminatorOk() (*string, bool)`

GetDiscriminatorOk returns a tuple with the Discriminator field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDiscriminator

`func (o *DtoUser) SetDiscriminator(v string)`

SetDiscriminator sets Discriminator field to given value.

### HasDiscriminator

`func (o *DtoUser) HasDiscriminator() bool`

HasDiscriminator returns a boolean if a field has been set.

### GetId

`func (o *DtoUser) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DtoUser) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DtoUser) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *DtoUser) HasId() bool`

HasId returns a boolean if a field has been set.

### GetName

`func (o *DtoUser) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DtoUser) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DtoUser) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *DtoUser) HasName() bool`

HasName returns a boolean if a field has been set.

### GetPanelColor

`func (o *DtoUser) GetPanelColor() int32`

GetPanelColor returns the PanelColor field if non-nil, zero value otherwise.

### GetPanelColorOk

`func (o *DtoUser) GetPanelColorOk() (*int32, bool)`

GetPanelColorOk returns a tuple with the PanelColor field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPanelColor

`func (o *DtoUser) SetPanelColor(v int32)`

SetPanelColor sets PanelColor field to given value.

### HasPanelColor

`func (o *DtoUser) HasPanelColor() bool`

HasPanelColor returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


