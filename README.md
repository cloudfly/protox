# protox

## Usage

### 1. Copy `protox.proto` files.
Add the `protox.proto` into your proto folder. The folder includes by folder configured in inputs section in `buf.yaml`

### 2. Import `protox.proto` in your on `.proto` file.

```
syntax = "proto3";

package test;

import "protox.proto";
```

### Use field and message descripters

```protobuf
message Test {
  // ignore the field in json marshal and unmarshal
  int64 id = 1 [(protox.gojson) = '-'];
  // json name is 'NAME', and only used to marshal into json, but can not unmarshal from json
  optional string name   = 2 [(protox.gojson) = "NAME,readonly"];
  // json name is '_public', and only used to unmarshal from json, but can not marshal into json
  optional string Public = 3 [(protox.gojson) = "_public,writeonly"];

  // command fields
  string publicId = 5;
  string type = 8; 
  string description = 9;
  int64 projectId = 11;
  // protox.Timestamp will be marshaled into a timestamp integer in seconds
  protox.Timestamp createTime = 12;

  // this strutcure will be serialized into json while being inserted into database
  // the json serializer will ignore the (protox.gojson) option defined above.
  option (protox.gosql) = {
    serializer: "json"
  };
  // custom a golang method for structure
  option (protox.gomethod) = {
    name  : "Table",
    return: "Test"
  };
}

```