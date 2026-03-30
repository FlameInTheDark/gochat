# AuthLoginRecoveryCodeRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ChallengeId** | Pointer to **string** |  | [optional] 
**Code** | Pointer to **string** |  | [optional] 

## Methods

### NewAuthLoginRecoveryCodeRequest

`func NewAuthLoginRecoveryCodeRequest() *AuthLoginRecoveryCodeRequest`

NewAuthLoginRecoveryCodeRequest instantiates a new AuthLoginRecoveryCodeRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAuthLoginRecoveryCodeRequestWithDefaults

`func NewAuthLoginRecoveryCodeRequestWithDefaults() *AuthLoginRecoveryCodeRequest`

NewAuthLoginRecoveryCodeRequestWithDefaults instantiates a new AuthLoginRecoveryCodeRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetChallengeId

`func (o *AuthLoginRecoveryCodeRequest) GetChallengeId() string`

GetChallengeId returns the ChallengeId field if non-nil, zero value otherwise.

### GetChallengeIdOk

`func (o *AuthLoginRecoveryCodeRequest) GetChallengeIdOk() (*string, bool)`

GetChallengeIdOk returns a tuple with the ChallengeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChallengeId

`func (o *AuthLoginRecoveryCodeRequest) SetChallengeId(v string)`

SetChallengeId sets ChallengeId field to given value.

### HasChallengeId

`func (o *AuthLoginRecoveryCodeRequest) HasChallengeId() bool`

HasChallengeId returns a boolean if a field has been set.

### GetCode

`func (o *AuthLoginRecoveryCodeRequest) GetCode() string`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *AuthLoginRecoveryCodeRequest) GetCodeOk() (*string, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *AuthLoginRecoveryCodeRequest) SetCode(v string)`

SetCode sets Code field to given value.

### HasCode

`func (o *AuthLoginRecoveryCodeRequest) HasCode() bool`

HasCode returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


