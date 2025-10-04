# ModelUserSettingsGuilds

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**GuildId** | Pointer to **int32** |  | [optional] 
**Notifications** | Pointer to [**ModelUserSettingsNotifications**](ModelUserSettingsNotifications.md) |  | [optional] 
**Position** | Pointer to **int32** |  | [optional] 
**ReadStates** | Pointer to [**[]ModelGuildChannelReadState**](ModelGuildChannelReadState.md) |  | [optional] 
**SelectedChannel** | Pointer to **int32** |  | [optional] 

## Methods

### NewModelUserSettingsGuilds

`func NewModelUserSettingsGuilds() *ModelUserSettingsGuilds`

NewModelUserSettingsGuilds instantiates a new ModelUserSettingsGuilds object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewModelUserSettingsGuildsWithDefaults

`func NewModelUserSettingsGuildsWithDefaults() *ModelUserSettingsGuilds`

NewModelUserSettingsGuildsWithDefaults instantiates a new ModelUserSettingsGuilds object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGuildId

`func (o *ModelUserSettingsGuilds) GetGuildId() int32`

GetGuildId returns the GuildId field if non-nil, zero value otherwise.

### GetGuildIdOk

`func (o *ModelUserSettingsGuilds) GetGuildIdOk() (*int32, bool)`

GetGuildIdOk returns a tuple with the GuildId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGuildId

`func (o *ModelUserSettingsGuilds) SetGuildId(v int32)`

SetGuildId sets GuildId field to given value.

### HasGuildId

`func (o *ModelUserSettingsGuilds) HasGuildId() bool`

HasGuildId returns a boolean if a field has been set.

### GetNotifications

`func (o *ModelUserSettingsGuilds) GetNotifications() ModelUserSettingsNotifications`

GetNotifications returns the Notifications field if non-nil, zero value otherwise.

### GetNotificationsOk

`func (o *ModelUserSettingsGuilds) GetNotificationsOk() (*ModelUserSettingsNotifications, bool)`

GetNotificationsOk returns a tuple with the Notifications field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNotifications

`func (o *ModelUserSettingsGuilds) SetNotifications(v ModelUserSettingsNotifications)`

SetNotifications sets Notifications field to given value.

### HasNotifications

`func (o *ModelUserSettingsGuilds) HasNotifications() bool`

HasNotifications returns a boolean if a field has been set.

### GetPosition

`func (o *ModelUserSettingsGuilds) GetPosition() int32`

GetPosition returns the Position field if non-nil, zero value otherwise.

### GetPositionOk

`func (o *ModelUserSettingsGuilds) GetPositionOk() (*int32, bool)`

GetPositionOk returns a tuple with the Position field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPosition

`func (o *ModelUserSettingsGuilds) SetPosition(v int32)`

SetPosition sets Position field to given value.

### HasPosition

`func (o *ModelUserSettingsGuilds) HasPosition() bool`

HasPosition returns a boolean if a field has been set.

### GetReadStates

`func (o *ModelUserSettingsGuilds) GetReadStates() []ModelGuildChannelReadState`

GetReadStates returns the ReadStates field if non-nil, zero value otherwise.

### GetReadStatesOk

`func (o *ModelUserSettingsGuilds) GetReadStatesOk() (*[]ModelGuildChannelReadState, bool)`

GetReadStatesOk returns a tuple with the ReadStates field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadStates

`func (o *ModelUserSettingsGuilds) SetReadStates(v []ModelGuildChannelReadState)`

SetReadStates sets ReadStates field to given value.

### HasReadStates

`func (o *ModelUserSettingsGuilds) HasReadStates() bool`

HasReadStates returns a boolean if a field has been set.

### GetSelectedChannel

`func (o *ModelUserSettingsGuilds) GetSelectedChannel() int32`

GetSelectedChannel returns the SelectedChannel field if non-nil, zero value otherwise.

### GetSelectedChannelOk

`func (o *ModelUserSettingsGuilds) GetSelectedChannelOk() (*int32, bool)`

GetSelectedChannelOk returns a tuple with the SelectedChannel field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSelectedChannel

`func (o *ModelUserSettingsGuilds) SetSelectedChannel(v int32)`

SetSelectedChannel sets SelectedChannel field to given value.

### HasSelectedChannel

`func (o *ModelUserSettingsGuilds) HasSelectedChannel() bool`

HasSelectedChannel returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


