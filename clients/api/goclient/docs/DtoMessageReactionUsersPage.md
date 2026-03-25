# DtoMessageReactionUsersPage

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Items** | Pointer to [**[]DtoUser**](DtoUser.md) |  | [optional] 
**NextAfter** | Pointer to **int32** |  | [optional] 

## Methods

### NewDtoMessageReactionUsersPage

`func NewDtoMessageReactionUsersPage() *DtoMessageReactionUsersPage`

NewDtoMessageReactionUsersPage instantiates a new DtoMessageReactionUsersPage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDtoMessageReactionUsersPageWithDefaults

`func NewDtoMessageReactionUsersPageWithDefaults() *DtoMessageReactionUsersPage`

NewDtoMessageReactionUsersPageWithDefaults instantiates a new DtoMessageReactionUsersPage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetItems

`func (o *DtoMessageReactionUsersPage) GetItems() []DtoUser`

GetItems returns the Items field if non-nil, zero value otherwise.

### GetItemsOk

`func (o *DtoMessageReactionUsersPage) GetItemsOk() (*[]DtoUser, bool)`

GetItemsOk returns a tuple with the Items field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetItems

`func (o *DtoMessageReactionUsersPage) SetItems(v []DtoUser)`

SetItems sets Items field to given value.

### HasItems

`func (o *DtoMessageReactionUsersPage) HasItems() bool`

HasItems returns a boolean if a field has been set.

### GetNextAfter

`func (o *DtoMessageReactionUsersPage) GetNextAfter() int32`

GetNextAfter returns the NextAfter field if non-nil, zero value otherwise.

### GetNextAfterOk

`func (o *DtoMessageReactionUsersPage) GetNextAfterOk() (*int32, bool)`

GetNextAfterOk returns a tuple with the NextAfter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNextAfter

`func (o *DtoMessageReactionUsersPage) SetNextAfter(v int32)`

SetNextAfter sets NextAfter field to given value.

### HasNextAfter

`func (o *DtoMessageReactionUsersPage) HasNextAfter() bool`

HasNextAfter returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


