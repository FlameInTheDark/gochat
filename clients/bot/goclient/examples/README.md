# GoChat Bot Client Examples

These examples are runnable from the client module:

```powershell
cd clients\bot\goclient
go run .\examples\pingpong -t gcb_your_token
go run .\examples\send_message -t gcb_your_token -channel 123 -content "hello"
go run .\examples\presence -t gcb_your_token -status online -text "building things"
```

The client defaults to `https://gochat.anticode.dev`. Use `-endpoint` in any
example to point it at another GoChat deployment.
