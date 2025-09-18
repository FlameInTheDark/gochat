# WebhookS3Bucket

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Arn** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**OwnerIdentity** | Pointer to [**WebhookS3Identity**](WebhookS3Identity.md) |  | [optional] 

## Methods

### NewWebhookS3Bucket

`func NewWebhookS3Bucket() *WebhookS3Bucket`

NewWebhookS3Bucket instantiates a new WebhookS3Bucket object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWebhookS3BucketWithDefaults

`func NewWebhookS3BucketWithDefaults() *WebhookS3Bucket`

NewWebhookS3BucketWithDefaults instantiates a new WebhookS3Bucket object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetArn

`func (o *WebhookS3Bucket) GetArn() string`

GetArn returns the Arn field if non-nil, zero value otherwise.

### GetArnOk

`func (o *WebhookS3Bucket) GetArnOk() (*string, bool)`

GetArnOk returns a tuple with the Arn field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArn

`func (o *WebhookS3Bucket) SetArn(v string)`

SetArn sets Arn field to given value.

### HasArn

`func (o *WebhookS3Bucket) HasArn() bool`

HasArn returns a boolean if a field has been set.

### GetName

`func (o *WebhookS3Bucket) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WebhookS3Bucket) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WebhookS3Bucket) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *WebhookS3Bucket) HasName() bool`

HasName returns a boolean if a field has been set.

### GetOwnerIdentity

`func (o *WebhookS3Bucket) GetOwnerIdentity() WebhookS3Identity`

GetOwnerIdentity returns the OwnerIdentity field if non-nil, zero value otherwise.

### GetOwnerIdentityOk

`func (o *WebhookS3Bucket) GetOwnerIdentityOk() (*WebhookS3Identity, bool)`

GetOwnerIdentityOk returns a tuple with the OwnerIdentity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOwnerIdentity

`func (o *WebhookS3Bucket) SetOwnerIdentity(v WebhookS3Identity)`

SetOwnerIdentity sets OwnerIdentity field to given value.

### HasOwnerIdentity

`func (o *WebhookS3Bucket) HasOwnerIdentity() bool`

HasOwnerIdentity returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


