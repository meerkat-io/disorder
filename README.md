# Disorder

Disorder is schema free, strong type, streamable bson like binary protocol.
It can used as communication protocol or data storage format.

## Data bype

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

//TO-DO add schema format
//TO-DO support void in rpc