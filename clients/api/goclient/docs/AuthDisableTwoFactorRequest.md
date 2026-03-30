# AuthDisableTwoFactorRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Code** | Pointer to **string** |  | [optional] 
**CodeType** | Pointer to **string** |  | [optional] 
**CurrentPassword** | Pointer to **string** |  | [optional] 

## Methods

### NewAuthDisableTwoFactorRequest

`func NewAuthDisableTwoFactorRequest() *AuthDisableTwoFactorRequest`

NewAuthDisableTwoFactorRequest instantiates a new AuthDisableTwoFactorRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAuthDisableTwoFactorRequestWithDefaults

`func NewAuthDisableTwoFactorRequestWithDefaults() *AuthDisableTwoFactorRequest`

NewAuthDisableTwoFactorRequestWithDefaults instantiates a new AuthDisableTwoFactorRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCode

`func (o *AuthDisableTwoFactorRequest) GetCode() string`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *AuthDisableTwoFactorRequest) GetCodeOk() (*string, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *AuthDisableTwoFactorRequest) SetCode(v string)`

SetCode sets Code field to given value.

### HasCode

`func (o *AuthDisableTwoFactorRequest) HasCode() bool`

HasCode returns a boolean if a field has been set.

### GetCodeType

`func (o *AuthDisableTwoFactorRequest) GetCodeType() string`

GetCodeType returns the CodeType field if non-nil, zero value otherwise.

### GetCodeTypeOk

`func (o *AuthDisableTwoFactorRequest) GetCodeTypeOk() (*string, bool)`

GetCodeTypeOk returns a tuple with the CodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCodeType

`func (o *AuthDisableTwoFactorRequest) SetCodeType(v string)`

SetCodeType sets CodeType field to given value.

### HasCodeType

`func (o *AuthDisableTwoFactorRequest) HasCodeType() bool`

HasCodeType returns a boolean if a field has been set.

### GetCurrentPassword

`func (o *AuthDisableTwoFactorRequest) GetCurrentPassword() string`

GetCurrentPassword returns the CurrentPassword field if non-nil, zero value otherwise.

### GetCurrentPasswordOk

`func (o *AuthDisableTwoFactorRequest) GetCurrentPasswordOk() (*string, bool)`

GetCurrentPasswordOk returns a tuple with the CurrentPassword field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentPassword

`func (o *AuthDisableTwoFactorRequest) SetCurrentPassword(v string)`

SetCurrentPassword sets CurrentPassword field to given value.

### HasCurrentPassword

`func (o *AuthDisableTwoFactorRequest) HasCurrentPassword() bool`

HasCurrentPassword returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


