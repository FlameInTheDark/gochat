# DtoAttachmentUpload

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ChannelId** | Pointer to **int32** | Channel ID the attachment was sent to | [optional] 
**FileName** | Pointer to **string** | File name | [optional] 
**Id** | Pointer to **int32** | Attachment ID | [optional] 
**UploadUrl** | Pointer to **string** | Upload URL. S3 presigned URL | [optional] 

## Methods

### NewDtoAttachmentUpload

`func NewDtoAttachmentUpload() *DtoAttachmentUpload`

NewDtoAttachmentUpload instantiates a new DtoAttachmentUpload object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDtoAttachmentUploadWithDefaults

`func NewDtoAttachmentUploadWithDefaults() *DtoAttachmentUpload`

NewDtoAttachmentUploadWithDefaults instantiates a new DtoAttachmentUpload object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetChannelId

`func (o *DtoAttachmentUpload) GetChannelId() int32`

GetChannelId returns the ChannelId field if non-nil, zero value otherwise.

### GetChannelIdOk

`func (o *DtoAttachmentUpload) GetChannelIdOk() (*int32, bool)`

GetChannelIdOk returns a tuple with the ChannelId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetChannelId

`func (o *DtoAttachmentUpload) SetChannelId(v int32)`

SetChannelId sets ChannelId field to given value.

### HasChannelId

`func (o *DtoAttachmentUpload) HasChannelId() bool`

HasChannelId returns a boolean if a field has been set.

### GetFileName

`func (o *DtoAttachmentUpload) GetFileName() string`

GetFileName returns the FileName field if non-nil, zero value otherwise.

### GetFileNameOk

`func (o *DtoAttachmentUpload) GetFileNameOk() (*string, bool)`

GetFileNameOk returns a tuple with the FileName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFileName

`func (o *DtoAttachmentUpload) SetFileName(v string)`

SetFileName sets FileName field to given value.

### HasFileName

`func (o *DtoAttachmentUpload) HasFileName() bool`

HasFileName returns a boolean if a field has been set.

### GetId

`func (o *DtoAttachmentUpload) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DtoAttachmentUpload) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DtoAttachmentUpload) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *DtoAttachmentUpload) HasId() bool`

HasId returns a boolean if a field has been set.

### GetUploadUrl

`func (o *DtoAttachmentUpload) GetUploadUrl() string`

GetUploadUrl returns the UploadUrl field if non-nil, zero value otherwise.

### GetUploadUrlOk

`func (o *DtoAttachmentUpload) GetUploadUrlOk() (*string, bool)`

GetUploadUrlOk returns a tuple with the UploadUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUploadUrl

`func (o *DtoAttachmentUpload) SetUploadUrl(v string)`

SetUploadUrl sets UploadUrl field to given value.

### HasUploadUrl

`func (o *DtoAttachmentUpload) HasUploadUrl() bool`

HasUploadUrl returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


