# WebhookS3Event

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**EventName** | Pointer to **string** |  | [optional] 
**Key** | Pointer to **string** |  | [optional] 
**Records** | Pointer to [**[]WebhookS3EventRecord**](WebhookS3EventRecord.md) |  | [optional] 

## Methods

### NewWebhookS3Event

`func NewWebhookS3Event() *WebhookS3Event`

NewWebhookS3Event instantiates a new WebhookS3Event object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWebhookS3EventWithDefaults

`func NewWebhookS3EventWithDefaults() *WebhookS3Event`

NewWebhookS3EventWithDefaults instantiates a new WebhookS3Event object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEventName

`func (o *WebhookS3Event) GetEventName() string`

GetEventName returns the EventName field if non-nil, zero value otherwise.

### GetEventNameOk

`func (o *WebhookS3Event) GetEventNameOk() (*string, bool)`

GetEventNameOk returns a tuple with the EventName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEventName

`func (o *WebhookS3Event) SetEventName(v string)`

SetEventName sets EventName field to given value.

### HasEventName

`func (o *WebhookS3Event) HasEventName() bool`

HasEventName returns a boolean if a field has been set.

### GetKey

`func (o *WebhookS3Event) GetKey() string`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *WebhookS3Event) GetKeyOk() (*string, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *WebhookS3Event) SetKey(v string)`

SetKey sets Key field to given value.

### HasKey

`func (o *WebhookS3Event) HasKey() bool`

HasKey returns a boolean if a field has been set.

### GetRecords

`func (o *WebhookS3Event) GetRecords() []WebhookS3EventRecord`

GetRecords returns the Records field if non-nil, zero value otherwise.

### GetRecordsOk

`func (o *WebhookS3Event) GetRecordsOk() (*[]WebhookS3EventRecord, bool)`

GetRecordsOk returns a tuple with the Records field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecords

`func (o *WebhookS3Event) SetRecords(v []WebhookS3EventRecord)`

SetRecords sets Records field to given value.

### HasRecords

`func (o *WebhookS3Event) HasRecords() bool`

HasRecords returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


