# GuildCreateGuildRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**IconId** | Pointer to **int32** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**Public** | Pointer to **bool** |  | [optional] 

## Methods

### NewGuildCreateGuildRequest

`func NewGuildCreateGuildRequest() *GuildCreateGuildRequest`

NewGuildCreateGuildRequest instantiates a new GuildCreateGuildRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGuildCreateGuildRequestWithDefaults

`func NewGuildCreateGuildRequestWithDefaults() *GuildCreateGuildRequest`

NewGuildCreateGuildRequestWithDefaults instantiates a new GuildCreateGuildRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetIconId

`func (o *GuildCreateGuildRequest) GetIconId() int32`

GetIconId returns the IconId field if non-nil, zero value otherwise.

### GetIconIdOk

`func (o *GuildCreateGuildRequest) GetIconIdOk() (*int32, bool)`

GetIconIdOk returns a tuple with the IconId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIconId

`func (o *GuildCreateGuildRequest) SetIconId(v int32)`

SetIconId sets IconId field to given value.

### HasIconId

`func (o *GuildCreateGuildRequest) HasIconId() bool`

HasIconId returns a boolean if a field has been set.

### GetName

`func (o *GuildCreateGuildRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *GuildCreateGuildRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *GuildCreateGuildRequest) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *GuildCreateGuildRequest) HasName() bool`

HasName returns a boolean if a field has been set.

### GetPublic

`func (o *GuildCreateGuildRequest) GetPublic() bool`

GetPublic returns the Public field if non-nil, zero value otherwise.

### GetPublicOk

`func (o *GuildCreateGuildRequest) GetPublicOk() (*bool, bool)`

GetPublicOk returns a tuple with the Public field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublic

`func (o *GuildCreateGuildRequest) SetPublic(v bool)`

SetPublic sets Public field to given value.

### HasPublic

`func (o *GuildCreateGuildRequest) HasPublic() bool`

HasPublic returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


