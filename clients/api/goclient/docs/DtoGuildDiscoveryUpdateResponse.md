# DtoGuildDiscoveryUpdateResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Guild** | Pointer to [**DtoGuildDiscovery**](DtoGuildDiscovery.md) |  | [optional] 
**IndexingPending** | Pointer to **bool** |  | [optional] 

## Methods

### NewDtoGuildDiscoveryUpdateResponse

`func NewDtoGuildDiscoveryUpdateResponse() *DtoGuildDiscoveryUpdateResponse`

NewDtoGuildDiscoveryUpdateResponse instantiates a new DtoGuildDiscoveryUpdateResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDtoGuildDiscoveryUpdateResponseWithDefaults

`func NewDtoGuildDiscoveryUpdateResponseWithDefaults() *DtoGuildDiscoveryUpdateResponse`

NewDtoGuildDiscoveryUpdateResponseWithDefaults instantiates a new DtoGuildDiscoveryUpdateResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGuild

`func (o *DtoGuildDiscoveryUpdateResponse) GetGuild() DtoGuildDiscovery`

GetGuild returns the Guild field if non-nil, zero value otherwise.

### GetGuildOk

`func (o *DtoGuildDiscoveryUpdateResponse) GetGuildOk() (*DtoGuildDiscovery, bool)`

GetGuildOk returns a tuple with the Guild field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGuild

`func (o *DtoGuildDiscoveryUpdateResponse) SetGuild(v DtoGuildDiscovery)`

SetGuild sets Guild field to given value.

### HasGuild

`func (o *DtoGuildDiscoveryUpdateResponse) HasGuild() bool`

HasGuild returns a boolean if a field has been set.

### GetIndexingPending

`func (o *DtoGuildDiscoveryUpdateResponse) GetIndexingPending() bool`

GetIndexingPending returns the IndexingPending field if non-nil, zero value otherwise.

### GetIndexingPendingOk

`func (o *DtoGuildDiscoveryUpdateResponse) GetIndexingPendingOk() (*bool, bool)`

GetIndexingPendingOk returns a tuple with the IndexingPending field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIndexingPending

`func (o *DtoGuildDiscoveryUpdateResponse) SetIndexingPending(v bool)`

SetIndexingPending sets IndexingPending field to given value.

### HasIndexingPending

`func (o *DtoGuildDiscoveryUpdateResponse) HasIndexingPending() bool`

HasIndexingPending returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


