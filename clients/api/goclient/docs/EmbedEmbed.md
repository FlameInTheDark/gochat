# EmbedEmbed

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Author** | Pointer to [**EmbedEmbedAuthor**](EmbedEmbedAuthor.md) | Embed author metadata. | [optional] 
**Color** | Pointer to **int32** | Decimal RGB color value. | [optional] 
**Description** | Pointer to **string** | Main embed description. | [optional] 
**Fields** | Pointer to [**[]EmbedEmbedField**](EmbedEmbedField.md) | Up to 25 structured fields. | [optional] 
**Footer** | Pointer to [**EmbedEmbedFooter**](EmbedEmbedFooter.md) | Optional footer block. | [optional] 
**Image** | Pointer to [**EmbedEmbedMedia**](EmbedEmbedMedia.md) | Full-size image block. | [optional] 
**Provider** | Pointer to [**EmbedEmbedProvider**](EmbedEmbedProvider.md) | Content provider metadata. | [optional] 
**Thumbnail** | Pointer to [**EmbedEmbedMedia**](EmbedEmbedMedia.md) | Thumbnail image block. | [optional] 
**Timestamp** | Pointer to **time.Time** | Optional ISO timestamp shown by the client. | [optional] 
**Title** | Pointer to **string** | Embed title. | [optional] 
**Type** | Pointer to **string** | Embed type. | [optional] 
**Url** | Pointer to **string** | Canonical URL opened when the embed title is clicked. | [optional] 
**Video** | Pointer to [**EmbedEmbedMedia**](EmbedEmbedMedia.md) | Embedded video metadata. | [optional] 

## Methods

### NewEmbedEmbed

`func NewEmbedEmbed() *EmbedEmbed`

NewEmbedEmbed instantiates a new EmbedEmbed object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEmbedEmbedWithDefaults

`func NewEmbedEmbedWithDefaults() *EmbedEmbed`

NewEmbedEmbedWithDefaults instantiates a new EmbedEmbed object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAuthor

`func (o *EmbedEmbed) GetAuthor() EmbedEmbedAuthor`

GetAuthor returns the Author field if non-nil, zero value otherwise.

### GetAuthorOk

`func (o *EmbedEmbed) GetAuthorOk() (*EmbedEmbedAuthor, bool)`

GetAuthorOk returns a tuple with the Author field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthor

`func (o *EmbedEmbed) SetAuthor(v EmbedEmbedAuthor)`

SetAuthor sets Author field to given value.

### HasAuthor

`func (o *EmbedEmbed) HasAuthor() bool`

HasAuthor returns a boolean if a field has been set.

### GetColor

`func (o *EmbedEmbed) GetColor() int32`

GetColor returns the Color field if non-nil, zero value otherwise.

### GetColorOk

`func (o *EmbedEmbed) GetColorOk() (*int32, bool)`

GetColorOk returns a tuple with the Color field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetColor

`func (o *EmbedEmbed) SetColor(v int32)`

SetColor sets Color field to given value.

### HasColor

`func (o *EmbedEmbed) HasColor() bool`

HasColor returns a boolean if a field has been set.

### GetDescription

`func (o *EmbedEmbed) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *EmbedEmbed) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *EmbedEmbed) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *EmbedEmbed) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetFields

`func (o *EmbedEmbed) GetFields() []EmbedEmbedField`

GetFields returns the Fields field if non-nil, zero value otherwise.

### GetFieldsOk

`func (o *EmbedEmbed) GetFieldsOk() (*[]EmbedEmbedField, bool)`

GetFieldsOk returns a tuple with the Fields field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFields

`func (o *EmbedEmbed) SetFields(v []EmbedEmbedField)`

SetFields sets Fields field to given value.

### HasFields

`func (o *EmbedEmbed) HasFields() bool`

HasFields returns a boolean if a field has been set.

### GetFooter

`func (o *EmbedEmbed) GetFooter() EmbedEmbedFooter`

GetFooter returns the Footer field if non-nil, zero value otherwise.

### GetFooterOk

`func (o *EmbedEmbed) GetFooterOk() (*EmbedEmbedFooter, bool)`

GetFooterOk returns a tuple with the Footer field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFooter

`func (o *EmbedEmbed) SetFooter(v EmbedEmbedFooter)`

SetFooter sets Footer field to given value.

### HasFooter

`func (o *EmbedEmbed) HasFooter() bool`

HasFooter returns a boolean if a field has been set.

### GetImage

`func (o *EmbedEmbed) GetImage() EmbedEmbedMedia`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *EmbedEmbed) GetImageOk() (*EmbedEmbedMedia, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *EmbedEmbed) SetImage(v EmbedEmbedMedia)`

SetImage sets Image field to given value.

### HasImage

`func (o *EmbedEmbed) HasImage() bool`

HasImage returns a boolean if a field has been set.

### GetProvider

`func (o *EmbedEmbed) GetProvider() EmbedEmbedProvider`

GetProvider returns the Provider field if non-nil, zero value otherwise.

### GetProviderOk

`func (o *EmbedEmbed) GetProviderOk() (*EmbedEmbedProvider, bool)`

GetProviderOk returns a tuple with the Provider field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProvider

`func (o *EmbedEmbed) SetProvider(v EmbedEmbedProvider)`

SetProvider sets Provider field to given value.

### HasProvider

`func (o *EmbedEmbed) HasProvider() bool`

HasProvider returns a boolean if a field has been set.

### GetThumbnail

`func (o *EmbedEmbed) GetThumbnail() EmbedEmbedMedia`

GetThumbnail returns the Thumbnail field if non-nil, zero value otherwise.

### GetThumbnailOk

`func (o *EmbedEmbed) GetThumbnailOk() (*EmbedEmbedMedia, bool)`

GetThumbnailOk returns a tuple with the Thumbnail field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThumbnail

`func (o *EmbedEmbed) SetThumbnail(v EmbedEmbedMedia)`

SetThumbnail sets Thumbnail field to given value.

### HasThumbnail

`func (o *EmbedEmbed) HasThumbnail() bool`

HasThumbnail returns a boolean if a field has been set.

### GetTimestamp

`func (o *EmbedEmbed) GetTimestamp() time.Time`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *EmbedEmbed) GetTimestampOk() (*time.Time, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *EmbedEmbed) SetTimestamp(v time.Time)`

SetTimestamp sets Timestamp field to given value.

### HasTimestamp

`func (o *EmbedEmbed) HasTimestamp() bool`

HasTimestamp returns a boolean if a field has been set.

### GetTitle

`func (o *EmbedEmbed) GetTitle() string`

GetTitle returns the Title field if non-nil, zero value otherwise.

### GetTitleOk

`func (o *EmbedEmbed) GetTitleOk() (*string, bool)`

GetTitleOk returns a tuple with the Title field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTitle

`func (o *EmbedEmbed) SetTitle(v string)`

SetTitle sets Title field to given value.

### HasTitle

`func (o *EmbedEmbed) HasTitle() bool`

HasTitle returns a boolean if a field has been set.

### GetType

`func (o *EmbedEmbed) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *EmbedEmbed) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *EmbedEmbed) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *EmbedEmbed) HasType() bool`

HasType returns a boolean if a field has been set.

### GetUrl

`func (o *EmbedEmbed) GetUrl() string`

GetUrl returns the Url field if non-nil, zero value otherwise.

### GetUrlOk

`func (o *EmbedEmbed) GetUrlOk() (*string, bool)`

GetUrlOk returns a tuple with the Url field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUrl

`func (o *EmbedEmbed) SetUrl(v string)`

SetUrl sets Url field to given value.

### HasUrl

`func (o *EmbedEmbed) HasUrl() bool`

HasUrl returns a boolean if a field has been set.

### GetVideo

`func (o *EmbedEmbed) GetVideo() EmbedEmbedMedia`

GetVideo returns the Video field if non-nil, zero value otherwise.

### GetVideoOk

`func (o *EmbedEmbed) GetVideoOk() (*EmbedEmbedMedia, bool)`

GetVideoOk returns a tuple with the Video field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVideo

`func (o *EmbedEmbed) SetVideo(v EmbedEmbedMedia)`

SetVideo sets Video field to given value.

### HasVideo

`func (o *EmbedEmbed) HasVideo() bool`

HasVideo returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


