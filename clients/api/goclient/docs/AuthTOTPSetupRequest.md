# AuthTOTPSetupRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CurrentPassword** | Pointer to **string** |  | [optional] 

## Methods

### NewAuthTOTPSetupRequest

`func NewAuthTOTPSetupRequest() *AuthTOTPSetupRequest`

NewAuthTOTPSetupRequest instantiates a new AuthTOTPSetupRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAuthTOTPSetupRequestWithDefaults

`func NewAuthTOTPSetupRequestWithDefaults() *AuthTOTPSetupRequest`

NewAuthTOTPSetupRequestWithDefaults instantiates a new AuthTOTPSetupRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCurrentPassword

`func (o *AuthTOTPSetupRequest) GetCurrentPassword() string`

GetCurrentPassword returns the CurrentPassword field if non-nil, zero value otherwise.

### GetCurrentPasswordOk

`func (o *AuthTOTPSetupRequest) GetCurrentPasswordOk() (*string, bool)`

GetCurrentPasswordOk returns a tuple with the CurrentPassword field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentPassword

`func (o *AuthTOTPSetupRequest) SetCurrentPassword(v string)`

SetCurrentPassword sets CurrentPassword field to given value.

### HasCurrentPassword

`func (o *AuthTOTPSetupRequest) HasCurrentPassword() bool`

HasCurrentPassword returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


