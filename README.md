# Disorder

Disorder is a strongly typed data format with two interchangeable encodings.
It only encodes and decodes data, there is no RPC layer.

| encoding | use it for                                                            |
| -------- | --------------------------------------------------------------------- |
| JSON     | development, config, tests: human and AI friendly, easy to edit       |
| binary   | release builds: smaller, faster, fewer allocations, hard to edit      |

Author data as JSON, pack it to binary in the release pipeline. Both encodings
are described by the same yaml schema and decode into the same generated Go types.

## Usage

Generate Go types from a schema:

```
go run github.com/meerkat-io/disorder/internal/cmd/disorder -i schema.yaml -o ./out
```

Generated structs carry both `disorder` and `json` tags, so they work with
`encoding/json` and with this package:

```go
// binary
data, err := disorder.Marshal(&object)
err = disorder.Unmarshal(data, &object)

// json
data, err = json.Marshal(&object)
err = json.Unmarshal(data, &object)
```

Binary decoding also accepts `*interface{}`: you get `bool`, `int32`, `int64`,
`float32`, `float64`, `string`, `[]byte`, `time.Time`, `[]interface{}` and
`map[string]interface{}`. Unknown fields are skipped and reported by
`Decoder.Warnings()`. Map keys are written sorted, so output is deterministic.

JSON and binary are not self-describing in the same way. Encodings are only
guaranteed to match when decoding into a generated type, not into `interface{}`.

### JSON mapping

| type      | JSON                                                    |
| --------- | ------------------------------------------------------- |
| bool      | `true` / `false`                                        |
| int, long | number (long above 2^53 loses precision in JavaScript)  |
| float     | number                                                  |
| string    | string                                                  |
| bytes     | standard Base64 string (`encoding/json` default)        |
| timestamp | RFC 3339 string                                         |
| enum      | string, validated against the schema on encode/decode   |
| array     | array                                                   |
| object    | object                                                  |

Bytes use Base64 because the schema already says the field is `bytes`, so the
decoder knows to decode the string. No marker is needed and it round-trips exactly.



## Data bypes

| type      | tag   | byte | kind      | format                                                  |
| --------- | ----- | ---- | --------- | ------------------------------------------------------- |
| bool      | 1     | 1    | primary   | tag(1) + byte(0/1)                                      |
| int       | 2     | 4    | primary   | tag(2) + 32 bit int                                     |
| long      | 3     | 8    | primary   | tag(3) + 64 bit int                                     |
| float     | 4     | 4    | primary   | tag(4) + 32 bit float                                   |
| double    | 5     | 8    | primary   | tag(5) + 64 bit float                                   |
| bytes     | 6     | var  | primary   | tag(6) + 4 bytes length + [raw bytes]                   |
|           |       |      |           |                                                         |
| string    | 11    | var  | util      | alias "bytes"                                           |
| timestamp | 12    | 8    | util      | alias "long", miliseconds unix time from 1970           |
| enum      | 13    | var  | util      | tag(13) + (short string)*                               |
|           |       |      |           |                                                         |
| array     | 21/22 | var  | container | start(21) + [tag + data] + end(22)                      |
| object    | 23/24 | var  | container | start(23) + [tag + key(short string)* + data] + end(24) |

* Short string: 1 byte length + [raw string], string length < 256
* Different items can belong to the same container, since each item has its own tag

## Schema format

Disorder use yaml as schema file format
Schema file fields:

```
schema: disorder
version: v1
package: test

option:
  go_package_prefix: github.com/meerkat-io/disorder/internal

import: 
  - sub.yaml

messages:
  number:
    value: int

enums:
  color:
    - red
    - green
    - blue
```

* schema and version are fixed fields
* unknown top level keys are rejected, e.g. the removed `services`
* package works as namespace or package in a program language, to prevent name conflict
* option is a string map, used to store extra data for code generation
* import field is external schema files list, now only relative path is supported, later remote (http) schema will be supported
* messages is message map[message name -> message body], nested structures are not allowed. instead we can use complex object type as member type
* message itself is a types map[string -> type]
* enums is a map[enum name -> enum values list], enum values are strings only
* containers type:  array[element_type] and map[element_type]. the key type of map is always string, we don't need to assign type for key again
* nested containers are allowed, eg: map[map[array[array[map[int]]]]]

A more complex example:

```
schema: disorder
version: v1
package: test

import: 
  - sub.yaml

option:
  go_package_prefix: github.com/meerkat-io/disorder/internal

messages:
  object:
    bool_field: bool
    int_field: int
    string_field: string
    bytes_fields: bytes
    enum_field: color
    time_field: timestamp
    obj_field: number
    int_array: array[int]
    int_map: map[int]
    obj_array: array[number]
    obj_map: map[number]
    nested: map[map[array[array[map[color]]]]]

enums:
  color:
    - red
    - green
    - blue
```

## TO-DO list
* support remote schema <https/http>
* check if import schema is used
* validate JSON and binary data against the yaml schema
* support schema version