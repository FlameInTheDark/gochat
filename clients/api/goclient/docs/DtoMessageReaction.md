# DtoMessageReaction

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Count** | Pointer to **int32** |  | [optional] 
**Emoji** | Pointer to [**DtoMessageReactionEmoji**](DtoMessageReactionEmoji.md) |  | [optional] 
**Me** | Pointer to **bool** |  | [optional] 

## Methods

### NewDtoMessageReaction

`func NewDtoMessageReaction() *DtoMessageReaction`

NewDtoMessageReaction instantiates a new DtoMessageReaction object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDtoMessageReactionWithDefaults

`func NewDtoMessageReactionWithDefaults() *DtoMessageReaction`

NewDtoMessageReactionWithDefaults instantiates a new DtoMessageReaction object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCount

`func (o *DtoMessageReaction) GetCount() int32`

GetCount returns the Count field if non-nil, zero value otherwise.

### GetCountOk

`func (o *DtoMessageReaction) GetCountOk() (*int32, bool)`

GetCountOk returns a tuple with the Count field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCount

`func (o *DtoMessageReaction) SetCount(v int32)`

SetCount sets Count field to given value.

### HasCount

`func (o *DtoMessageReaction) HasCount() bool`

HasCount returns a boolean if a field has been set.

### GetEmoji

`func (o *DtoMessageReaction) GetEmoji() DtoMessageReactionEmoji`

GetEmoji returns the Emoji field if non-nil, zero value otherwise.

### GetEmojiOk

`func (o *DtoMessageReaction) GetEmojiOk() (*DtoMessageReactionEmoji, bool)`

GetEmojiOk returns a tuple with the Emoji field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmoji

`func (o *DtoMessageReaction) SetEmoji(v DtoMessageReactionEmoji)`

SetEmoji sets Emoji field to given value.

### HasEmoji

`func (o *DtoMessageReaction) HasEmoji() bool`

HasEmoji returns a boolean if a field has been set.

### GetMe

`func (o *DtoMessageReaction) GetMe() bool`

GetMe returns the Me field if non-nil, zero value otherwise.

### GetMeOk

`func (o *DtoMessageReaction) GetMeOk() (*bool, bool)`

GetMeOk returns a tuple with the Me field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMe

`func (o *DtoMessageReaction) SetMe(v bool)`

SetMe sets Me field to given value.

### HasMe

`func (o *DtoMessageReaction) HasMe() bool`

HasMe returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


