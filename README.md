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

### GDScript (Godot 4)

Generate GDScript classes and use the codec in `lib/gdscript/disorder.gd`
(copy it into your project, it registers `class_name Disorder`):

```
go run github.com/meerkat-io/disorder/internal/cmd/disorder -l gdscript -i schema.yaml -o ./out
```

```gdscript
const Schema = preload("res://out/test/schema.gd")

var object := Schema.Object.new()
object.int_field = 1

var bytes := Disorder.encode(object)                    # binary
var copy = Disorder.decode(bytes, Schema.Object)

var json := Disorder.to_json(object)                    # json
var copy2 = Disorder.from_json(json, Schema.Object)

if copy == null:
    push_error(Disorder.last_error)                     # failures never throw
```

* each schema file becomes one `.gd` file, messages are inner classes, enums are native enums
* `int`/`long` are `int`, `float`/`double` are `float`, `timestamp` is `int` (unix milliseconds),
  `bytes` is `PackedByteArray`, arrays are `Array[T]` (plain `Array` when nested), maps are `Dictionary`
* GDScript cannot tell an unset enum, timestamp, array or map from an empty one, so they are always written;
  an unset message field is `null` and omitted
* generated names must not shadow Godot built-ins (`Object`, `Color`, ...). The generator rejects them,
  add `gdscript_class_prefix: Dis` to the schema `option` to prefix every class and enum
* the Go tests run `lib/gdscript/test/run.gd` against Go-encoded data when Godot 4 is available
  (set `GODOT` or put `godot` on `PATH`)

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

## Binary format

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
| timestamp | 12    | 8    | util      | alias "long", unix time in milliseconds           |
| enum      | 13    | var  | util      | tag(13) + (short string)*                               |
|           |       |      |           |                                                         |
| array     | 21/22 | var  | container | start(21) + [tag + data] + end(22)                      |
| object    | 23/24 | var  | container | start(23) + [tag + key(short string)* + data] + end(24) |

* Short string: 1 byte length + [raw string], string length < 256
* Different items can belong to the same container, since each item has its own tag

## Schema format

Disorder uses yaml as the schema file format:

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
* package works as a namespace (a package in Go) to prevent name conflicts
* option is a string map holding extra data for code generation (`go_package_prefix`, `gdscript_class_prefix`)
* import is a list of external schema files, only relative paths are supported for now
* messages is a map[message name -> message body]. Nested structures are not allowed, use another message as the member type instead
* a message body is a map[field name -> type]
* enums is a map[enum name -> list of values], values are strings
* container types: array[element_type] and map[element_type]. The key type of a map is always string
* nested containers are allowed, e.g. map[map[array[array[map[int]]]]]

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
