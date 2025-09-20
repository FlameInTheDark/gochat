# AuthLoginRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Email** | Pointer to **string** | User Email | [optional] 
**Password** | Pointer to **string** | User password | [optional] 

## Methods

### NewAuthLoginRequest

`func NewAuthLoginRequest() *AuthLoginRequest`

NewAuthLoginRequest instantiates a new AuthLoginRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAuthLoginRequestWithDefaults

`func NewAuthLoginRequestWithDefaults() *AuthLoginRequest`

NewAuthLoginRequestWithDefaults instantiates a new AuthLoginRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEmail

`func (o *AuthLoginRequest) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *AuthLoginRequest) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *AuthLoginRequest) SetEmail(v string)`

SetEmail sets Email field to given value.

### HasEmail

`func (o *AuthLoginRequest) HasEmail() bool`

HasEmail returns a boolean if a field has been set.

### GetPassword

`func (o *AuthLoginRequest) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *AuthLoginRequest) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *AuthLoginRequest) SetPassword(v string)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *AuthLoginRequest) HasPassword() bool`

HasPassword returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


