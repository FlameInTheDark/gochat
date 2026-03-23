# ModelUserSettingsNotifications

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Muted** | Pointer to **bool** |  | [optional] 
**MutedUntil** | Pointer to **string** |  | [optional] 
**Notifications** | Pointer to [**ModelNotificationsType**](ModelNotificationsType.md) |  | [optional] 
**SuppressEveryoneMentions** | Pointer to **bool** |  | [optional] 
**SuppressHereMentions** | Pointer to **bool** |  | [optional] 
**SuppressRoleMentions** | Pointer to **bool** |  | [optional] 
**SuppressUserMentions** | Pointer to **bool** |  | [optional] 

## Methods

### NewModelUserSettingsNotifications

`func NewModelUserSettingsNotifications() *ModelUserSettingsNotifications`

NewModelUserSettingsNotifications instantiates a new ModelUserSettingsNotifications object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewModelUserSettingsNotificationsWithDefaults

`func NewModelUserSettingsNotificationsWithDefaults() *ModelUserSettingsNotifications`

NewModelUserSettingsNotificationsWithDefaults instantiates a new ModelUserSettingsNotifications object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMuted

`func (o *ModelUserSettingsNotifications) GetMuted() bool`

GetMuted returns the Muted field if non-nil, zero value otherwise.

### GetMutedOk

`func (o *ModelUserSettingsNotifications) GetMutedOk() (*bool, bool)`

GetMutedOk returns a tuple with the Muted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMuted

`func (o *ModelUserSettingsNotifications) SetMuted(v bool)`

SetMuted sets Muted field to given value.

### HasMuted

`func (o *ModelUserSettingsNotifications) HasMuted() bool`

HasMuted returns a boolean if a field has been set.

### GetMutedUntil

`func (o *ModelUserSettingsNotifications) GetMutedUntil() string`

GetMutedUntil returns the MutedUntil field if non-nil, zero value otherwise.

### GetMutedUntilOk

`func (o *ModelUserSettingsNotifications) GetMutedUntilOk() (*string, bool)`

GetMutedUntilOk returns a tuple with the MutedUntil field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMutedUntil

`func (o *ModelUserSettingsNotifications) SetMutedUntil(v string)`

SetMutedUntil sets MutedUntil field to given value.

### HasMutedUntil

`func (o *ModelUserSettingsNotifications) HasMutedUntil() bool`

HasMutedUntil returns a boolean if a field has been set.

### GetNotifications

`func (o *ModelUserSettingsNotifications) GetNotifications() ModelNotificationsType`

GetNotifications returns the Notifications field if non-nil, zero value otherwise.

### GetNotificationsOk

`func (o *ModelUserSettingsNotifications) GetNotificationsOk() (*ModelNotificationsType, bool)`

GetNotificationsOk returns a tuple with the Notifications field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNotifications

`func (o *ModelUserSettingsNotifications) SetNotifications(v ModelNotificationsType)`

SetNotifications sets Notifications field to given value.

### HasNotifications

`func (o *ModelUserSettingsNotifications) HasNotifications() bool`

HasNotifications returns a boolean if a field has been set.

### GetSuppressEveryoneMentions

`func (o *ModelUserSettingsNotifications) GetSuppressEveryoneMentions() bool`

GetSuppressEveryoneMentions returns the SuppressEveryoneMentions field if non-nil, zero value otherwise.

### GetSuppressEveryoneMentionsOk

`func (o *ModelUserSettingsNotifications) GetSuppressEveryoneMentionsOk() (*bool, bool)`

GetSuppressEveryoneMentionsOk returns a tuple with the SuppressEveryoneMentions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSuppressEveryoneMentions

`func (o *ModelUserSettingsNotifications) SetSuppressEveryoneMentions(v bool)`

SetSuppressEveryoneMentions sets SuppressEveryoneMentions field to given value.

### HasSuppressEveryoneMentions

`func (o *ModelUserSettingsNotifications) HasSuppressEveryoneMentions() bool`

HasSuppressEveryoneMentions returns a boolean if a field has been set.

### GetSuppressHereMentions

`func (o *ModelUserSettingsNotifications) GetSuppressHereMentions() bool`

GetSuppressHereMentions returns the SuppressHereMentions field if non-nil, zero value otherwise.

### GetSuppressHereMentionsOk

`func (o *ModelUserSettingsNotifications) GetSuppressHereMentionsOk() (*bool, bool)`

GetSuppressHereMentionsOk returns a tuple with the SuppressHereMentions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSuppressHereMentions

`func (o *ModelUserSettingsNotifications) SetSuppressHereMentions(v bool)`

SetSuppressHereMentions sets SuppressHereMentions field to given value.

### HasSuppressHereMentions

`func (o *ModelUserSettingsNotifications) HasSuppressHereMentions() bool`

HasSuppressHereMentions returns a boolean if a field has been set.

### GetSuppressRoleMentions

`func (o *ModelUserSettingsNotifications) GetSuppressRoleMentions() bool`

GetSuppressRoleMentions returns the SuppressRoleMentions field if non-nil, zero value otherwise.

### GetSuppressRoleMentionsOk

`func (o *ModelUserSettingsNotifications) GetSuppressRoleMentionsOk() (*bool, bool)`

GetSuppressRoleMentionsOk returns a tuple with the SuppressRoleMentions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSuppressRoleMentions

`func (o *ModelUserSettingsNotifications) SetSuppressRoleMentions(v bool)`

SetSuppressRoleMentions sets SuppressRoleMentions field to given value.

### HasSuppressRoleMentions

`func (o *ModelUserSettingsNotifications) HasSuppressRoleMentions() bool`

HasSuppressRoleMentions returns a boolean if a field has been set.

### GetSuppressUserMentions

`func (o *ModelUserSettingsNotifications) GetSuppressUserMentions() bool`

GetSuppressUserMentions returns the SuppressUserMentions field if non-nil, zero value otherwise.

### GetSuppressUserMentionsOk

`func (o *ModelUserSettingsNotifications) GetSuppressUserMentionsOk() (*bool, bool)`

GetSuppressUserMentionsOk returns a tuple with the SuppressUserMentions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSuppressUserMentions

`func (o *ModelUserSettingsNotifications) SetSuppressUserMentions(v bool)`

SetSuppressUserMentions sets SuppressUserMentions field to given value.

### HasSuppressUserMentions

`func (o *ModelUserSettingsNotifications) HasSuppressUserMentions() bool`

HasSuppressUserMentions returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


