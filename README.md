# Disorder

Disorder is schema free, strong type, streamable bson like binary protocol.
It can used as communication protocol or data storage format.

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

services:
  math_service:
    increase: int -> int
```

* schema and version are fixed fields
* package works as namespace or package in a program language, to prevent name conflict
* option is a string map, used to store extra data for code generation
* import field is external schema files list, now only relative path is supported, later remote (http) schema will be supported
* messages is message map[message name -> message body], nested structures are not allowed. instead we can use complex object type as member type
* message itself is a types map[string -> type]
* enums is a map[enum name -> enum values list], enum values are strings only
* services is service map[name -> service]
* service is rpc methods map[method name : input -> output]
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

services:
  printer:
    print_object: object -> object
    print_time: timestamp -> timestamp
    print_array: array[int] -> array[int]
    print_enum: color -> color
    print_map: map[string] -> map[string]
    print_nested: map[map[array[array[map[color]]]]] -> map[map[array[array[map[color]]]]]
  health_check:
    hello: bool -> bool
```

## TO-DO list
* support remote schema <https/http>
* check if import schema is used
* validate define and rpc separately
* support schema version