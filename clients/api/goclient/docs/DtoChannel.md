# DtoChannel

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CreatedAt** | Pointer to **string** | Timestamp of channel creation | [optional] 
**GuildId** | Pointer to **int32** | Guild ID channel was created in | [optional] 
**Id** | Pointer to **int32** | Channel ID | [optional] 
**Name** | Pointer to **string** | Channel name, without spaces | [optional] 
**ParentId** | Pointer to **int32** | Parent channel id | [optional] 
**Permissions** | Pointer to **int32** | Permissions. Check the permissions documentation for more info. | [optional] 
**Position** | Pointer to **int32** | Channel position | [optional] 
**Private** | Pointer to **bool** | Whether the channel is private. Private channels can only be seen by users with roles assigned to this channel. | [optional] [default to false]
**Roles** | Pointer to **[]int32** | Roles IDs | [optional] 
**Topic** | Pointer to **string** | Channel topic. | [optional] 
**Type** | Pointer to **int32** | Channel type | [optional] 

## Methods

### NewDtoChannel

`func NewDtoChannel() *DtoChannel`

NewDtoChannel instantiates a new DtoChannel object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDtoChannelWithDefaults

`func NewDtoChannelWithDefaults() *DtoChannel`

NewDtoChannelWithDefaults instantiates a new DtoChannel object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCreatedAt

`func (o *DtoChannel) GetCreatedAt() string`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *DtoChannel) GetCreatedAtOk() (*string, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *DtoChannel) SetCreatedAt(v string)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *DtoChannel) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetGuildId

`func (o *DtoChannel) GetGuildId() int32`

GetGuildId returns the GuildId field if non-nil, zero value otherwise.

### GetGuildIdOk

`func (o *DtoChannel) GetGuildIdOk() (*int32, bool)`

GetGuildIdOk returns a tuple with the GuildId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGuildId

`func (o *DtoChannel) SetGuildId(v int32)`

SetGuildId sets GuildId field to given value.

### HasGuildId

`func (o *DtoChannel) HasGuildId() bool`

HasGuildId returns a boolean if a field has been set.

### GetId

`func (o *DtoChannel) GetId() int32`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *DtoChannel) GetIdOk() (*int32, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *DtoChannel) SetId(v int32)`

SetId sets Id field to given value.

### HasId

`func (o *DtoChannel) HasId() bool`

HasId returns a boolean if a field has been set.

### GetName

`func (o *DtoChannel) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DtoChannel) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DtoChannel) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *DtoChannel) HasName() bool`

HasName returns a boolean if a field has been set.

### GetParentId

`func (o *DtoChannel) GetParentId() int32`

GetParentId returns the ParentId field if non-nil, zero value otherwise.

### GetParentIdOk

`func (o *DtoChannel) GetParentIdOk() (*int32, bool)`

GetParentIdOk returns a tuple with the ParentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParentId

`func (o *DtoChannel) SetParentId(v int32)`

SetParentId sets ParentId field to given value.

### HasParentId

`func (o *DtoChannel) HasParentId() bool`

HasParentId returns a boolean if a field has been set.

### GetPermissions

`func (o *DtoChannel) GetPermissions() int32`

GetPermissions returns the Permissions field if non-nil, zero value otherwise.

### GetPermissionsOk

`func (o *DtoChannel) GetPermissionsOk() (*int32, bool)`

GetPermissionsOk returns a tuple with the Permissions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPermissions

`func (o *DtoChannel) SetPermissions(v int32)`

SetPermissions sets Permissions field to given value.

### HasPermissions

`func (o *DtoChannel) HasPermissions() bool`

HasPermissions returns a boolean if a field has been set.

### GetPosition

`func (o *DtoChannel) GetPosition() int32`

GetPosition returns the Position field if non-nil, zero value otherwise.

### GetPositionOk

`func (o *DtoChannel) GetPositionOk() (*int32, bool)`

GetPositionOk returns a tuple with the Position field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPosition

`func (o *DtoChannel) SetPosition(v int32)`

SetPosition sets Position field to given value.

### HasPosition

`func (o *DtoChannel) HasPosition() bool`

HasPosition returns a boolean if a field has been set.

### GetPrivate

`func (o *DtoChannel) GetPrivate() bool`

GetPrivate returns the Private field if non-nil, zero value otherwise.

### GetPrivateOk

`func (o *DtoChannel) GetPrivateOk() (*bool, bool)`

GetPrivateOk returns a tuple with the Private field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrivate

`func (o *DtoChannel) SetPrivate(v bool)`

SetPrivate sets Private field to given value.

### HasPrivate

`func (o *DtoChannel) HasPrivate() bool`

HasPrivate returns a boolean if a field has been set.

### GetRoles

`func (o *DtoChannel) GetRoles() []int32`

GetRoles returns the Roles field if non-nil, zero value otherwise.

### GetRolesOk

`func (o *DtoChannel) GetRolesOk() (*[]int32, bool)`

GetRolesOk returns a tuple with the Roles field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoles

`func (o *DtoChannel) SetRoles(v []int32)`

SetRoles sets Roles field to given value.

### HasRoles

`func (o *DtoChannel) HasRoles() bool`

HasRoles returns a boolean if a field has been set.

### GetTopic

`func (o *DtoChannel) GetTopic() string`

GetTopic returns the Topic field if non-nil, zero value otherwise.

### GetTopicOk

`func (o *DtoChannel) GetTopicOk() (*string, bool)`

GetTopicOk returns a tuple with the Topic field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTopic

`func (o *DtoChannel) SetTopic(v string)`

SetTopic sets Topic field to given value.

### HasTopic

`func (o *DtoChannel) HasTopic() bool`

HasTopic returns a boolean if a field has been set.

### GetType

`func (o *DtoChannel) GetType() int32`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *DtoChannel) GetTypeOk() (*int32, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *DtoChannel) SetType(v int32)`

SetType sets Type field to given value.

### HasType

`func (o *DtoChannel) HasType() bool`

HasType returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


