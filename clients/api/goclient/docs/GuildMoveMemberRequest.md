# GuildMoveMemberRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ChannelId** | Pointer to **int32** |  | [optional] 
**From** | Pointer to **int32** |  | [optional] 
**UserId** | Pointer to **int32** |  | [optional] 

## Methods

### NewGuildMoveMemberRequest

`func NewGuildMoveMemberRequest() *GuildMoveMemberRequest`

NewGuildMoveMemberRequest instantiates a new GuildMoveMemberRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGuildMoveMemberRequestWithDefaults

`func NewGuildMoveMemberRequestWithDefaults() *GuildMoveMemberRequest`

NewGuildMoveMemberRequestWithDefaults instantiates a new GuildMoveMemberRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetChannelId

`func (o *GuildMoveMemberRequest) GetChannelId() int32`

GetChannelId returns the ChannelId field if non-nil, zero value otherwise.

### GetChannelIdOk

`func (o *GuildMoveMemberRequest) GetChannelIdOk() (*int32, bool)`

GetChannelIdOk returns a tuple with the ChannelId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChannelId

`func (o *GuildMoveMemberRequest) SetChannelId(v int32)`

SetChannelId sets ChannelId field to given value.

### HasChannelId

`func (o *GuildMoveMemberRequest) HasChannelId() bool`

HasChannelId returns a boolean if a field has been set.

### GetFrom

`func (o *GuildMoveMemberRequest) GetFrom() int32`

GetFrom returns the From field if non-nil, zero value otherwise.

### GetFromOk

`func (o *GuildMoveMemberRequest) GetFromOk() (*int32, bool)`

GetFromOk returns a tuple with the From field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFrom

`func (o *GuildMoveMemberRequest) SetFrom(v int32)`

SetFrom sets From field to given value.

### HasFrom

`func (o *GuildMoveMemberRequest) HasFrom() bool`

HasFrom returns a boolean if a field has been set.

### GetUserId

`func (o *GuildMoveMemberRequest) GetUserId() int32`

GetUserId returns the UserId field if non-nil, zero value otherwise.

### GetUserIdOk

`func (o *GuildMoveMemberRequest) GetUserIdOk() (*int32, bool)`

GetUserIdOk returns a tuple with the UserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserId

`func (o *GuildMoveMemberRequest) SetUserId(v int32)`

SetUserId sets UserId field to given value.

### HasUserId

`func (o *GuildMoveMemberRequest) HasUserId() bool`

HasUserId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


