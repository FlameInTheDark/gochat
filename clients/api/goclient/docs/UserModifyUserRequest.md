# UserModifyUserRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Avatar** | Pointer to **int32** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 

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


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


