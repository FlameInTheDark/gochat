# UserUserSettingsResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Guilds** | Pointer to [**[]DtoGuild**](DtoGuild.md) |  | [optional] 
**GuildsLastMessages** | Pointer to **map[string]map[string]int32** |  | [optional] 
**ReadStates** | Pointer to **map[string]int32** |  | [optional] 
**Settings** | Pointer to [**ModelUserSettingsData**](ModelUserSettingsData.md) |  | [optional] 
**Version** | Pointer to **int32** |  | [optional] 

## Methods

### NewUserUserSettingsResponse

`func NewUserUserSettingsResponse() *UserUserSettingsResponse`

NewUserUserSettingsResponse instantiates a new UserUserSettingsResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUserUserSettingsResponseWithDefaults

`func NewUserUserSettingsResponseWithDefaults() *UserUserSettingsResponse`

NewUserUserSettingsResponseWithDefaults instantiates a new UserUserSettingsResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGuilds

`func (o *UserUserSettingsResponse) GetGuilds() []DtoGuild`

GetGuilds returns the Guilds field if non-nil, zero value otherwise.

### GetGuildsOk

`func (o *UserUserSettingsResponse) GetGuildsOk() (*[]DtoGuild, bool)`

GetGuildsOk returns a tuple with the Guilds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGuilds

`func (o *UserUserSettingsResponse) SetGuilds(v []DtoGuild)`

SetGuilds sets Guilds field to given value.

### HasGuilds

`func (o *UserUserSettingsResponse) HasGuilds() bool`

HasGuilds returns a boolean if a field has been set.

### GetGuildsLastMessages

`func (o *UserUserSettingsResponse) GetGuildsLastMessages() map[string]map[string]int32`

GetGuildsLastMessages returns the GuildsLastMessages field if non-nil, zero value otherwise.

### GetGuildsLastMessagesOk

`func (o *UserUserSettingsResponse) GetGuildsLastMessagesOk() (*map[string]map[string]int32, bool)`

GetGuildsLastMessagesOk returns a tuple with the GuildsLastMessages field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGuildsLastMessages

`func (o *UserUserSettingsResponse) SetGuildsLastMessages(v map[string]map[string]int32)`

SetGuildsLastMessages sets GuildsLastMessages field to given value.

### HasGuildsLastMessages

`func (o *UserUserSettingsResponse) HasGuildsLastMessages() bool`

HasGuildsLastMessages returns a boolean if a field has been set.

### GetReadStates

`func (o *UserUserSettingsResponse) GetReadStates() map[string]int32`

GetReadStates returns the ReadStates field if non-nil, zero value otherwise.

### GetReadStatesOk

`func (o *UserUserSettingsResponse) GetReadStatesOk() (*map[string]int32, bool)`

GetReadStatesOk returns a tuple with the ReadStates field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadStates

`func (o *UserUserSettingsResponse) SetReadStates(v map[string]int32)`

SetReadStates sets ReadStates field to given value.

### HasReadStates

`func (o *UserUserSettingsResponse) HasReadStates() bool`

HasReadStates returns a boolean if a field has been set.

### GetSettings

`func (o *UserUserSettingsResponse) GetSettings() ModelUserSettingsData`

GetSettings returns the Settings field if non-nil, zero value otherwise.

### GetSettingsOk

`func (o *UserUserSettingsResponse) GetSettingsOk() (*ModelUserSettingsData, bool)`

GetSettingsOk returns a tuple with the Settings field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSettings

`func (o *UserUserSettingsResponse) SetSettings(v ModelUserSettingsData)`

SetSettings sets Settings field to given value.

### HasSettings

`func (o *UserUserSettingsResponse) HasSettings() bool`

HasSettings returns a boolean if a field has been set.

### GetVersion

`func (o *UserUserSettingsResponse) GetVersion() int32`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *UserUserSettingsResponse) GetVersionOk() (*int32, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *UserUserSettingsResponse) SetVersion(v int32)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *UserUserSettingsResponse) HasVersion() bool`

HasVersion returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


