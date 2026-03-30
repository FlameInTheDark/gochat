# AuthLoginChallengeResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ChallengeId** | Pointer to **string** |  | [optional] 
**ExpiresAt** | Pointer to **string** |  | [optional] 
**Methods** | Pointer to **[]string** |  | [optional] 

## Methods

### NewAuthLoginChallengeResponse

`func NewAuthLoginChallengeResponse() *AuthLoginChallengeResponse`

NewAuthLoginChallengeResponse instantiates a new AuthLoginChallengeResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAuthLoginChallengeResponseWithDefaults

`func NewAuthLoginChallengeResponseWithDefaults() *AuthLoginChallengeResponse`

NewAuthLoginChallengeResponseWithDefaults instantiates a new AuthLoginChallengeResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetChallengeId

`func (o *AuthLoginChallengeResponse) GetChallengeId() string`

GetChallengeId returns the ChallengeId field if non-nil, zero value otherwise.

### GetChallengeIdOk

`func (o *AuthLoginChallengeResponse) GetChallengeIdOk() (*string, bool)`

GetChallengeIdOk returns a tuple with the ChallengeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChallengeId

`func (o *AuthLoginChallengeResponse) SetChallengeId(v string)`

SetChallengeId sets ChallengeId field to given value.

### HasChallengeId

`func (o *AuthLoginChallengeResponse) HasChallengeId() bool`

HasChallengeId returns a boolean if a field has been set.

### GetExpiresAt

`func (o *AuthLoginChallengeResponse) GetExpiresAt() string`

GetExpiresAt returns the ExpiresAt field if non-nil, zero value otherwise.

### GetExpiresAtOk

`func (o *AuthLoginChallengeResponse) GetExpiresAtOk() (*string, bool)`

GetExpiresAtOk returns a tuple with the ExpiresAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpiresAt

`func (o *AuthLoginChallengeResponse) SetExpiresAt(v string)`

SetExpiresAt sets ExpiresAt field to given value.

### HasExpiresAt

`func (o *AuthLoginChallengeResponse) HasExpiresAt() bool`

HasExpiresAt returns a boolean if a field has been set.

### GetMethods

`func (o *AuthLoginChallengeResponse) GetMethods() []string`

GetMethods returns the Methods field if non-nil, zero value otherwise.

### GetMethodsOk

`func (o *AuthLoginChallengeResponse) GetMethodsOk() (*[]string, bool)`

GetMethodsOk returns a tuple with the Methods field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMethods

`func (o *AuthLoginChallengeResponse) SetMethods(v []string)`

SetMethods sets Methods field to given value.

### HasMethods

`func (o *AuthLoginChallengeResponse) HasMethods() bool`

HasMethods returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


