# AuthPasswordResetRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **int32** |  | [optional] 
**Password** | Pointer to **string** |  | [optional] 
**Token** | Pointer to **string** |  | [optional] 

## Methods

### NewAuthPasswordResetRequest

`func NewAuthPasswordResetRequest() *AuthPasswordResetRequest`

NewAuthPasswordResetRequest instantiates a new AuthPasswordResetRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAuthPasswordResetRequestWithDefaults

`func NewAuthPasswordResetRequestWithDefaults() *AuthPasswordResetRequest`

NewAuthPasswordResetRequestWithDefaults instantiates a new AuthPasswordResetRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *AuthPasswordResetRequest) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AuthPasswordResetRequest) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AuthPasswordResetRequest) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *AuthPasswordResetRequest) HasId() bool`

HasId returns a boolean if a field has been set.

### GetPassword

`func (o *AuthPasswordResetRequest) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *AuthPasswordResetRequest) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *AuthPasswordResetRequest) SetPassword(v string)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *AuthPasswordResetRequest) HasPassword() bool`

HasPassword returns a boolean if a field has been set.

### GetToken

`func (o *AuthPasswordResetRequest) GetToken() string`

GetToken returns the Token field if non-nil, zero value otherwise.

### GetTokenOk

`func (o *AuthPasswordResetRequest) GetTokenOk() (*string, bool)`

GetTokenOk returns a tuple with the Token field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToken

`func (o *AuthPasswordResetRequest) SetToken(v string)`

SetToken sets Token field to given value.

### HasToken

`func (o *AuthPasswordResetRequest) HasToken() bool`

HasToken returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


