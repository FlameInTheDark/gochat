# SearchMessageSearchResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Messages** | Pointer to [**[]DtoMessage**](DtoMessage.md) |  | [optional] 
**Pages** | Pointer to **int32** |  | [optional] 

## Methods

### NewSearchMessageSearchResponse

`func NewSearchMessageSearchResponse() *SearchMessageSearchResponse`

NewSearchMessageSearchResponse instantiates a new SearchMessageSearchResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSearchMessageSearchResponseWithDefaults

`func NewSearchMessageSearchResponseWithDefaults() *SearchMessageSearchResponse`

NewSearchMessageSearchResponseWithDefaults instantiates a new SearchMessageSearchResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMessages

`func (o *SearchMessageSearchResponse) GetMessages() []DtoMessage`

GetMessages returns the Messages field if non-nil, zero value otherwise.

### GetMessagesOk

`func (o *SearchMessageSearchResponse) GetMessagesOk() (*[]DtoMessage, bool)`

GetMessagesOk returns a tuple with the Messages field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessages

`func (o *SearchMessageSearchResponse) SetMessages(v []DtoMessage)`

SetMessages sets Messages field to given value.

### HasMessages

`func (o *SearchMessageSearchResponse) HasMessages() bool`

HasMessages returns a boolean if a field has been set.

### GetPages

`func (o *SearchMessageSearchResponse) GetPages() int32`

GetPages returns the Pages field if non-nil, zero value otherwise.

### GetPagesOk

`func (o *SearchMessageSearchResponse) GetPagesOk() (*int32, bool)`

GetPagesOk returns a tuple with the Pages field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPages

`func (o *SearchMessageSearchResponse) SetPages(v int32)`

SetPages sets Pages field to given value.

### HasPages

`func (o *SearchMessageSearchResponse) HasPages() bool`

HasPages returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


