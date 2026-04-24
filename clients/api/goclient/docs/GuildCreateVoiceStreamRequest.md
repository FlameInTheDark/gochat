# GuildCreateVoiceStreamRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AudioMode** | Pointer to **string** |  | [optional] 
**SourceType** | Pointer to **string** |  | [optional] 

## Methods

### NewGuildCreateVoiceStreamRequest

`func NewGuildCreateVoiceStreamRequest() *GuildCreateVoiceStreamRequest`

NewGuildCreateVoiceStreamRequest instantiates a new GuildCreateVoiceStreamRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGuildCreateVoiceStreamRequestWithDefaults

`func NewGuildCreateVoiceStreamRequestWithDefaults() *GuildCreateVoiceStreamRequest`

NewGuildCreateVoiceStreamRequestWithDefaults instantiates a new GuildCreateVoiceStreamRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAudioMode

`func (o *GuildCreateVoiceStreamRequest) GetAudioMode() string`

GetAudioMode returns the AudioMode field if non-nil, zero value otherwise.

### GetAudioModeOk

`func (o *GuildCreateVoiceStreamRequest) GetAudioModeOk() (*string, bool)`

GetAudioModeOk returns a tuple with the AudioMode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAudioMode

`func (o *GuildCreateVoiceStreamRequest) SetAudioMode(v string)`

SetAudioMode sets AudioMode field to given value.

### HasAudioMode

`func (o *GuildCreateVoiceStreamRequest) HasAudioMode() bool`

HasAudioMode returns a boolean if a field has been set.

### GetSourceType

`func (o *GuildCreateVoiceStreamRequest) GetSourceType() string`

GetSourceType returns the SourceType field if non-nil, zero value otherwise.

### GetSourceTypeOk

`func (o *GuildCreateVoiceStreamRequest) GetSourceTypeOk() (*string, bool)`

GetSourceTypeOk returns a tuple with the SourceType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceType

`func (o *GuildCreateVoiceStreamRequest) SetSourceType(v string)`

SetSourceType sets SourceType field to given value.

### HasSourceType

`func (o *GuildCreateVoiceStreamRequest) HasSourceType() bool`

HasSourceType returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


