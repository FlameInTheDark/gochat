# WebhookS3EventRecord

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AwsRegion** | Pointer to **string** |  | [optional] 
**EventName** | Pointer to **string** |  | [optional] 
**EventSource** | Pointer to **string** |  | [optional] 
**EventTime** | Pointer to **string** |  | [optional] 
**EventVersion** | Pointer to **string** |  | [optional] 
**RequestParameters** | Pointer to [**WebhookS3RequestParameters**](WebhookS3RequestParameters.md) |  | [optional] 
**ResponseElements** | Pointer to **map[string]string** |  | [optional] 
**S3** | Pointer to [**WebhookS3Element**](WebhookS3Element.md) |  | [optional] 
**Source** | Pointer to [**WebhookS3Source**](WebhookS3Source.md) |  | [optional] 
**UserIdentity** | Pointer to [**WebhookS3Identity**](WebhookS3Identity.md) |  | [optional] 

## Methods

### NewWebhookS3EventRecord

`func NewWebhookS3EventRecord() *WebhookS3EventRecord`

NewWebhookS3EventRecord instantiates a new WebhookS3EventRecord object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWebhookS3EventRecordWithDefaults

`func NewWebhookS3EventRecordWithDefaults() *WebhookS3EventRecord`

NewWebhookS3EventRecordWithDefaults instantiates a new WebhookS3EventRecord object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAwsRegion

`func (o *WebhookS3EventRecord) GetAwsRegion() string`

GetAwsRegion returns the AwsRegion field if non-nil, zero value otherwise.

### GetAwsRegionOk

`func (o *WebhookS3EventRecord) GetAwsRegionOk() (*string, bool)`

GetAwsRegionOk returns a tuple with the AwsRegion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAwsRegion

`func (o *WebhookS3EventRecord) SetAwsRegion(v string)`

SetAwsRegion sets AwsRegion field to given value.

### HasAwsRegion

`func (o *WebhookS3EventRecord) HasAwsRegion() bool`

HasAwsRegion returns a boolean if a field has been set.

### GetEventName

`func (o *WebhookS3EventRecord) GetEventName() string`

GetEventName returns the EventName field if non-nil, zero value otherwise.

### GetEventNameOk

`func (o *WebhookS3EventRecord) GetEventNameOk() (*string, bool)`

GetEventNameOk returns a tuple with the EventName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEventName

`func (o *WebhookS3EventRecord) SetEventName(v string)`

SetEventName sets EventName field to given value.

### HasEventName

`func (o *WebhookS3EventRecord) HasEventName() bool`

HasEventName returns a boolean if a field has been set.

### GetEventSource

`func (o *WebhookS3EventRecord) GetEventSource() string`

GetEventSource returns the EventSource field if non-nil, zero value otherwise.

### GetEventSourceOk

`func (o *WebhookS3EventRecord) GetEventSourceOk() (*string, bool)`

GetEventSourceOk returns a tuple with the EventSource field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEventSource

`func (o *WebhookS3EventRecord) SetEventSource(v string)`

SetEventSource sets EventSource field to given value.

### HasEventSource

`func (o *WebhookS3EventRecord) HasEventSource() bool`

HasEventSource returns a boolean if a field has been set.

### GetEventTime

`func (o *WebhookS3EventRecord) GetEventTime() string`

GetEventTime returns the EventTime field if non-nil, zero value otherwise.

### GetEventTimeOk

`func (o *WebhookS3EventRecord) GetEventTimeOk() (*string, bool)`

GetEventTimeOk returns a tuple with the EventTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEventTime

`func (o *WebhookS3EventRecord) SetEventTime(v string)`

SetEventTime sets EventTime field to given value.

### HasEventTime

`func (o *WebhookS3EventRecord) HasEventTime() bool`

HasEventTime returns a boolean if a field has been set.

### GetEventVersion

`func (o *WebhookS3EventRecord) GetEventVersion() string`

GetEventVersion returns the EventVersion field if non-nil, zero value otherwise.

### GetEventVersionOk

`func (o *WebhookS3EventRecord) GetEventVersionOk() (*string, bool)`

GetEventVersionOk returns a tuple with the EventVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEventVersion

`func (o *WebhookS3EventRecord) SetEventVersion(v string)`

SetEventVersion sets EventVersion field to given value.

### HasEventVersion

`func (o *WebhookS3EventRecord) HasEventVersion() bool`

HasEventVersion returns a boolean if a field has been set.

### GetRequestParameters

`func (o *WebhookS3EventRecord) GetRequestParameters() WebhookS3RequestParameters`

GetRequestParameters returns the RequestParameters field if non-nil, zero value otherwise.

### GetRequestParametersOk

`func (o *WebhookS3EventRecord) GetRequestParametersOk() (*WebhookS3RequestParameters, bool)`

GetRequestParametersOk returns a tuple with the RequestParameters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequestParameters

`func (o *WebhookS3EventRecord) SetRequestParameters(v WebhookS3RequestParameters)`

SetRequestParameters sets RequestParameters field to given value.

### HasRequestParameters

`func (o *WebhookS3EventRecord) HasRequestParameters() bool`

HasRequestParameters returns a boolean if a field has been set.

### GetResponseElements

`func (o *WebhookS3EventRecord) GetResponseElements() map[string]string`

GetResponseElements returns the ResponseElements field if non-nil, zero value otherwise.

### GetResponseElementsOk

`func (o *WebhookS3EventRecord) GetResponseElementsOk() (*map[string]string, bool)`

GetResponseElementsOk returns a tuple with the ResponseElements field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResponseElements

`func (o *WebhookS3EventRecord) SetResponseElements(v map[string]string)`

SetResponseElements sets ResponseElements field to given value.

### HasResponseElements

`func (o *WebhookS3EventRecord) HasResponseElements() bool`

HasResponseElements returns a boolean if a field has been set.

### GetS3

`func (o *WebhookS3EventRecord) GetS3() WebhookS3Element`

GetS3 returns the S3 field if non-nil, zero value otherwise.

### GetS3Ok

`func (o *WebhookS3EventRecord) GetS3Ok() (*WebhookS3Element, bool)`

GetS3Ok returns a tuple with the S3 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetS3

`func (o *WebhookS3EventRecord) SetS3(v WebhookS3Element)`

SetS3 sets S3 field to given value.

### HasS3

`func (o *WebhookS3EventRecord) HasS3() bool`

HasS3 returns a boolean if a field has been set.

### GetSource

`func (o *WebhookS3EventRecord) GetSource() WebhookS3Source`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *WebhookS3EventRecord) GetSourceOk() (*WebhookS3Source, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *WebhookS3EventRecord) SetSource(v WebhookS3Source)`

SetSource sets Source field to given value.

### HasSource

`func (o *WebhookS3EventRecord) HasSource() bool`

HasSource returns a boolean if a field has been set.

### GetUserIdentity

`func (o *WebhookS3EventRecord) GetUserIdentity() WebhookS3Identity`

GetUserIdentity returns the UserIdentity field if non-nil, zero value otherwise.

### GetUserIdentityOk

`func (o *WebhookS3EventRecord) GetUserIdentityOk() (*WebhookS3Identity, bool)`

GetUserIdentityOk returns a tuple with the UserIdentity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserIdentity

`func (o *WebhookS3EventRecord) SetUserIdentity(v WebhookS3Identity)`

SetUserIdentity sets UserIdentity field to given value.

### HasUserIdentity

`func (o *WebhookS3EventRecord) HasUserIdentity() bool`

HasUserIdentity returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


