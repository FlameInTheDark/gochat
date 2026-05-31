# MqmsgDMCallSummary

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CallId** | Pointer to **int32** |  | [optional] 
**CallerId** | Pointer to **int32** |  | [optional] 
**ChannelId** | Pointer to **int32** |  | [optional] 
**Dismissed** | Pointer to **bool** |  | [optional] 
**Participants** | Pointer to **map[string]int32** |  | [optional] 
**RecipientId** | Pointer to **int32** |  | [optional] 
**Region** | Pointer to **string** |  | [optional] 
**SoloSince** | Pointer to **int32** |  | [optional] 
**StartedAt** | Pointer to **int32** |  | [optional] 

## Methods

### NewMqmsgDMCallSummary

`func NewMqmsgDMCallSummary() *MqmsgDMCallSummary`

NewMqmsgDMCallSummary instantiates a new MqmsgDMCallSummary object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewMqmsgDMCallSummaryWithDefaults

`func NewMqmsgDMCallSummaryWithDefaults() *MqmsgDMCallSummary`

NewMqmsgDMCallSummaryWithDefaults instantiates a new MqmsgDMCallSummary object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCallId

`func (o *MqmsgDMCallSummary) GetCallId() int32`

GetCallId returns the CallId field if non-nil, zero value otherwise.

### GetCallIdOk

`func (o *MqmsgDMCallSummary) GetCallIdOk() (*int32, bool)`

GetCallIdOk returns a tuple with the CallId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCallId

`func (o *MqmsgDMCallSummary) SetCallId(v int32)`

SetCallId sets CallId field to given value.

### HasCallId

`func (o *MqmsgDMCallSummary) HasCallId() bool`

HasCallId returns a boolean if a field has been set.

### GetCallerId

`func (o *MqmsgDMCallSummary) GetCallerId() int32`

GetCallerId returns the CallerId field if non-nil, zero value otherwise.

### GetCallerIdOk

`func (o *MqmsgDMCallSummary) GetCallerIdOk() (*int32, bool)`

GetCallerIdOk returns a tuple with the CallerId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCallerId

`func (o *MqmsgDMCallSummary) SetCallerId(v int32)`

SetCallerId sets CallerId field to given value.

### HasCallerId

`func (o *MqmsgDMCallSummary) HasCallerId() bool`

HasCallerId returns a boolean if a field has been set.

### GetChannelId

`func (o *MqmsgDMCallSummary) GetChannelId() int32`

GetChannelId returns the ChannelId field if non-nil, zero value otherwise.

### GetChannelIdOk

`func (o *MqmsgDMCallSummary) GetChannelIdOk() (*int32, bool)`

GetChannelIdOk returns a tuple with the ChannelId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChannelId

`func (o *MqmsgDMCallSummary) SetChannelId(v int32)`

SetChannelId sets ChannelId field to given value.

### HasChannelId

`func (o *MqmsgDMCallSummary) HasChannelId() bool`

HasChannelId returns a boolean if a field has been set.

### GetDismissed

`func (o *MqmsgDMCallSummary) GetDismissed() bool`

GetDismissed returns the Dismissed field if non-nil, zero value otherwise.

### GetDismissedOk

`func (o *MqmsgDMCallSummary) GetDismissedOk() (*bool, bool)`

GetDismissedOk returns a tuple with the Dismissed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDismissed

`func (o *MqmsgDMCallSummary) SetDismissed(v bool)`

SetDismissed sets Dismissed field to given value.

### HasDismissed

`func (o *MqmsgDMCallSummary) HasDismissed() bool`

HasDismissed returns a boolean if a field has been set.

### GetParticipants

`func (o *MqmsgDMCallSummary) GetParticipants() map[string]int32`

GetParticipants returns the Participants field if non-nil, zero value otherwise.

### GetParticipantsOk

`func (o *MqmsgDMCallSummary) GetParticipantsOk() (*map[string]int32, bool)`

GetParticipantsOk returns a tuple with the Participants field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParticipants

`func (o *MqmsgDMCallSummary) SetParticipants(v map[string]int32)`

SetParticipants sets Participants field to given value.

### HasParticipants

`func (o *MqmsgDMCallSummary) HasParticipants() bool`

HasParticipants returns a boolean if a field has been set.

### GetRecipientId

`func (o *MqmsgDMCallSummary) GetRecipientId() int32`

GetRecipientId returns the RecipientId field if non-nil, zero value otherwise.

### GetRecipientIdOk

`func (o *MqmsgDMCallSummary) GetRecipientIdOk() (*int32, bool)`

GetRecipientIdOk returns a tuple with the RecipientId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecipientId

`func (o *MqmsgDMCallSummary) SetRecipientId(v int32)`

SetRecipientId sets RecipientId field to given value.

### HasRecipientId

`func (o *MqmsgDMCallSummary) HasRecipientId() bool`

HasRecipientId returns a boolean if a field has been set.

### GetRegion

`func (o *MqmsgDMCallSummary) GetRegion() string`

GetRegion returns the Region field if non-nil, zero value otherwise.

### GetRegionOk

`func (o *MqmsgDMCallSummary) GetRegionOk() (*string, bool)`

GetRegionOk returns a tuple with the Region field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRegion

`func (o *MqmsgDMCallSummary) SetRegion(v string)`

SetRegion sets Region field to given value.

### HasRegion

`func (o *MqmsgDMCallSummary) HasRegion() bool`

HasRegion returns a boolean if a field has been set.

### GetSoloSince

`func (o *MqmsgDMCallSummary) GetSoloSince() int32`

GetSoloSince returns the SoloSince field if non-nil, zero value otherwise.

### GetSoloSinceOk

`func (o *MqmsgDMCallSummary) GetSoloSinceOk() (*int32, bool)`

GetSoloSinceOk returns a tuple with the SoloSince field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSoloSince

`func (o *MqmsgDMCallSummary) SetSoloSince(v int32)`

SetSoloSince sets SoloSince field to given value.

### HasSoloSince

`func (o *MqmsgDMCallSummary) HasSoloSince() bool`

HasSoloSince returns a boolean if a field has been set.

### GetStartedAt

`func (o *MqmsgDMCallSummary) GetStartedAt() int32`

GetStartedAt returns the StartedAt field if non-nil, zero value otherwise.

### GetStartedAtOk

`func (o *MqmsgDMCallSummary) GetStartedAtOk() (*int32, bool)`

GetStartedAtOk returns a tuple with the StartedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStartedAt

`func (o *MqmsgDMCallSummary) SetStartedAt(v int32)`

SetStartedAt sets StartedAt field to given value.

### HasStartedAt

`func (o *MqmsgDMCallSummary) HasStartedAt() bool`

HasStartedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


