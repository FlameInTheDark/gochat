# AuthRegisterRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Email** | Pointer to **string** |  | [optional] 

## Methods

### NewAuthRegisterRequest

`func NewAuthRegisterRequest() *AuthRegisterRequest`

NewAuthRegisterRequest instantiates a new AuthRegisterRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAuthRegisterRequestWithDefaults

`func NewAuthRegisterRequestWithDefaults() *AuthRegisterRequest`

NewAuthRegisterRequestWithDefaults instantiates a new AuthRegisterRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEmail

`func (o *AuthRegisterRequest) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *AuthRegisterRequest) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *AuthRegisterRequest) SetEmail(v string)`

SetEmail sets Email field to given value.

### HasEmail

`func (o *AuthRegisterRequest) HasEmail() bool`

HasEmail returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


