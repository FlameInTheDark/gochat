# AuthTwoFactorStatusResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Enabled** | Pointer to **bool** |  | [optional] 
**FactorType** | Pointer to **string** |  | [optional] 
**RecoveryCodesRemaining** | Pointer to **int32** |  | [optional] 

## Methods

### NewAuthTwoFactorStatusResponse

`func NewAuthTwoFactorStatusResponse() *AuthTwoFactorStatusResponse`

NewAuthTwoFactorStatusResponse instantiates a new AuthTwoFactorStatusResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAuthTwoFactorStatusResponseWithDefaults

`func NewAuthTwoFactorStatusResponseWithDefaults() *AuthTwoFactorStatusResponse`

NewAuthTwoFactorStatusResponseWithDefaults instantiates a new AuthTwoFactorStatusResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnabled

`func (o *AuthTwoFactorStatusResponse) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *AuthTwoFactorStatusResponse) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *AuthTwoFactorStatusResponse) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *AuthTwoFactorStatusResponse) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### GetFactorType

`func (o *AuthTwoFactorStatusResponse) GetFactorType() string`

GetFactorType returns the FactorType field if non-nil, zero value otherwise.

### GetFactorTypeOk

`func (o *AuthTwoFactorStatusResponse) GetFactorTypeOk() (*string, bool)`

GetFactorTypeOk returns a tuple with the FactorType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFactorType

`func (o *AuthTwoFactorStatusResponse) SetFactorType(v string)`

SetFactorType sets FactorType field to given value.

### HasFactorType

`func (o *AuthTwoFactorStatusResponse) HasFactorType() bool`

HasFactorType returns a boolean if a field has been set.

### GetRecoveryCodesRemaining

`func (o *AuthTwoFactorStatusResponse) GetRecoveryCodesRemaining() int32`

GetRecoveryCodesRemaining returns the RecoveryCodesRemaining field if non-nil, zero value otherwise.

### GetRecoveryCodesRemainingOk

`func (o *AuthTwoFactorStatusResponse) GetRecoveryCodesRemainingOk() (*int32, bool)`

GetRecoveryCodesRemainingOk returns a tuple with the RecoveryCodesRemaining field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecoveryCodesRemaining

`func (o *AuthTwoFactorStatusResponse) SetRecoveryCodesRemaining(v int32)`

SetRecoveryCodesRemaining sets RecoveryCodesRemaining field to given value.

### HasRecoveryCodesRemaining

`func (o *AuthTwoFactorStatusResponse) HasRecoveryCodesRemaining() bool`

HasRecoveryCodesRemaining returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


