# WebhookS3Element

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Bucket** | Pointer to [**WebhookS3Bucket**](WebhookS3Bucket.md) |  | [optional] 
**ConfigurationId** | Pointer to **string** |  | [optional] 
**Object** | Pointer to [**WebhookS3Object**](WebhookS3Object.md) |  | [optional] 
**S3SchemaVersion** | Pointer to **string** |  | [optional] 

## Methods

### NewWebhookS3Element

`func NewWebhookS3Element() *WebhookS3Element`

NewWebhookS3Element instantiates a new WebhookS3Element object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWebhookS3ElementWithDefaults

`func NewWebhookS3ElementWithDefaults() *WebhookS3Element`

NewWebhookS3ElementWithDefaults instantiates a new WebhookS3Element object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBucket

`func (o *WebhookS3Element) GetBucket() WebhookS3Bucket`

GetBucket returns the Bucket field if non-nil, zero value otherwise.

### GetBucketOk

`func (o *WebhookS3Element) GetBucketOk() (*WebhookS3Bucket, bool)`

GetBucketOk returns a tuple with the Bucket field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBucket

`func (o *WebhookS3Element) SetBucket(v WebhookS3Bucket)`

SetBucket sets Bucket field to given value.

### HasBucket

`func (o *WebhookS3Element) HasBucket() bool`

HasBucket returns a boolean if a field has been set.

### GetConfigurationId

`func (o *WebhookS3Element) GetConfigurationId() string`

GetConfigurationId returns the ConfigurationId field if non-nil, zero value otherwise.

### GetConfigurationIdOk

`func (o *WebhookS3Element) GetConfigurationIdOk() (*string, bool)`

GetConfigurationIdOk returns a tuple with the ConfigurationId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigurationId

`func (o *WebhookS3Element) SetConfigurationId(v string)`

SetConfigurationId sets ConfigurationId field to given value.

### HasConfigurationId

`func (o *WebhookS3Element) HasConfigurationId() bool`

HasConfigurationId returns a boolean if a field has been set.

### GetObject

`func (o *WebhookS3Element) GetObject() WebhookS3Object`

GetObject returns the Object field if non-nil, zero value otherwise.

### GetObjectOk

`func (o *WebhookS3Element) GetObjectOk() (*WebhookS3Object, bool)`

GetObjectOk returns a tuple with the Object field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObject

`func (o *WebhookS3Element) SetObject(v WebhookS3Object)`

SetObject sets Object field to given value.

### HasObject

`func (o *WebhookS3Element) HasObject() bool`

HasObject returns a boolean if a field has been set.

### GetS3SchemaVersion

`func (o *WebhookS3Element) GetS3SchemaVersion() string`

GetS3SchemaVersion returns the S3SchemaVersion field if non-nil, zero value otherwise.

### GetS3SchemaVersionOk

`func (o *WebhookS3Element) GetS3SchemaVersionOk() (*string, bool)`

GetS3SchemaVersionOk returns a tuple with the S3SchemaVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetS3SchemaVersion

`func (o *WebhookS3Element) SetS3SchemaVersion(v string)`

SetS3SchemaVersion sets S3SchemaVersion field to given value.

### HasS3SchemaVersion

`func (o *WebhookS3Element) HasS3SchemaVersion() bool`

HasS3SchemaVersion returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


