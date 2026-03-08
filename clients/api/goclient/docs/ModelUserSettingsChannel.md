# ModelUserSettingsChannel

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ChannelId** | Pointer to **int32** |  | [optional] 
**Notifications** | Pointer to [**ModelUserSettingsNotifications**](ModelUserSettingsNotifications.md) |  | [optional] 

## Methods

### NewModelUserSettingsChannel

`func NewModelUserSettingsChannel() *ModelUserSettingsChannel`

NewModelUserSettingsChannel instantiates a new ModelUserSettingsChannel object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewModelUserSettingsChannelWithDefaults

`func NewModelUserSettingsChannelWithDefaults() *ModelUserSettingsChannel`

NewModelUserSettingsChannelWithDefaults instantiates a new ModelUserSettingsChannel object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetChannelId

`func (o *ModelUserSettingsChannel) GetChannelId() int32`

GetChannelId returns the ChannelId field if non-nil, zero value otherwise.

### GetChannelIdOk

`func (o *ModelUserSettingsChannel) GetChannelIdOk() (*int32, bool)`

GetChannelIdOk returns a tuple with the ChannelId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChannelId

`func (o *ModelUserSettingsChannel) SetChannelId(v int32)`

SetChannelId sets ChannelId field to given value.

### HasChannelId

`func (o *ModelUserSettingsChannel) HasChannelId() bool`

HasChannelId returns a boolean if a field has been set.

### GetNotifications

`func (o *ModelUserSettingsChannel) GetNotifications() ModelUserSettingsNotifications`

GetNotifications returns the Notifications field if non-nil, zero value otherwise.

### GetNotificationsOk

`func (o *ModelUserSettingsChannel) GetNotificationsOk() (*ModelUserSettingsNotifications, bool)`

GetNotificationsOk returns a tuple with the Notifications field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNotifications

`func (o *ModelUserSettingsChannel) SetNotifications(v ModelUserSettingsNotifications)`

SetNotifications sets Notifications field to given value.

### HasNotifications

`func (o *ModelUserSettingsChannel) HasNotifications() bool`

HasNotifications returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


