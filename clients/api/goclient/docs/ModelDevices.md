# ModelDevices

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AudioInputDevice** | Pointer to **string** |  | [optional] 
**AudioInputLevel** | Pointer to **float32** |  | [optional] 
**AudioInputThreshold** | Pointer to **float32** |  | [optional] 
**AudioOutputDevice** | Pointer to **string** |  | [optional] 
**AudioOutputLevel** | Pointer to **float32** |  | [optional] 
**AutoGainControl** | Pointer to **bool** |  | [optional] 
**DenoiserType** | Pointer to **string** |  | [optional] 
**EchoCancellation** | Pointer to **bool** |  | [optional] 
**NoiseSuppression** | Pointer to **bool** |  | [optional] 
**VideoDevice** | Pointer to **string** |  | [optional] 

## Methods

### NewModelDevices

`func NewModelDevices() *ModelDevices`

NewModelDevices instantiates a new ModelDevices object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewModelDevicesWithDefaults

`func NewModelDevicesWithDefaults() *ModelDevices`

NewModelDevicesWithDefaults instantiates a new ModelDevices object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAudioInputDevice

`func (o *ModelDevices) GetAudioInputDevice() string`

GetAudioInputDevice returns the AudioInputDevice field if non-nil, zero value otherwise.

### GetAudioInputDeviceOk

`func (o *ModelDevices) GetAudioInputDeviceOk() (*string, bool)`

GetAudioInputDeviceOk returns a tuple with the AudioInputDevice field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAudioInputDevice

`func (o *ModelDevices) SetAudioInputDevice(v string)`

SetAudioInputDevice sets AudioInputDevice field to given value.

### HasAudioInputDevice

`func (o *ModelDevices) HasAudioInputDevice() bool`

HasAudioInputDevice returns a boolean if a field has been set.

### GetAudioInputLevel

`func (o *ModelDevices) GetAudioInputLevel() float32`

GetAudioInputLevel returns the AudioInputLevel field if non-nil, zero value otherwise.

### GetAudioInputLevelOk

`func (o *ModelDevices) GetAudioInputLevelOk() (*float32, bool)`

GetAudioInputLevelOk returns a tuple with the AudioInputLevel field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAudioInputLevel

`func (o *ModelDevices) SetAudioInputLevel(v float32)`

SetAudioInputLevel sets AudioInputLevel field to given value.

### HasAudioInputLevel

`func (o *ModelDevices) HasAudioInputLevel() bool`

HasAudioInputLevel returns a boolean if a field has been set.

### GetAudioInputThreshold

`func (o *ModelDevices) GetAudioInputThreshold() float32`

GetAudioInputThreshold returns the AudioInputThreshold field if non-nil, zero value otherwise.

### GetAudioInputThresholdOk

`func (o *ModelDevices) GetAudioInputThresholdOk() (*float32, bool)`

GetAudioInputThresholdOk returns a tuple with the AudioInputThreshold field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAudioInputThreshold

`func (o *ModelDevices) SetAudioInputThreshold(v float32)`

SetAudioInputThreshold sets AudioInputThreshold field to given value.

### HasAudioInputThreshold

`func (o *ModelDevices) HasAudioInputThreshold() bool`

HasAudioInputThreshold returns a boolean if a field has been set.

### GetAudioOutputDevice

`func (o *ModelDevices) GetAudioOutputDevice() string`

GetAudioOutputDevice returns the AudioOutputDevice field if non-nil, zero value otherwise.

### GetAudioOutputDeviceOk

`func (o *ModelDevices) GetAudioOutputDeviceOk() (*string, bool)`

GetAudioOutputDeviceOk returns a tuple with the AudioOutputDevice field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAudioOutputDevice

`func (o *ModelDevices) SetAudioOutputDevice(v string)`

SetAudioOutputDevice sets AudioOutputDevice field to given value.

### HasAudioOutputDevice

`func (o *ModelDevices) HasAudioOutputDevice() bool`

HasAudioOutputDevice returns a boolean if a field has been set.

### GetAudioOutputLevel

`func (o *ModelDevices) GetAudioOutputLevel() float32`

GetAudioOutputLevel returns the AudioOutputLevel field if non-nil, zero value otherwise.

### GetAudioOutputLevelOk

`func (o *ModelDevices) GetAudioOutputLevelOk() (*float32, bool)`

GetAudioOutputLevelOk returns a tuple with the AudioOutputLevel field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAudioOutputLevel

`func (o *ModelDevices) SetAudioOutputLevel(v float32)`

SetAudioOutputLevel sets AudioOutputLevel field to given value.

### HasAudioOutputLevel

`func (o *ModelDevices) HasAudioOutputLevel() bool`

HasAudioOutputLevel returns a boolean if a field has been set.

### GetAutoGainControl

`func (o *ModelDevices) GetAutoGainControl() bool`

GetAutoGainControl returns the AutoGainControl field if non-nil, zero value otherwise.

### GetAutoGainControlOk

`func (o *ModelDevices) GetAutoGainControlOk() (*bool, bool)`

GetAutoGainControlOk returns a tuple with the AutoGainControl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoGainControl

`func (o *ModelDevices) SetAutoGainControl(v bool)`

SetAutoGainControl sets AutoGainControl field to given value.

### HasAutoGainControl

`func (o *ModelDevices) HasAutoGainControl() bool`

HasAutoGainControl returns a boolean if a field has been set.

### GetDenoiserType

`func (o *ModelDevices) GetDenoiserType() string`

GetDenoiserType returns the DenoiserType field if non-nil, zero value otherwise.

### GetDenoiserTypeOk

`func (o *ModelDevices) GetDenoiserTypeOk() (*string, bool)`

GetDenoiserTypeOk returns a tuple with the DenoiserType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDenoiserType

`func (o *ModelDevices) SetDenoiserType(v string)`

SetDenoiserType sets DenoiserType field to given value.

### HasDenoiserType

`func (o *ModelDevices) HasDenoiserType() bool`

HasDenoiserType returns a boolean if a field has been set.

### GetEchoCancellation

`func (o *ModelDevices) GetEchoCancellation() bool`

GetEchoCancellation returns the EchoCancellation field if non-nil, zero value otherwise.

### GetEchoCancellationOk

`func (o *ModelDevices) GetEchoCancellationOk() (*bool, bool)`

GetEchoCancellationOk returns a tuple with the EchoCancellation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEchoCancellation

`func (o *ModelDevices) SetEchoCancellation(v bool)`

SetEchoCancellation sets EchoCancellation field to given value.

### HasEchoCancellation

`func (o *ModelDevices) HasEchoCancellation() bool`

HasEchoCancellation returns a boolean if a field has been set.

### GetNoiseSuppression

`func (o *ModelDevices) GetNoiseSuppression() bool`

GetNoiseSuppression returns the NoiseSuppression field if non-nil, zero value otherwise.

### GetNoiseSuppressionOk

`func (o *ModelDevices) GetNoiseSuppressionOk() (*bool, bool)`

GetNoiseSuppressionOk returns a tuple with the NoiseSuppression field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNoiseSuppression

`func (o *ModelDevices) SetNoiseSuppression(v bool)`

SetNoiseSuppression sets NoiseSuppression field to given value.

### HasNoiseSuppression

`func (o *ModelDevices) HasNoiseSuppression() bool`

HasNoiseSuppression returns a boolean if a field has been set.

### GetVideoDevice

`func (o *ModelDevices) GetVideoDevice() string`

GetVideoDevice returns the VideoDevice field if non-nil, zero value otherwise.

### GetVideoDeviceOk

`func (o *ModelDevices) GetVideoDeviceOk() (*string, bool)`

GetVideoDeviceOk returns a tuple with the VideoDevice field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVideoDevice

`func (o *ModelDevices) SetVideoDevice(v string)`

SetVideoDevice sets VideoDevice field to given value.

### HasVideoDevice

`func (o *ModelDevices) HasVideoDevice() bool`

HasVideoDevice returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


