# AuthRecoveryCodesRegenerateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Code** | Pointer to **string** |  | [optional] 
**CodeType** | Pointer to **string** |  | [optional] 
**CurrentPassword** | Pointer to **string** |  | [optional] 

## Methods

### NewAuthRecoveryCodesRegenerateRequest

`func NewAuthRecoveryCodesRegenerateRequest() *AuthRecoveryCodesRegenerateRequest`

NewAuthRecoveryCodesRegenerateRequest instantiates a new AuthRecoveryCodesRegenerateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAuthRecoveryCodesRegenerateRequestWithDefaults

`func NewAuthRecoveryCodesRegenerateRequestWithDefaults() *AuthRecoveryCodesRegenerateRequest`

NewAuthRecoveryCodesRegenerateRequestWithDefaults instantiates a new AuthRecoveryCodesRegenerateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCode

`func (o *AuthRecoveryCodesRegenerateRequest) GetCode() string`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *AuthRecoveryCodesRegenerateRequest) GetCodeOk() (*string, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *AuthRecoveryCodesRegenerateRequest) SetCode(v string)`

SetCode sets Code field to given value.

### HasCode

`func (o *AuthRecoveryCodesRegenerateRequest) HasCode() bool`

HasCode returns a boolean if a field has been set.

### GetCodeType

`func (o *AuthRecoveryCodesRegenerateRequest) GetCodeType() string`

GetCodeType returns the CodeType field if non-nil, zero value otherwise.

### GetCodeTypeOk

`func (o *AuthRecoveryCodesRegenerateRequest) GetCodeTypeOk() (*string, bool)`

GetCodeTypeOk returns a tuple with the CodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCodeType

`func (o *AuthRecoveryCodesRegenerateRequest) SetCodeType(v string)`

SetCodeType sets CodeType field to given value.

### HasCodeType

`func (o *AuthRecoveryCodesRegenerateRequest) HasCodeType() bool`

HasCodeType returns a boolean if a field has been set.

### GetCurrentPassword

`func (o *AuthRecoveryCodesRegenerateRequest) GetCurrentPassword() string`

GetCurrentPassword returns the CurrentPassword field if non-nil, zero value otherwise.

### GetCurrentPasswordOk

`func (o *AuthRecoveryCodesRegenerateRequest) GetCurrentPasswordOk() (*string, bool)`

GetCurrentPasswordOk returns a tuple with the CurrentPassword field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentPassword

`func (o *AuthRecoveryCodesRegenerateRequest) SetCurrentPassword(v string)`

SetCurrentPassword sets CurrentPassword field to given value.

### HasCurrentPassword

`func (o *AuthRecoveryCodesRegenerateRequest) HasCurrentPassword() bool`

HasCurrentPassword returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


