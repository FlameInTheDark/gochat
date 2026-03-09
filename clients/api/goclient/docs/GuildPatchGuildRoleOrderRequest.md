# GuildPatchGuildRoleOrderRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Roles** | Pointer to [**[]GuildRoleOrder**](GuildRoleOrder.md) | List of roles to change order. | [optional] 

## Methods

### NewGuildPatchGuildRoleOrderRequest

`func NewGuildPatchGuildRoleOrderRequest() *GuildPatchGuildRoleOrderRequest`

NewGuildPatchGuildRoleOrderRequest instantiates a new GuildPatchGuildRoleOrderRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGuildPatchGuildRoleOrderRequestWithDefaults

`func NewGuildPatchGuildRoleOrderRequestWithDefaults() *GuildPatchGuildRoleOrderRequest`

NewGuildPatchGuildRoleOrderRequestWithDefaults instantiates a new GuildPatchGuildRoleOrderRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRoles

`func (o *GuildPatchGuildRoleOrderRequest) GetRoles() []GuildRoleOrder`

GetRoles returns the Roles field if non-nil, zero value otherwise.

### GetRolesOk

`func (o *GuildPatchGuildRoleOrderRequest) GetRolesOk() (*[]GuildRoleOrder, bool)`

GetRolesOk returns a tuple with the Roles field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRoles

`func (o *GuildPatchGuildRoleOrderRequest) SetRoles(v []GuildRoleOrder)`

SetRoles sets Roles field to given value.

### HasRoles

`func (o *GuildPatchGuildRoleOrderRequest) HasRoles() bool`

HasRoles returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


