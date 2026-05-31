# GuildInstallBotRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BotUserId** | Pointer to **int32** |  | [optional] 
**GrantToken** | Pointer to **string** |  | [optional] 
**GrantedPermissions** | Pointer to **int32** |  | [optional] 

## Methods

### NewGuildInstallBotRequest

`func NewGuildInstallBotRequest() *GuildInstallBotRequest`

NewGuildInstallBotRequest instantiates a new GuildInstallBotRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGuildInstallBotRequestWithDefaults

`func NewGuildInstallBotRequestWithDefaults() *GuildInstallBotRequest`

NewGuildInstallBotRequestWithDefaults instantiates a new GuildInstallBotRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBotUserId

`func (o *GuildInstallBotRequest) GetBotUserId() int32`

GetBotUserId returns the BotUserId field if non-nil, zero value otherwise.

### GetBotUserIdOk

`func (o *GuildInstallBotRequest) GetBotUserIdOk() (*int32, bool)`

GetBotUserIdOk returns a tuple with the BotUserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBotUserId

`func (o *GuildInstallBotRequest) SetBotUserId(v int32)`

SetBotUserId sets BotUserId field to given value.

### HasBotUserId

`func (o *GuildInstallBotRequest) HasBotUserId() bool`

HasBotUserId returns a boolean if a field has been set.

### GetGrantToken

`func (o *GuildInstallBotRequest) GetGrantToken() string`

GetGrantToken returns the GrantToken field if non-nil, zero value otherwise.

### GetGrantTokenOk

`func (o *GuildInstallBotRequest) GetGrantTokenOk() (*string, bool)`

GetGrantTokenOk returns a tuple with the GrantToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGrantToken

`func (o *GuildInstallBotRequest) SetGrantToken(v string)`

SetGrantToken sets GrantToken field to given value.

### HasGrantToken

`func (o *GuildInstallBotRequest) HasGrantToken() bool`

HasGrantToken returns a boolean if a field has been set.

### GetGrantedPermissions

`func (o *GuildInstallBotRequest) GetGrantedPermissions() int32`

GetGrantedPermissions returns the GrantedPermissions field if non-nil, zero value otherwise.

### GetGrantedPermissionsOk

`func (o *GuildInstallBotRequest) GetGrantedPermissionsOk() (*int32, bool)`

GetGrantedPermissionsOk returns a tuple with the GrantedPermissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGrantedPermissions

`func (o *GuildInstallBotRequest) SetGrantedPermissions(v int32)`

SetGrantedPermissions sets GrantedPermissions field to given value.

### HasGrantedPermissions

`func (o *GuildInstallBotRequest) HasGrantedPermissions() bool`

HasGrantedPermissions returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


