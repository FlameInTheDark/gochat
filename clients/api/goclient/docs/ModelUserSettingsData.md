# ModelUserSettingsData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Appearance** | Pointer to [**ModelUserSettingsAppearance**](ModelUserSettingsAppearance.md) |  | [optional] 
**DmChannels** | Pointer to [**[]ModelUserDMChannels**](ModelUserDMChannels.md) |  | [optional] 
**FavoriteGifs** | Pointer to **[]string** |  | [optional] 
**ForcedPresence** | Pointer to **string** |  | [optional] 
**GuildFolders** | Pointer to [**[]ModelUserSettingsGuildFolders**](ModelUserSettingsGuildFolders.md) |  | [optional] 
**Guilds** | Pointer to [**[]ModelUserSettingsGuilds**](ModelUserSettingsGuilds.md) |  | [optional] 
**Language** | Pointer to **string** |  | [optional] 
**Status** | Pointer to [**ModelStatus**](ModelStatus.md) |  | [optional] 

## Methods

### NewModelUserSettingsData

`func NewModelUserSettingsData() *ModelUserSettingsData`

NewModelUserSettingsData instantiates a new ModelUserSettingsData object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewModelUserSettingsDataWithDefaults

`func NewModelUserSettingsDataWithDefaults() *ModelUserSettingsData`

NewModelUserSettingsDataWithDefaults instantiates a new ModelUserSettingsData object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAppearance

`func (o *ModelUserSettingsData) GetAppearance() ModelUserSettingsAppearance`

GetAppearance returns the Appearance field if non-nil, zero value otherwise.

### GetAppearanceOk

`func (o *ModelUserSettingsData) GetAppearanceOk() (*ModelUserSettingsAppearance, bool)`

GetAppearanceOk returns a tuple with the Appearance field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAppearance

`func (o *ModelUserSettingsData) SetAppearance(v ModelUserSettingsAppearance)`

SetAppearance sets Appearance field to given value.

### HasAppearance

`func (o *ModelUserSettingsData) HasAppearance() bool`

HasAppearance returns a boolean if a field has been set.

### GetDmChannels

`func (o *ModelUserSettingsData) GetDmChannels() []ModelUserDMChannels`

GetDmChannels returns the DmChannels field if non-nil, zero value otherwise.

### GetDmChannelsOk

`func (o *ModelUserSettingsData) GetDmChannelsOk() (*[]ModelUserDMChannels, bool)`

GetDmChannelsOk returns a tuple with the DmChannels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDmChannels

`func (o *ModelUserSettingsData) SetDmChannels(v []ModelUserDMChannels)`

SetDmChannels sets DmChannels field to given value.

### HasDmChannels

`func (o *ModelUserSettingsData) HasDmChannels() bool`

HasDmChannels returns a boolean if a field has been set.

### GetFavoriteGifs

`func (o *ModelUserSettingsData) GetFavoriteGifs() []string`

GetFavoriteGifs returns the FavoriteGifs field if non-nil, zero value otherwise.

### GetFavoriteGifsOk

`func (o *ModelUserSettingsData) GetFavoriteGifsOk() (*[]string, bool)`

GetFavoriteGifsOk returns a tuple with the FavoriteGifs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFavoriteGifs

`func (o *ModelUserSettingsData) SetFavoriteGifs(v []string)`

SetFavoriteGifs sets FavoriteGifs field to given value.

### HasFavoriteGifs

`func (o *ModelUserSettingsData) HasFavoriteGifs() bool`

HasFavoriteGifs returns a boolean if a field has been set.

### GetForcedPresence

`func (o *ModelUserSettingsData) GetForcedPresence() string`

GetForcedPresence returns the ForcedPresence field if non-nil, zero value otherwise.

### GetForcedPresenceOk

`func (o *ModelUserSettingsData) GetForcedPresenceOk() (*string, bool)`

GetForcedPresenceOk returns a tuple with the ForcedPresence field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetForcedPresence

`func (o *ModelUserSettingsData) SetForcedPresence(v string)`

SetForcedPresence sets ForcedPresence field to given value.

### HasForcedPresence

`func (o *ModelUserSettingsData) HasForcedPresence() bool`

HasForcedPresence returns a boolean if a field has been set.

### GetGuildFolders

`func (o *ModelUserSettingsData) GetGuildFolders() []ModelUserSettingsGuildFolders`

GetGuildFolders returns the GuildFolders field if non-nil, zero value otherwise.

### GetGuildFoldersOk

`func (o *ModelUserSettingsData) GetGuildFoldersOk() (*[]ModelUserSettingsGuildFolders, bool)`

GetGuildFoldersOk returns a tuple with the GuildFolders field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGuildFolders

`func (o *ModelUserSettingsData) SetGuildFolders(v []ModelUserSettingsGuildFolders)`

SetGuildFolders sets GuildFolders field to given value.

### HasGuildFolders

`func (o *ModelUserSettingsData) HasGuildFolders() bool`

HasGuildFolders returns a boolean if a field has been set.

### GetGuilds

`func (o *ModelUserSettingsData) GetGuilds() []ModelUserSettingsGuilds`

GetGuilds returns the Guilds field if non-nil, zero value otherwise.

### GetGuildsOk

`func (o *ModelUserSettingsData) GetGuildsOk() (*[]ModelUserSettingsGuilds, bool)`

GetGuildsOk returns a tuple with the Guilds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGuilds

`func (o *ModelUserSettingsData) SetGuilds(v []ModelUserSettingsGuilds)`

SetGuilds sets Guilds field to given value.

### HasGuilds

`func (o *ModelUserSettingsData) HasGuilds() bool`

HasGuilds returns a boolean if a field has been set.

### GetLanguage

`func (o *ModelUserSettingsData) GetLanguage() string`

GetLanguage returns the Language field if non-nil, zero value otherwise.

### GetLanguageOk

`func (o *ModelUserSettingsData) GetLanguageOk() (*string, bool)`

GetLanguageOk returns a tuple with the Language field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLanguage

`func (o *ModelUserSettingsData) SetLanguage(v string)`

SetLanguage sets Language field to given value.

### HasLanguage

`func (o *ModelUserSettingsData) HasLanguage() bool`

HasLanguage returns a boolean if a field has been set.

### GetStatus

`func (o *ModelUserSettingsData) GetStatus() ModelStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ModelUserSettingsData) GetStatusOk() (*ModelStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ModelUserSettingsData) SetStatus(v ModelStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *ModelUserSettingsData) HasStatus() bool`

HasStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


