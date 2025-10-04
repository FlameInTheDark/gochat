# ModelUserSettingsData

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Appearance** | Pointer to [**ModelUserSettingsAppearance**](ModelUserSettingsAppearance.md) |  | [optional] 
**FavoriteGifs** | Pointer to **[]string** |  | [optional] 
**GuildFolders** | Pointer to [**[]ModelUserSettingsGuildFolders**](ModelUserSettingsGuildFolders.md) |  | [optional] 
**Guilds** | Pointer to [**[]ModelUserSettingsGuilds**](ModelUserSettingsGuilds.md) |  | [optional] 
**Language** | Pointer to **string** |  | [optional] 
**SelectedGuild** | Pointer to **int32** |  | [optional] 

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

### GetSelectedGuild

`func (o *ModelUserSettingsData) GetSelectedGuild() int32`

GetSelectedGuild returns the SelectedGuild field if non-nil, zero value otherwise.

### GetSelectedGuildOk

`func (o *ModelUserSettingsData) GetSelectedGuildOk() (*int32, bool)`

GetSelectedGuildOk returns a tuple with the SelectedGuild field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSelectedGuild

`func (o *ModelUserSettingsData) SetSelectedGuild(v int32)`

SetSelectedGuild sets SelectedGuild field to given value.

### HasSelectedGuild

`func (o *ModelUserSettingsData) HasSelectedGuild() bool`

HasSelectedGuild returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


