# UserCreateDMManyRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ChannelId** | Pointer to **int32** |  | [optional] 
**RecipientsId** | Pointer to **[]int32** |  | [optional] 

## Methods

### NewUserCreateDMManyRequest

`func NewUserCreateDMManyRequest() *UserCreateDMManyRequest`

NewUserCreateDMManyRequest instantiates a new UserCreateDMManyRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUserCreateDMManyRequestWithDefaults

`func NewUserCreateDMManyRequestWithDefaults() *UserCreateDMManyRequest`

NewUserCreateDMManyRequestWithDefaults instantiates a new UserCreateDMManyRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetChannelId

`func (o *UserCreateDMManyRequest) GetChannelId() int32`

GetChannelId returns the ChannelId field if non-nil, zero value otherwise.

### GetChannelIdOk

`func (o *UserCreateDMManyRequest) GetChannelIdOk() (*int32, bool)`

GetChannelIdOk returns a tuple with the ChannelId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChannelId

`func (o *UserCreateDMManyRequest) SetChannelId(v int32)`

SetChannelId sets ChannelId field to given value.

### HasChannelId

`func (o *UserCreateDMManyRequest) HasChannelId() bool`

HasChannelId returns a boolean if a field has been set.

### GetRecipientsId

`func (o *UserCreateDMManyRequest) GetRecipientsId() []int32`

GetRecipientsId returns the RecipientsId field if non-nil, zero value otherwise.

### GetRecipientsIdOk

`func (o *UserCreateDMManyRequest) GetRecipientsIdOk() (*[]int32, bool)`

GetRecipientsIdOk returns a tuple with the RecipientsId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecipientsId

`func (o *UserCreateDMManyRequest) SetRecipientsId(v []int32)`

SetRecipientsId sets RecipientsId field to given value.

### HasRecipientsId

`func (o *UserCreateDMManyRequest) HasRecipientsId() bool`

HasRecipientsId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


