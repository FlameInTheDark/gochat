# AuthLoginEmailVerifyRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ChallengeId** | Pointer to **string** |  | [optional] 
**Code** | Pointer to **string** |  | [optional] 

## Methods

### NewAuthLoginEmailVerifyRequest

`func NewAuthLoginEmailVerifyRequest() *AuthLoginEmailVerifyRequest`

NewAuthLoginEmailVerifyRequest instantiates a new AuthLoginEmailVerifyRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAuthLoginEmailVerifyRequestWithDefaults

`func NewAuthLoginEmailVerifyRequestWithDefaults() *AuthLoginEmailVerifyRequest`

NewAuthLoginEmailVerifyRequestWithDefaults instantiates a new AuthLoginEmailVerifyRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetChallengeId

`func (o *AuthLoginEmailVerifyRequest) GetChallengeId() string`

GetChallengeId returns the ChallengeId field if non-nil, zero value otherwise.

### GetChallengeIdOk

`func (o *AuthLoginEmailVerifyRequest) GetChallengeIdOk() (*string, bool)`

GetChallengeIdOk returns a tuple with the ChallengeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChallengeId

`func (o *AuthLoginEmailVerifyRequest) SetChallengeId(v string)`

SetChallengeId sets ChallengeId field to given value.

### HasChallengeId

`func (o *AuthLoginEmailVerifyRequest) HasChallengeId() bool`

HasChallengeId returns a boolean if a field has been set.

### GetCode

`func (o *AuthLoginEmailVerifyRequest) GetCode() string`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *AuthLoginEmailVerifyRequest) GetCodeOk() (*string, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *AuthLoginEmailVerifyRequest) SetCode(v string)`

SetCode sets Code field to given value.

### HasCode

`func (o *AuthLoginEmailVerifyRequest) HasCode() bool`

HasCode returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


