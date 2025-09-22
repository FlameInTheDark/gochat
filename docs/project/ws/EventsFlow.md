[<- Documentation](../README.md) - [WebSocket Events](README.md)

# Events Flow

The main flow of events and authentication is described in the diagram below.

### Message flow after connection to the WebSocket Gateway established
```mermaid
sequenceDiagram
    actor C as Client
    participant S as Server
    Note over C, S: Authentication sequence
    C->>+S: Auth message with token
    alt Authenticated
        S->>-C: Hearhbeat interval
    else Not authenticated
        S-xC: Connection closed
    end

    Note over C, S: Service interactions
    C->>+S: Request to subscribe to a topic (guild list, channel)
    S-->>C: Channel message
    S-->>C: Guild Update event
    S-->>C: Client Update event
    C->>S: Hearthbeat message according to interval
    alt If no hearthbeat message received in interval
        S-xC: Connection closed
    end
    S-->>-C: ...any other events
```
So far it's pretty much simplified, but it's still a work in progress.