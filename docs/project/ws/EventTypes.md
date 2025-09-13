[<- Documentation](../README.md) - [WebSocket Events](README.md)

# Event Types

Here will be the list of all the event types.
The event identifier that will be in the `"t": 100` parameter of the message JSON.

### Message Types
| Type | Description               |
|------|---------------------------|
| 100  | Message Create            |
| 101  | Message Update            |
| 102  | Message Delete            |

### Guild Types
| Type | Description  |
|------|--------------|
| 103  | Guild Create |
| 104  | Guild Update |
| 105  | Guild Delete |

### Channel Types
| Type | Description          |
|------|----------------------|
| 106  | Channel Create       |
| 107  | Channel Update       |
| 108  | Channel Order Update |
| 109  | Channel Delete       |

### Guild Role Types
| Type | Description       |
|------|-------------------|
| 110  | Guild Role Create |
| 111  | Guild Role Update |
| 112  | Guild Role Delete |

### Thread Types
| Type | Description   |
|------|---------------|
| 113  | Thread Create |
| 114  | Thread Update |
| 115  | Thread Delete |

### Guild Member Types
| Type | Description         |
|------|---------------------|
| 200  | Guild Member Added  |
| 201  | Guild Member Update |
| 102  | Guild Member Remove |