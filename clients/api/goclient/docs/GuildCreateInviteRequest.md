# GuildCreateInviteRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ExpiresInSec** | Pointer to **int32** | ExpiresInSec is a TTL in seconds; 0 means unlimited invite If omitted, server uses default (7 days) | [optional] 

## Methods

### NewGuildCreateInviteRequest

`func NewGuildCreateInviteRequest() *GuildCreateInviteRequest`

NewGuildCreateInviteRequest instantiates a new GuildCreateInviteRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGuildCreateInviteRequestWithDefaults

`func NewGuildCreateInviteRequestWithDefaults() *GuildCreateInviteRequest`

NewGuildCreateInviteRequestWithDefaults instantiates a new GuildCreateInviteRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetExpiresInSec

`func (o *GuildCreateInviteRequest) GetExpiresInSec() int32`

GetExpiresInSec returns the ExpiresInSec field if non-nil, zero value otherwise.

### GetExpiresInSecOk

`func (o *GuildCreateInviteRequest) GetExpiresInSecOk() (*int32, bool)`

GetExpiresInSecOk returns a tuple with the ExpiresInSec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpiresInSec

`func (o *GuildCreateInviteRequest) SetExpiresInSec(v int32)`

SetExpiresInSec sets ExpiresInSec field to given value.

### HasExpiresInSec

`func (o *GuildCreateInviteRequest) HasExpiresInSec() bool`

HasExpiresInSec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


