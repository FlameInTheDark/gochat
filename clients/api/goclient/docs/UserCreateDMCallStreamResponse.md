# UserCreateDMCallStreamResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Stream** | Pointer to [**UserVoiceStreamSummary**](UserVoiceStreamSummary.md) |  | [optional] 
**StreamId** | Pointer to **int32** |  | [optional] 
**StreamToken** | Pointer to **string** |  | [optional] 
**StreamUrl** | Pointer to **string** |  | [optional] 

## Methods

### NewUserCreateDMCallStreamResponse

`func NewUserCreateDMCallStreamResponse() *UserCreateDMCallStreamResponse`

NewUserCreateDMCallStreamResponse instantiates a new UserCreateDMCallStreamResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUserCreateDMCallStreamResponseWithDefaults

`func NewUserCreateDMCallStreamResponseWithDefaults() *UserCreateDMCallStreamResponse`

NewUserCreateDMCallStreamResponseWithDefaults instantiates a new UserCreateDMCallStreamResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStream

`func (o *UserCreateDMCallStreamResponse) GetStream() UserVoiceStreamSummary`

GetStream returns the Stream field if non-nil, zero value otherwise.

### GetStreamOk

`func (o *UserCreateDMCallStreamResponse) GetStreamOk() (*UserVoiceStreamSummary, bool)`

GetStreamOk returns a tuple with the Stream field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStream

`func (o *UserCreateDMCallStreamResponse) SetStream(v UserVoiceStreamSummary)`

SetStream sets Stream field to given value.

### HasStream

`func (o *UserCreateDMCallStreamResponse) HasStream() bool`

HasStream returns a boolean if a field has been set.

### GetStreamId

`func (o *UserCreateDMCallStreamResponse) GetStreamId() int32`

GetStreamId returns the StreamId field if non-nil, zero value otherwise.

### GetStreamIdOk

`func (o *UserCreateDMCallStreamResponse) GetStreamIdOk() (*int32, bool)`

GetStreamIdOk returns a tuple with the StreamId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStreamId

`func (o *UserCreateDMCallStreamResponse) SetStreamId(v int32)`

SetStreamId sets StreamId field to given value.

### HasStreamId

`func (o *UserCreateDMCallStreamResponse) HasStreamId() bool`

HasStreamId returns a boolean if a field has been set.

### GetStreamToken

`func (o *UserCreateDMCallStreamResponse) GetStreamToken() string`

GetStreamToken returns the StreamToken field if non-nil, zero value otherwise.

### GetStreamTokenOk

`func (o *UserCreateDMCallStreamResponse) GetStreamTokenOk() (*string, bool)`

GetStreamTokenOk returns a tuple with the StreamToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStreamToken

`func (o *UserCreateDMCallStreamResponse) SetStreamToken(v string)`

SetStreamToken sets StreamToken field to given value.

### HasStreamToken

`func (o *UserCreateDMCallStreamResponse) HasStreamToken() bool`

HasStreamToken returns a boolean if a field has been set.

### GetStreamUrl

`func (o *UserCreateDMCallStreamResponse) GetStreamUrl() string`

GetStreamUrl returns the StreamUrl field if non-nil, zero value otherwise.

### GetStreamUrlOk

`func (o *UserCreateDMCallStreamResponse) GetStreamUrlOk() (*string, bool)`

GetStreamUrlOk returns a tuple with the StreamUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStreamUrl

`func (o *UserCreateDMCallStreamResponse) SetStreamUrl(v string)`

SetStreamUrl sets StreamUrl field to given value.

### HasStreamUrl

`func (o *UserCreateDMCallStreamResponse) HasStreamUrl() bool`

HasStreamUrl returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


