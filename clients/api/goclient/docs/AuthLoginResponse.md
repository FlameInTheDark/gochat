# AuthLoginResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**RefreshToken** | Pointer to **string** | Refresh token. Used to refresh authentication token. | [optional] 
**Token** | Pointer to **string** | Authentication token | [optional] 

## Methods

### NewAuthLoginResponse

`func NewAuthLoginResponse() *AuthLoginResponse`

NewAuthLoginResponse instantiates a new AuthLoginResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAuthLoginResponseWithDefaults

`func NewAuthLoginResponseWithDefaults() *AuthLoginResponse`

NewAuthLoginResponseWithDefaults instantiates a new AuthLoginResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRefreshToken

`func (o *AuthLoginResponse) GetRefreshToken() string`

GetRefreshToken returns the RefreshToken field if non-nil, zero value otherwise.

### GetRefreshTokenOk

`func (o *AuthLoginResponse) GetRefreshTokenOk() (*string, bool)`

GetRefreshTokenOk returns a tuple with the RefreshToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRefreshToken

`func (o *AuthLoginResponse) SetRefreshToken(v string)`

SetRefreshToken sets RefreshToken field to given value.

### HasRefreshToken

`func (o *AuthLoginResponse) HasRefreshToken() bool`

HasRefreshToken returns a boolean if a field has been set.

### GetToken

`func (o *AuthLoginResponse) GetToken() string`

GetToken returns the Token field if non-nil, zero value otherwise.

### GetTokenOk

`func (o *AuthLoginResponse) GetTokenOk() (*string, bool)`

GetTokenOk returns a tuple with the Token field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToken

`func (o *AuthLoginResponse) SetToken(v string)`

SetToken sets Token field to given value.

### HasToken

`func (o *AuthLoginResponse) HasToken() bool`

HasToken returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


