# AuthConfirmationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Discriminator** | Pointer to **string** |  | [optional] 
**Id** | Pointer to **int32** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**Password** | Pointer to **string** |  | [optional] 
**Token** | Pointer to **string** |  | [optional] 

## Methods

### NewAuthConfirmationRequest

`func NewAuthConfirmationRequest() *AuthConfirmationRequest`

NewAuthConfirmationRequest instantiates a new AuthConfirmationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAuthConfirmationRequestWithDefaults

`func NewAuthConfirmationRequestWithDefaults() *AuthConfirmationRequest`

NewAuthConfirmationRequestWithDefaults instantiates a new AuthConfirmationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDiscriminator

`func (o *AuthConfirmationRequest) GetDiscriminator() string`

GetDiscriminator returns the Discriminator field if non-nil, zero value otherwise.

### GetDiscriminatorOk

`func (o *AuthConfirmationRequest) GetDiscriminatorOk() (*string, bool)`

GetDiscriminatorOk returns a tuple with the Discriminator field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDiscriminator

`func (o *AuthConfirmationRequest) SetDiscriminator(v string)`

SetDiscriminator sets Discriminator field to given value.

### HasDiscriminator

`func (o *AuthConfirmationRequest) HasDiscriminator() bool`

HasDiscriminator returns a boolean if a field has been set.

### GetId

`func (o *AuthConfirmationRequest) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *AuthConfirmationRequest) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *AuthConfirmationRequest) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *AuthConfirmationRequest) HasId() bool`

HasId returns a boolean if a field has been set.

### GetName

`func (o *AuthConfirmationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *AuthConfirmationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *AuthConfirmationRequest) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *AuthConfirmationRequest) HasName() bool`

HasName returns a boolean if a field has been set.

### GetPassword

`func (o *AuthConfirmationRequest) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *AuthConfirmationRequest) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *AuthConfirmationRequest) SetPassword(v string)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *AuthConfirmationRequest) HasPassword() bool`

HasPassword returns a boolean if a field has been set.

### GetToken

`func (o *AuthConfirmationRequest) GetToken() string`

GetToken returns the Token field if non-nil, zero value otherwise.

### GetTokenOk

`func (o *AuthConfirmationRequest) GetTokenOk() (*string, bool)`

GetTokenOk returns a tuple with the Token field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToken

`func (o *AuthConfirmationRequest) SetToken(v string)`

SetToken sets Token field to given value.

### HasToken

`func (o *AuthConfirmationRequest) HasToken() bool`

HasToken returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


