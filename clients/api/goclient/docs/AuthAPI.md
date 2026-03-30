# \AuthAPI

All URIs are relative to *http://localhost/api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**Auth2faDelete**](AuthAPI.md#Auth2faDelete) | **Delete** /auth/2fa | Disable two-factor auth
[**Auth2faGet**](AuthAPI.md#Auth2faGet) | **Get** /auth/2fa | Get two-factor auth status
[**Auth2faRecoveryCodesRegeneratePost**](AuthAPI.md#Auth2faRecoveryCodesRegeneratePost) | **Post** /auth/2fa/recovery-codes/regenerate | Regenerate recovery codes
[**Auth2faTotpConfirmPost**](AuthAPI.md#Auth2faTotpConfirmPost) | **Post** /auth/2fa/totp/confirm | Confirm TOTP setup
[**Auth2faTotpSetupPost**](AuthAPI.md#Auth2faTotpSetupPost) | **Post** /auth/2fa/totp/setup | Start TOTP setup
[**AuthConfirmationPost**](AuthAPI.md#AuthConfirmationPost) | **Post** /auth/confirmation | Confirmation
[**AuthLogin2faEmailStartPost**](AuthAPI.md#AuthLogin2faEmailStartPost) | **Post** /auth/login/2fa/email/start | Send email recovery code for login
[**AuthLogin2faEmailVerifyPost**](AuthAPI.md#AuthLogin2faEmailVerifyPost) | **Post** /auth/login/2fa/email/verify | Complete login with email recovery code
[**AuthLogin2faRecoveryCodePost**](AuthAPI.md#AuthLogin2faRecoveryCodePost) | **Post** /auth/login/2fa/recovery-code | Complete login with recovery code
[**AuthLogin2faTotpPost**](AuthAPI.md#AuthLogin2faTotpPost) | **Post** /auth/login/2fa/totp | Complete login with TOTP
[**AuthLoginPost**](AuthAPI.md#AuthLoginPost) | **Post** /auth/login | Authentication
[**AuthPasswordChangePost**](AuthAPI.md#AuthPasswordChangePost) | **Post** /auth/password/change | Change password
[**AuthRecoveryPost**](AuthAPI.md#AuthRecoveryPost) | **Post** /auth/recovery | Password Recovery
[**AuthRefreshGet**](AuthAPI.md#AuthRefreshGet) | **Get** /auth/refresh | Refresh authentication token
[**AuthRegistrationPost**](AuthAPI.md#AuthRegistrationPost) | **Post** /auth/registration | Registration
[**AuthResetPost**](AuthAPI.md#AuthResetPost) | **Post** /auth/reset | Password Reset



## Auth2faDelete

> AuthLoginResponse Auth2faDelete(ctx).Request(request).Execute()

Disable two-factor auth



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
	request := *openapiclient.NewAuthDisableTwoFactorRequest() // AuthDisableTwoFactorRequest | Disable two-factor auth request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuthAPI.Auth2faDelete(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthAPI.Auth2faDelete``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `Auth2faDelete`: AuthLoginResponse
	fmt.Fprintf(os.Stdout, "Response from `AuthAPI.Auth2faDelete`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuth2faDeleteRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**AuthDisableTwoFactorRequest**](AuthDisableTwoFactorRequest.md) | Disable two-factor auth request | 

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


## Auth2faGet

> AuthTwoFactorStatusResponse Auth2faGet(ctx).Execute()

Get two-factor auth status



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
	resp, r, err := apiClient.AuthAPI.Auth2faGet(context.Background()).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthAPI.Auth2faGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `Auth2faGet`: AuthTwoFactorStatusResponse
	fmt.Fprintf(os.Stdout, "Response from `AuthAPI.Auth2faGet`: %v\n", resp)
}
```

### Path Parameters

This endpoint does not need any parameter.

### Other Parameters

Other parameters are passed through a pointer to a apiAuth2faGetRequest struct via the builder pattern


### Return type

[**AuthTwoFactorStatusResponse**](AuthTwoFactorStatusResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## Auth2faRecoveryCodesRegeneratePost

> AuthRecoveryCodesResponse Auth2faRecoveryCodesRegeneratePost(ctx).Request(request).Execute()

Regenerate recovery codes



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
	request := *openapiclient.NewAuthRecoveryCodesRegenerateRequest() // AuthRecoveryCodesRegenerateRequest | Recovery code regeneration request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuthAPI.Auth2faRecoveryCodesRegeneratePost(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthAPI.Auth2faRecoveryCodesRegeneratePost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `Auth2faRecoveryCodesRegeneratePost`: AuthRecoveryCodesResponse
	fmt.Fprintf(os.Stdout, "Response from `AuthAPI.Auth2faRecoveryCodesRegeneratePost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuth2faRecoveryCodesRegeneratePostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**AuthRecoveryCodesRegenerateRequest**](AuthRecoveryCodesRegenerateRequest.md) | Recovery code regeneration request | 

### Return type

[**AuthRecoveryCodesResponse**](AuthRecoveryCodesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## Auth2faTotpConfirmPost

> AuthRecoveryCodesResponse Auth2faTotpConfirmPost(ctx).Request(request).Execute()

Confirm TOTP setup



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
	request := *openapiclient.NewAuthTOTPConfirmRequest() // AuthTOTPConfirmRequest | Pending setup ID and authenticator code

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuthAPI.Auth2faTotpConfirmPost(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthAPI.Auth2faTotpConfirmPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `Auth2faTotpConfirmPost`: AuthRecoveryCodesResponse
	fmt.Fprintf(os.Stdout, "Response from `AuthAPI.Auth2faTotpConfirmPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuth2faTotpConfirmPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**AuthTOTPConfirmRequest**](AuthTOTPConfirmRequest.md) | Pending setup ID and authenticator code | 

### Return type

[**AuthRecoveryCodesResponse**](AuthRecoveryCodesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## Auth2faTotpSetupPost

> AuthTOTPSetupResponse Auth2faTotpSetupPost(ctx).Request(request).Execute()

Start TOTP setup



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
	request := *openapiclient.NewAuthTOTPSetupRequest() // AuthTOTPSetupRequest | Current password for TOTP setup

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuthAPI.Auth2faTotpSetupPost(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthAPI.Auth2faTotpSetupPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `Auth2faTotpSetupPost`: AuthTOTPSetupResponse
	fmt.Fprintf(os.Stdout, "Response from `AuthAPI.Auth2faTotpSetupPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuth2faTotpSetupPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**AuthTOTPSetupRequest**](AuthTOTPSetupRequest.md) | Current password for TOTP setup | 

### Return type

[**AuthTOTPSetupResponse**](AuthTOTPSetupResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthConfirmationPost

> string AuthConfirmationPost(ctx).Request(request).Execute()

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
	request := *openapiclient.NewAuthConfirmationRequest() // AuthConfirmationRequest | Login data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuthAPI.AuthConfirmationPost(context.Background()).Request(request).Execute()
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
 **request** | [**AuthConfirmationRequest**](AuthConfirmationRequest.md) | Login data | 

### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthLogin2faEmailStartPost

> string AuthLogin2faEmailStartPost(ctx).Request(request).Execute()

Send email recovery code for login



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
	request := *openapiclient.NewAuthLoginEmailStartRequest() // AuthLoginEmailStartRequest | Login challenge ID

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuthAPI.AuthLogin2faEmailStartPost(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthAPI.AuthLogin2faEmailStartPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthLogin2faEmailStartPost`: string
	fmt.Fprintf(os.Stdout, "Response from `AuthAPI.AuthLogin2faEmailStartPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthLogin2faEmailStartPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**AuthLoginEmailStartRequest**](AuthLoginEmailStartRequest.md) | Login challenge ID | 

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


## AuthLogin2faEmailVerifyPost

> AuthLoginResponse AuthLogin2faEmailVerifyPost(ctx).Request(request).Execute()

Complete login with email recovery code



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
	request := *openapiclient.NewAuthLoginEmailVerifyRequest() // AuthLoginEmailVerifyRequest | Login challenge ID and email recovery code

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuthAPI.AuthLogin2faEmailVerifyPost(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthAPI.AuthLogin2faEmailVerifyPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthLogin2faEmailVerifyPost`: AuthLoginResponse
	fmt.Fprintf(os.Stdout, "Response from `AuthAPI.AuthLogin2faEmailVerifyPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthLogin2faEmailVerifyPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**AuthLoginEmailVerifyRequest**](AuthLoginEmailVerifyRequest.md) | Login challenge ID and email recovery code | 

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


## AuthLogin2faRecoveryCodePost

> AuthLoginResponse AuthLogin2faRecoveryCodePost(ctx).Request(request).Execute()

Complete login with recovery code



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
	request := *openapiclient.NewAuthLoginRecoveryCodeRequest() // AuthLoginRecoveryCodeRequest | Login challenge ID and recovery code

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuthAPI.AuthLogin2faRecoveryCodePost(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthAPI.AuthLogin2faRecoveryCodePost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthLogin2faRecoveryCodePost`: AuthLoginResponse
	fmt.Fprintf(os.Stdout, "Response from `AuthAPI.AuthLogin2faRecoveryCodePost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthLogin2faRecoveryCodePostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**AuthLoginRecoveryCodeRequest**](AuthLoginRecoveryCodeRequest.md) | Login challenge ID and recovery code | 

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


## AuthLogin2faTotpPost

> AuthLoginResponse AuthLogin2faTotpPost(ctx).Request(request).Execute()

Complete login with TOTP



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
	request := *openapiclient.NewAuthLoginTOTPRequest() // AuthLoginTOTPRequest | Login challenge ID and TOTP code

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuthAPI.AuthLogin2faTotpPost(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthAPI.AuthLogin2faTotpPost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthLogin2faTotpPost`: AuthLoginResponse
	fmt.Fprintf(os.Stdout, "Response from `AuthAPI.AuthLogin2faTotpPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthLogin2faTotpPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**AuthLoginTOTPRequest**](AuthLoginTOTPRequest.md) | Login challenge ID and TOTP code | 

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


## AuthLoginPost

> AuthLoginResponse AuthLoginPost(ctx).Request(request).Execute()

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
	request := *openapiclient.NewAuthLoginRequest() // AuthLoginRequest | Login data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuthAPI.AuthLoginPost(context.Background()).Request(request).Execute()
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
 **request** | [**AuthLoginRequest**](AuthLoginRequest.md) | Login data | 

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


## AuthPasswordChangePost

> AuthLoginResponse AuthPasswordChangePost(ctx).Request(request).Execute()

Change password



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
	request := *openapiclient.NewAuthPasswordChangeRequest() // AuthPasswordChangeRequest | Password change request

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuthAPI.AuthPasswordChangePost(context.Background()).Request(request).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthAPI.AuthPasswordChangePost``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthPasswordChangePost`: AuthLoginResponse
	fmt.Fprintf(os.Stdout, "Response from `AuthAPI.AuthPasswordChangePost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthPasswordChangePostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **request** | [**AuthPasswordChangeRequest**](AuthPasswordChangeRequest.md) | Password change request | 

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

> string AuthRecoveryPost(ctx).Request(request).Execute()

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
	request := *openapiclient.NewAuthPasswordRecoveryRequest() // AuthPasswordRecoveryRequest | Email for password recovery

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuthAPI.AuthRecoveryPost(context.Background()).Request(request).Execute()
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
 **request** | [**AuthPasswordRecoveryRequest**](AuthPasswordRecoveryRequest.md) | Email for password recovery | 

### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthRefreshGet

> AuthRefreshTokenResponse AuthRefreshGet(ctx).Authorization(authorization).Execute()

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
	authorization := "authorization_example" // string | Refresh token instead of auth

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuthAPI.AuthRefreshGet(context.Background()).Authorization(authorization).Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `AuthAPI.AuthRefreshGet``: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
	}
	// response from `AuthRefreshGet`: AuthRefreshTokenResponse
	fmt.Fprintf(os.Stdout, "Response from `AuthAPI.AuthRefreshGet`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiAuthRefreshGetRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **authorization** | **string** | Refresh token instead of auth | 

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

> string AuthRegistrationPost(ctx).Request(request).Execute()

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
	request := *openapiclient.NewAuthRegisterRequest() // AuthRegisterRequest | Login data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuthAPI.AuthRegistrationPost(context.Background()).Request(request).Execute()
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
 **request** | [**AuthRegisterRequest**](AuthRegisterRequest.md) | Login data | 

### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## AuthResetPost

> string AuthResetPost(ctx).Request(request).Execute()

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
	request := *openapiclient.NewAuthPasswordResetRequest() // AuthPasswordResetRequest | Password reset data

	configuration := openapiclient.NewConfiguration()
	apiClient := openapiclient.NewAPIClient(configuration)
	resp, r, err := apiClient.AuthAPI.AuthResetPost(context.Background()).Request(request).Execute()
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
 **request** | [**AuthPasswordResetRequest**](AuthPasswordResetRequest.md) | Password reset data | 

### Return type

**string**

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

