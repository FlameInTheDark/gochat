# GuildBanMemberRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Reason** | Pointer to **string** |  | [optional] 

## Methods

### NewGuildBanMemberRequest

`func NewGuildBanMemberRequest() *GuildBanMemberRequest`

NewGuildBanMemberRequest instantiates a new GuildBanMemberRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGuildBanMemberRequestWithDefaults

`func NewGuildBanMemberRequestWithDefaults() *GuildBanMemberRequest`

NewGuildBanMemberRequestWithDefaults instantiates a new GuildBanMemberRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetReason

`func (o *GuildBanMemberRequest) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *GuildBanMemberRequest) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *GuildBanMemberRequest) SetReason(v string)`

SetReason sets Reason field to given value.

### HasReason

`func (o *GuildBanMemberRequest) HasReason() bool`

HasReason returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


