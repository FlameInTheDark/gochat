# DtoGuildBan

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Reason** | Pointer to **string** |  | [optional] 
**User** | Pointer to [**DtoUser**](DtoUser.md) |  | [optional] 

## Methods

### NewDtoGuildBan

`func NewDtoGuildBan() *DtoGuildBan`

NewDtoGuildBan instantiates a new DtoGuildBan object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDtoGuildBanWithDefaults

`func NewDtoGuildBanWithDefaults() *DtoGuildBan`

NewDtoGuildBanWithDefaults instantiates a new DtoGuildBan object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetReason

`func (o *DtoGuildBan) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *DtoGuildBan) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *DtoGuildBan) SetReason(v string)`

SetReason sets Reason field to given value.

### HasReason

`func (o *DtoGuildBan) HasReason() bool`

HasReason returns a boolean if a field has been set.

### GetUser

`func (o *DtoGuildBan) GetUser() DtoUser`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *DtoGuildBan) GetUserOk() (*DtoUser, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *DtoGuildBan) SetUser(v DtoUser)`

SetUser sets User field to given value.

### HasUser

`func (o *DtoGuildBan) HasUser() bool`

HasUser returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


