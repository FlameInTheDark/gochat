# DeveloperGrantCreateResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Grant** | Pointer to [**ModelBotInstallGrant**](ModelBotInstallGrant.md) |  | [optional] 
**Token** | Pointer to **string** |  | [optional] 

## Methods

### NewDeveloperGrantCreateResponse

`func NewDeveloperGrantCreateResponse() *DeveloperGrantCreateResponse`

NewDeveloperGrantCreateResponse instantiates a new DeveloperGrantCreateResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDeveloperGrantCreateResponseWithDefaults

`func NewDeveloperGrantCreateResponseWithDefaults() *DeveloperGrantCreateResponse`

NewDeveloperGrantCreateResponseWithDefaults instantiates a new DeveloperGrantCreateResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGrant

`func (o *DeveloperGrantCreateResponse) GetGrant() ModelBotInstallGrant`

GetGrant returns the Grant field if non-nil, zero value otherwise.

### GetGrantOk

`func (o *DeveloperGrantCreateResponse) GetGrantOk() (*ModelBotInstallGrant, bool)`

GetGrantOk returns a tuple with the Grant field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGrant

`func (o *DeveloperGrantCreateResponse) SetGrant(v ModelBotInstallGrant)`

SetGrant sets Grant field to given value.

### HasGrant

`func (o *DeveloperGrantCreateResponse) HasGrant() bool`

HasGrant returns a boolean if a field has been set.

### GetToken

`func (o *DeveloperGrantCreateResponse) GetToken() string`

GetToken returns the Token field if non-nil, zero value otherwise.

### GetTokenOk

`func (o *DeveloperGrantCreateResponse) GetTokenOk() (*string, bool)`

GetTokenOk returns a tuple with the Token field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToken

`func (o *DeveloperGrantCreateResponse) SetToken(v string)`

SetToken sets Token field to given value.

### HasToken

`func (o *DeveloperGrantCreateResponse) HasToken() bool`

HasToken returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


