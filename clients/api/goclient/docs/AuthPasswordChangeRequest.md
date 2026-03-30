# AuthPasswordChangeRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Code** | Pointer to **string** |  | [optional] 
**CodeType** | Pointer to **string** |  | [optional] 
**CurrentPassword** | Pointer to **string** |  | [optional] 
**NewPassword** | Pointer to **string** |  | [optional] 

## Methods

### NewAuthPasswordChangeRequest

`func NewAuthPasswordChangeRequest() *AuthPasswordChangeRequest`

NewAuthPasswordChangeRequest instantiates a new AuthPasswordChangeRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAuthPasswordChangeRequestWithDefaults

`func NewAuthPasswordChangeRequestWithDefaults() *AuthPasswordChangeRequest`

NewAuthPasswordChangeRequestWithDefaults instantiates a new AuthPasswordChangeRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCode

`func (o *AuthPasswordChangeRequest) GetCode() string`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *AuthPasswordChangeRequest) GetCodeOk() (*string, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *AuthPasswordChangeRequest) SetCode(v string)`

SetCode sets Code field to given value.

### HasCode

`func (o *AuthPasswordChangeRequest) HasCode() bool`

HasCode returns a boolean if a field has been set.

### GetCodeType

`func (o *AuthPasswordChangeRequest) GetCodeType() string`

GetCodeType returns the CodeType field if non-nil, zero value otherwise.

### GetCodeTypeOk

`func (o *AuthPasswordChangeRequest) GetCodeTypeOk() (*string, bool)`

GetCodeTypeOk returns a tuple with the CodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCodeType

`func (o *AuthPasswordChangeRequest) SetCodeType(v string)`

SetCodeType sets CodeType field to given value.

### HasCodeType

`func (o *AuthPasswordChangeRequest) HasCodeType() bool`

HasCodeType returns a boolean if a field has been set.

### GetCurrentPassword

`func (o *AuthPasswordChangeRequest) GetCurrentPassword() string`

GetCurrentPassword returns the CurrentPassword field if non-nil, zero value otherwise.

### GetCurrentPasswordOk

`func (o *AuthPasswordChangeRequest) GetCurrentPasswordOk() (*string, bool)`

GetCurrentPasswordOk returns a tuple with the CurrentPassword field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentPassword

`func (o *AuthPasswordChangeRequest) SetCurrentPassword(v string)`

SetCurrentPassword sets CurrentPassword field to given value.

### HasCurrentPassword

`func (o *AuthPasswordChangeRequest) HasCurrentPassword() bool`

HasCurrentPassword returns a boolean if a field has been set.

### GetNewPassword

`func (o *AuthPasswordChangeRequest) GetNewPassword() string`

GetNewPassword returns the NewPassword field if non-nil, zero value otherwise.

### GetNewPasswordOk

`func (o *AuthPasswordChangeRequest) GetNewPasswordOk() (*string, bool)`

GetNewPasswordOk returns a tuple with the NewPassword field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNewPassword

`func (o *AuthPasswordChangeRequest) SetNewPassword(v string)`

SetNewPassword sets NewPassword field to given value.

### HasNewPassword

`func (o *AuthPasswordChangeRequest) HasNewPassword() bool`

HasNewPassword returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


