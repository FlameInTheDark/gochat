# SendPulse REST client library (unofficial)
An unofficial SendPulse SDK client library for Go (Golang). This library allows to use 
the SendPulse API via Go.

SendPulse API official documentation [https://sendpulse.com/api](https://sendpulse.com/api)

[![Build Status](https://travis-ci.com/dimuska139/sendpulse-sdk-go.svg?branch=master)](https://travis-ci.org/dimuska139/sendpulse-sdk-go)
[![codecov](https://codecov.io/gh/dimuska139/sendpulse-sdk-go/branch/master/graph/badge.svg)](https://codecov.io/gh/dimuska139/sendpulse-sdk-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/dimuska139/sendpulse-sdk-go)](https://goreportcard.com/report/github.com/dimuska139/sendpulse-sdk-go)
[![License](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/dimuska139/sendpulse-sdk-go/blob/master/LICENSE)

### Installation

```shell
go get -u github.com/dimuska139/sendpulse-sdk-go/v7
```

### Usage
```go
package main

import (
    "context"
    "fmt"
    sendpulse "github.com/dimuska139/sendpulse-sdk-go/v7"
    "net/http"
)

func main() {
    config := &sendpulse.Config{
        UserID: "",
        Secret: "",
    }
    client := sendpulse.NewClient(http.DefaultClient, config)

    emails := []*sendpulse.EmailToAdd {
        &sendpulse.EmailToAdd{
            Email:     "test@test.com",
            Variables: map[string]interface{}{"age": 21, "weight": 99},
        },
    }
    
    ctx := context.Background()
    if err := client.Emails.MailingLists.SingleOptIn(ctx, 1266208, emails); err != nil {
        fmt.Println(err)
    }
    fmt.Println(*emails[0])
}
```

The tests should be considered a part of the documentation.

### License
[The MIT License (MIT)](LICENSE)
