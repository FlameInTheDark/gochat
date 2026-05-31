# UserDMCallJoinResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Call** | Pointer to [**MqmsgDMCallSummary**](MqmsgDMCallSummary.md) |  | [optional] 
**Region** | Pointer to **string** |  | [optional] 
**SfuToken** | Pointer to **string** |  | [optional] 
**SfuUrl** | Pointer to **string** |  | [optional] 

## Methods

### NewUserDMCallJoinResponse

`func NewUserDMCallJoinResponse() *UserDMCallJoinResponse`

NewUserDMCallJoinResponse instantiates a new UserDMCallJoinResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUserDMCallJoinResponseWithDefaults

`func NewUserDMCallJoinResponseWithDefaults() *UserDMCallJoinResponse`

NewUserDMCallJoinResponseWithDefaults instantiates a new UserDMCallJoinResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCall

`func (o *UserDMCallJoinResponse) GetCall() MqmsgDMCallSummary`

GetCall returns the Call field if non-nil, zero value otherwise.

### GetCallOk

`func (o *UserDMCallJoinResponse) GetCallOk() (*MqmsgDMCallSummary, bool)`

GetCallOk returns a tuple with the Call field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCall

`func (o *UserDMCallJoinResponse) SetCall(v MqmsgDMCallSummary)`

SetCall sets Call field to given value.

### HasCall

`func (o *UserDMCallJoinResponse) HasCall() bool`

HasCall returns a boolean if a field has been set.

### GetRegion

`func (o *UserDMCallJoinResponse) GetRegion() string`

GetRegion returns the Region field if non-nil, zero value otherwise.

### GetRegionOk

`func (o *UserDMCallJoinResponse) GetRegionOk() (*string, bool)`

GetRegionOk returns a tuple with the Region field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRegion

`func (o *UserDMCallJoinResponse) SetRegion(v string)`

SetRegion sets Region field to given value.

### HasRegion

`func (o *UserDMCallJoinResponse) HasRegion() bool`

HasRegion returns a boolean if a field has been set.

### GetSfuToken

`func (o *UserDMCallJoinResponse) GetSfuToken() string`

GetSfuToken returns the SfuToken field if non-nil, zero value otherwise.

### GetSfuTokenOk

`func (o *UserDMCallJoinResponse) GetSfuTokenOk() (*string, bool)`

GetSfuTokenOk returns a tuple with the SfuToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSfuToken

`func (o *UserDMCallJoinResponse) SetSfuToken(v string)`

SetSfuToken sets SfuToken field to given value.

### HasSfuToken

`func (o *UserDMCallJoinResponse) HasSfuToken() bool`

HasSfuToken returns a boolean if a field has been set.

### GetSfuUrl

`func (o *UserDMCallJoinResponse) GetSfuUrl() string`

GetSfuUrl returns the SfuUrl field if non-nil, zero value otherwise.

### GetSfuUrlOk

`func (o *UserDMCallJoinResponse) GetSfuUrlOk() (*string, bool)`

GetSfuUrlOk returns a tuple with the SfuUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSfuUrl

`func (o *UserDMCallJoinResponse) SetSfuUrl(v string)`

SetSfuUrl sets SfuUrl field to given value.

### HasSfuUrl

`func (o *UserDMCallJoinResponse) HasSfuUrl() bool`

HasSfuUrl returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


