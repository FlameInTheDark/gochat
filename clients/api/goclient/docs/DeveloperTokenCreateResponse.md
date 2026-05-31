# DeveloperTokenCreateResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Token** | Pointer to **string** |  | [optional] 
**TokenData** | Pointer to [**ModelBotToken**](ModelBotToken.md) |  | [optional] 

## Methods

### NewDeveloperTokenCreateResponse

`func NewDeveloperTokenCreateResponse() *DeveloperTokenCreateResponse`

NewDeveloperTokenCreateResponse instantiates a new DeveloperTokenCreateResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeveloperTokenCreateResponseWithDefaults

`func NewDeveloperTokenCreateResponseWithDefaults() *DeveloperTokenCreateResponse`

NewDeveloperTokenCreateResponseWithDefaults instantiates a new DeveloperTokenCreateResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetToken

`func (o *DeveloperTokenCreateResponse) GetToken() string`

GetToken returns the Token field if non-nil, zero value otherwise.

### GetTokenOk

`func (o *DeveloperTokenCreateResponse) GetTokenOk() (*string, bool)`

GetTokenOk returns a tuple with the Token field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToken

`func (o *DeveloperTokenCreateResponse) SetToken(v string)`

SetToken sets Token field to given value.

### HasToken

`func (o *DeveloperTokenCreateResponse) HasToken() bool`

HasToken returns a boolean if a field has been set.

### GetTokenData

`func (o *DeveloperTokenCreateResponse) GetTokenData() ModelBotToken`

GetTokenData returns the TokenData field if non-nil, zero value otherwise.

### GetTokenDataOk

`func (o *DeveloperTokenCreateResponse) GetTokenDataOk() (*ModelBotToken, bool)`

GetTokenDataOk returns a tuple with the TokenData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTokenData

`func (o *DeveloperTokenCreateResponse) SetTokenData(v ModelBotToken)`

SetTokenData sets TokenData field to given value.

### HasTokenData

`func (o *DeveloperTokenCreateResponse) HasTokenData() bool`

HasTokenData returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


