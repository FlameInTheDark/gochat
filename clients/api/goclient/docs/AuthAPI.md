# \AuthAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AuthConfirmationPost**](AuthAPI.md#AuthConfirmationPost) | **Post** /auth/confirmation | Confirmation
[**AuthLoginPost**](AuthAPI.md#AuthLoginPost) | **Post** /auth/login | Authentication
[**AuthRecoveryPost**](AuthAPI.md#AuthRecoveryPost) | **Post** /auth/recovery | Password Recovery
[**AuthRefreshGet**](AuthAPI.md#AuthRefreshGet) | **Get** /auth/refresh | Refresh authentication token
[**AuthRegistrationPost**](AuthAPI.md#AuthRegistrationPost) | **Post** /auth/registration | Registration
[**AuthResetPost**](AuthAPI.md#AuthResetPost) | **Post** /auth/reset | Password Reset



## AuthConfirmationPost

> string AuthConfirmationPost(ctx).AuthConfirmationRequest(authConfirmationRequest).Execute()

Confirmation

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/FlameInTheDark/gochat/clients/api/goclient"
)

func main() {
	authConfirmationRequest := *openapiclient.NewAuthConfirmationRequest() // AuthConfirmationRequest | Login data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuthAPI.AuthConfirmationPost(context.Background()).AuthConfirmationRequest(authConfirmationRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthAPI.AuthConfirmationPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthConfirmationPost`: string
	fmt.Fprintf(os.Stdout, "Response from `AuthAPI.AuthConfirmationPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthConfirmationPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **authConfirmationRequest** | [**AuthConfirmationRequest**](AuthConfirmationRequest.md) | Login data | 

### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthLoginPost

> AuthLoginResponse AuthLoginPost(ctx).AuthLoginRequest(authLoginRequest).Execute()

Authentication

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/FlameInTheDark/gochat/clients/api/goclient"
)

func main() {
	authLoginRequest := *openapiclient.NewAuthLoginRequest() // AuthLoginRequest | Login data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuthAPI.AuthLoginPost(context.Background()).AuthLoginRequest(authLoginRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthAPI.AuthLoginPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthLoginPost`: AuthLoginResponse
	fmt.Fprintf(os.Stdout, "Response from `AuthAPI.AuthLoginPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthLoginPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **authLoginRequest** | [**AuthLoginRequest**](AuthLoginRequest.md) | Login data | 

### Return type

[**AuthLoginResponse**](AuthLoginResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthRecoveryPost

> string AuthRecoveryPost(ctx).AuthPasswordRecoveryRequest(authPasswordRecoveryRequest).Execute()

Password Recovery

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/FlameInTheDark/gochat/clients/api/goclient"
)

func main() {
	authPasswordRecoveryRequest := *openapiclient.NewAuthPasswordRecoveryRequest() // AuthPasswordRecoveryRequest | Email for password recovery

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuthAPI.AuthRecoveryPost(context.Background()).AuthPasswordRecoveryRequest(authPasswordRecoveryRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthAPI.AuthRecoveryPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthRecoveryPost`: string
	fmt.Fprintf(os.Stdout, "Response from `AuthAPI.AuthRecoveryPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthRecoveryPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **authPasswordRecoveryRequest** | [**AuthPasswordRecoveryRequest**](AuthPasswordRecoveryRequest.md) | Email for password recovery | 

### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthRefreshGet

> AuthRefreshTokenResponse AuthRefreshGet(ctx).Execute()

Refresh authentication token

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/FlameInTheDark/gochat/clients/api/goclient"
)

func main() {

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuthAPI.AuthRefreshGet(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthAPI.AuthRefreshGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthRefreshGet`: AuthRefreshTokenResponse
	fmt.Fprintf(os.Stdout, "Response from `AuthAPI.AuthRefreshGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiAuthRefreshGetRequest struct via the builder pattern


### Return type

[**AuthRefreshTokenResponse**](AuthRefreshTokenResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthRegistrationPost

> string AuthRegistrationPost(ctx).AuthRegisterRequest(authRegisterRequest).Execute()

Registration

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/FlameInTheDark/gochat/clients/api/goclient"
)

func main() {
	authRegisterRequest := *openapiclient.NewAuthRegisterRequest() // AuthRegisterRequest | Login data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuthAPI.AuthRegistrationPost(context.Background()).AuthRegisterRequest(authRegisterRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthAPI.AuthRegistrationPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthRegistrationPost`: string
	fmt.Fprintf(os.Stdout, "Response from `AuthAPI.AuthRegistrationPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthRegistrationPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **authRegisterRequest** | [**AuthRegisterRequest**](AuthRegisterRequest.md) | Login data | 

### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthResetPost

> string AuthResetPost(ctx).AuthPasswordResetRequest(authPasswordResetRequest).Execute()

Password Reset

### Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	openapiclient "github.com/FlameInTheDark/gochat/clients/api/goclient"
)

func main() {
	authPasswordResetRequest := *openapiclient.NewAuthPasswordResetRequest() // AuthPasswordResetRequest | Password reset data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuthAPI.AuthResetPost(context.Background()).AuthPasswordResetRequest(authPasswordResetRequest).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthAPI.AuthResetPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthResetPost`: string
	fmt.Fprintf(os.Stdout, "Response from `AuthAPI.AuthResetPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthResetPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **authPasswordResetRequest** | [**AuthPasswordResetRequest**](AuthPasswordResetRequest.md) | Password reset data | 

### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

