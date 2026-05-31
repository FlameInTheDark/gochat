# DtoBotDiscoverySearchResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Bots** | Pointer to [**[]DtoBotDiscovery**](DtoBotDiscovery.md) |  | [optional] 
**Pages** | Pointer to **int32** |  | [optional] 

## Methods

### NewDtoBotDiscoverySearchResponse

`func NewDtoBotDiscoverySearchResponse() *DtoBotDiscoverySearchResponse`

NewDtoBotDiscoverySearchResponse instantiates a new DtoBotDiscoverySearchResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDtoBotDiscoverySearchResponseWithDefaults

`func NewDtoBotDiscoverySearchResponseWithDefaults() *DtoBotDiscoverySearchResponse`

NewDtoBotDiscoverySearchResponseWithDefaults instantiates a new DtoBotDiscoverySearchResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBots

`func (o *DtoBotDiscoverySearchResponse) GetBots() []DtoBotDiscovery`

GetBots returns the Bots field if non-nil, zero value otherwise.

### GetBotsOk

`func (o *DtoBotDiscoverySearchResponse) GetBotsOk() (*[]DtoBotDiscovery, bool)`

GetBotsOk returns a tuple with the Bots field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBots

`func (o *DtoBotDiscoverySearchResponse) SetBots(v []DtoBotDiscovery)`

SetBots sets Bots field to given value.

### HasBots

`func (o *DtoBotDiscoverySearchResponse) HasBots() bool`

HasBots returns a boolean if a field has been set.

### GetPages

`func (o *DtoBotDiscoverySearchResponse) GetPages() int32`

GetPages returns the Pages field if non-nil, zero value otherwise.

### GetPagesOk

`func (o *DtoBotDiscoverySearchResponse) GetPagesOk() (*int32, bool)`

GetPagesOk returns a tuple with the Pages field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPages

`func (o *DtoBotDiscoverySearchResponse) SetPages(v int32)`

SetPages sets Pages field to given value.

### HasPages

`func (o *DtoBotDiscoverySearchResponse) HasPages() bool`

HasPages returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


