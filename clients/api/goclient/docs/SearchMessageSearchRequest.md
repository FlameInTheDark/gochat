# SearchMessageSearchRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AuthorId** | Pointer to **int32** |  | [optional] 
**ChannelId** | Pointer to **int32** |  | [optional] 
**Content** | Pointer to **string** |  | [optional] 
**Has** | Pointer to **[]string** |  | [optional] 
**Mentions** | Pointer to **[]int32** |  | [optional] 
**Page** | Pointer to **int32** |  | [optional] 

## Methods

### NewSearchMessageSearchRequest

`func NewSearchMessageSearchRequest() *SearchMessageSearchRequest`

NewSearchMessageSearchRequest instantiates a new SearchMessageSearchRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSearchMessageSearchRequestWithDefaults

`func NewSearchMessageSearchRequestWithDefaults() *SearchMessageSearchRequest`

NewSearchMessageSearchRequestWithDefaults instantiates a new SearchMessageSearchRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAuthorId

`func (o *SearchMessageSearchRequest) GetAuthorId() int32`

GetAuthorId returns the AuthorId field if non-nil, zero value otherwise.

### GetAuthorIdOk

`func (o *SearchMessageSearchRequest) GetAuthorIdOk() (*int32, bool)`

GetAuthorIdOk returns a tuple with the AuthorId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthorId

`func (o *SearchMessageSearchRequest) SetAuthorId(v int32)`

SetAuthorId sets AuthorId field to given value.

### HasAuthorId

`func (o *SearchMessageSearchRequest) HasAuthorId() bool`

HasAuthorId returns a boolean if a field has been set.

### GetChannelId

`func (o *SearchMessageSearchRequest) GetChannelId() int32`

GetChannelId returns the ChannelId field if non-nil, zero value otherwise.

### GetChannelIdOk

`func (o *SearchMessageSearchRequest) GetChannelIdOk() (*int32, bool)`

GetChannelIdOk returns a tuple with the ChannelId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChannelId

`func (o *SearchMessageSearchRequest) SetChannelId(v int32)`

SetChannelId sets ChannelId field to given value.

### HasChannelId

`func (o *SearchMessageSearchRequest) HasChannelId() bool`

HasChannelId returns a boolean if a field has been set.

### GetContent

`func (o *SearchMessageSearchRequest) GetContent() string`

GetContent returns the Content field if non-nil, zero value otherwise.

### GetContentOk

`func (o *SearchMessageSearchRequest) GetContentOk() (*string, bool)`

GetContentOk returns a tuple with the Content field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContent

`func (o *SearchMessageSearchRequest) SetContent(v string)`

SetContent sets Content field to given value.

### HasContent

`func (o *SearchMessageSearchRequest) HasContent() bool`

HasContent returns a boolean if a field has been set.

### GetHas

`func (o *SearchMessageSearchRequest) GetHas() []string`

GetHas returns the Has field if non-nil, zero value otherwise.

### GetHasOk

`func (o *SearchMessageSearchRequest) GetHasOk() (*[]string, bool)`

GetHasOk returns a tuple with the Has field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHas

`func (o *SearchMessageSearchRequest) SetHas(v []string)`

SetHas sets Has field to given value.

### HasHas

`func (o *SearchMessageSearchRequest) HasHas() bool`

HasHas returns a boolean if a field has been set.

### GetMentions

`func (o *SearchMessageSearchRequest) GetMentions() []int32`

GetMentions returns the Mentions field if non-nil, zero value otherwise.

### GetMentionsOk

`func (o *SearchMessageSearchRequest) GetMentionsOk() (*[]int32, bool)`

GetMentionsOk returns a tuple with the Mentions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMentions

`func (o *SearchMessageSearchRequest) SetMentions(v []int32)`

SetMentions sets Mentions field to given value.

### HasMentions

`func (o *SearchMessageSearchRequest) HasMentions() bool`

HasMentions returns a boolean if a field has been set.

### GetPage

`func (o *SearchMessageSearchRequest) GetPage() int32`

GetPage returns the Page field if non-nil, zero value otherwise.

### GetPageOk

`func (o *SearchMessageSearchRequest) GetPageOk() (*int32, bool)`

GetPageOk returns a tuple with the Page field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPage

`func (o *SearchMessageSearchRequest) SetPage(v int32)`

SetPage sets Page field to given value.

### HasPage

`func (o *SearchMessageSearchRequest) HasPage() bool`

HasPage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


