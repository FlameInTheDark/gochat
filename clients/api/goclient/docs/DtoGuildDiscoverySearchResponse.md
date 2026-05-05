# DtoGuildDiscoverySearchResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Guilds** | Pointer to [**[]DtoGuildDiscovery**](DtoGuildDiscovery.md) |  | [optional] 
**Pages** | Pointer to **int32** |  | [optional] 

## Methods

### NewDtoGuildDiscoverySearchResponse

`func NewDtoGuildDiscoverySearchResponse() *DtoGuildDiscoverySearchResponse`

NewDtoGuildDiscoverySearchResponse instantiates a new DtoGuildDiscoverySearchResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDtoGuildDiscoverySearchResponseWithDefaults

`func NewDtoGuildDiscoverySearchResponseWithDefaults() *DtoGuildDiscoverySearchResponse`

NewDtoGuildDiscoverySearchResponseWithDefaults instantiates a new DtoGuildDiscoverySearchResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGuilds

`func (o *DtoGuildDiscoverySearchResponse) GetGuilds() []DtoGuildDiscovery`

GetGuilds returns the Guilds field if non-nil, zero value otherwise.

### GetGuildsOk

`func (o *DtoGuildDiscoverySearchResponse) GetGuildsOk() (*[]DtoGuildDiscovery, bool)`

GetGuildsOk returns a tuple with the Guilds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGuilds

`func (o *DtoGuildDiscoverySearchResponse) SetGuilds(v []DtoGuildDiscovery)`

SetGuilds sets Guilds field to given value.

### HasGuilds

`func (o *DtoGuildDiscoverySearchResponse) HasGuilds() bool`

HasGuilds returns a boolean if a field has been set.

### GetPages

`func (o *DtoGuildDiscoverySearchResponse) GetPages() int32`

GetPages returns the Pages field if non-nil, zero value otherwise.

### GetPagesOk

`func (o *DtoGuildDiscoverySearchResponse) GetPagesOk() (*int32, bool)`

GetPagesOk returns a tuple with the Pages field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPages

`func (o *DtoGuildDiscoverySearchResponse) SetPages(v int32)`

SetPages sets Pages field to given value.

### HasPages

`func (o *DtoGuildDiscoverySearchResponse) HasPages() bool`

HasPages returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


