# ModelStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CustomStatusText** | Pointer to **string** |  | [optional] 
**Status** | Pointer to **string** |  | [optional] 

## Methods

### NewModelStatus

`func NewModelStatus() *ModelStatus`

NewModelStatus instantiates a new ModelStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewModelStatusWithDefaults

`func NewModelStatusWithDefaults() *ModelStatus`

NewModelStatusWithDefaults instantiates a new ModelStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCustomStatusText

`func (o *ModelStatus) GetCustomStatusText() string`

GetCustomStatusText returns the CustomStatusText field if non-nil, zero value otherwise.

### GetCustomStatusTextOk

`func (o *ModelStatus) GetCustomStatusTextOk() (*string, bool)`

GetCustomStatusTextOk returns a tuple with the CustomStatusText field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCustomStatusText

`func (o *ModelStatus) SetCustomStatusText(v string)`

SetCustomStatusText sets CustomStatusText field to given value.

### HasCustomStatusText

`func (o *ModelStatus) HasCustomStatusText() bool`

HasCustomStatusText returns a boolean if a field has been set.

### GetStatus

`func (o *ModelStatus) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ModelStatus) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ModelStatus) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *ModelStatus) HasStatus() bool`

HasStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


