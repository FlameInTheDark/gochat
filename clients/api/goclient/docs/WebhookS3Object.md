# WebhookS3Object

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ContentType** | Pointer to **string** |  | [optional] 
**ETag** | Pointer to **string** |  | [optional] 
**Key** | Pointer to **string** |  | [optional] 
**Sequencer** | Pointer to **string** |  | [optional] 
**Size** | Pointer to **int32** |  | [optional] 
**UserMetadata** | Pointer to **map[string]string** |  | [optional] 

## Methods

### NewWebhookS3Object

`func NewWebhookS3Object() *WebhookS3Object`

NewWebhookS3Object instantiates a new WebhookS3Object object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWebhookS3ObjectWithDefaults

`func NewWebhookS3ObjectWithDefaults() *WebhookS3Object`

NewWebhookS3ObjectWithDefaults instantiates a new WebhookS3Object object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetContentType

`func (o *WebhookS3Object) GetContentType() string`

GetContentType returns the ContentType field if non-nil, zero value otherwise.

### GetContentTypeOk

`func (o *WebhookS3Object) GetContentTypeOk() (*string, bool)`

GetContentTypeOk returns a tuple with the ContentType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContentType

`func (o *WebhookS3Object) SetContentType(v string)`

SetContentType sets ContentType field to given value.

### HasContentType

`func (o *WebhookS3Object) HasContentType() bool`

HasContentType returns a boolean if a field has been set.

### GetETag

`func (o *WebhookS3Object) GetETag() string`

GetETag returns the ETag field if non-nil, zero value otherwise.

### GetETagOk

`func (o *WebhookS3Object) GetETagOk() (*string, bool)`

GetETagOk returns a tuple with the ETag field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetETag

`func (o *WebhookS3Object) SetETag(v string)`

SetETag sets ETag field to given value.

### HasETag

`func (o *WebhookS3Object) HasETag() bool`

HasETag returns a boolean if a field has been set.

### GetKey

`func (o *WebhookS3Object) GetKey() string`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *WebhookS3Object) GetKeyOk() (*string, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *WebhookS3Object) SetKey(v string)`

SetKey sets Key field to given value.

### HasKey

`func (o *WebhookS3Object) HasKey() bool`

HasKey returns a boolean if a field has been set.

### GetSequencer

`func (o *WebhookS3Object) GetSequencer() string`

GetSequencer returns the Sequencer field if non-nil, zero value otherwise.

### GetSequencerOk

`func (o *WebhookS3Object) GetSequencerOk() (*string, bool)`

GetSequencerOk returns a tuple with the Sequencer field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSequencer

`func (o *WebhookS3Object) SetSequencer(v string)`

SetSequencer sets Sequencer field to given value.

### HasSequencer

`func (o *WebhookS3Object) HasSequencer() bool`

HasSequencer returns a boolean if a field has been set.

### GetSize

`func (o *WebhookS3Object) GetSize() int32`

GetSize returns the Size field if non-nil, zero value otherwise.

### GetSizeOk

`func (o *WebhookS3Object) GetSizeOk() (*int32, bool)`

GetSizeOk returns a tuple with the Size field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSize

`func (o *WebhookS3Object) SetSize(v int32)`

SetSize sets Size field to given value.

### HasSize

`func (o *WebhookS3Object) HasSize() bool`

HasSize returns a boolean if a field has been set.

### GetUserMetadata

`func (o *WebhookS3Object) GetUserMetadata() map[string]string`

GetUserMetadata returns the UserMetadata field if non-nil, zero value otherwise.

### GetUserMetadataOk

`func (o *WebhookS3Object) GetUserMetadataOk() (*map[string]string, bool)`

GetUserMetadataOk returns a tuple with the UserMetadata field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserMetadata

`func (o *WebhookS3Object) SetUserMetadata(v map[string]string)`

SetUserMetadata sets UserMetadata field to given value.

### HasUserMetadata

`func (o *WebhookS3Object) HasUserMetadata() bool`

HasUserMetadata returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


