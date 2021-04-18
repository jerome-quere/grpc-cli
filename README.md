# GRPC-CLI

:warning: This project is still under active development

GRPC-CLI aim to be a human friendly cli that can make gRPC calls.
The core concepts are:
 - Takes a FileDescriptorSet to discover service and type dynamically
 - Great autocomplete event in complex message structure

# Quick Start
```
$> cd protobuf-root-folder
$> protoc -I . (find . -name "*.proto") -o ~/proto.descriptor 
$> grpc-cli --descriptor ~/proto.descriptor --target api.test.com:443 rpc package.Service Method param1=value param2=value
```

## Config 

For convenience many flags can be stored in a configuration file.
Config file is located in  "~/.config/grpc-cli/config.yaml" by default

```

target: api.test.com:443
descriptor: ~/proto.descriptor
metadata: 
  x-auth-token: [ "value1", "value2" ]
ca_cert: ~/path-to-ca-cert.pem
cert: ~/path-to-client-cert.pem
key: ~/path-to-client-key.pem
disable_tls: false

# Profiles allow you to easily override some varaibles
profiles:
  prod:
    target: api-prod.test.com:443
  staging:
    target: api-staging.test.com:443  
```


## Roadmap

 - [X] Core initialization
 - [X] Cobra command tree construction
 - [X] Top level flag parsing
 - [X] Basic configuration file support
 - [X] Basic TLS configuration
 - [ ] Basic Argument parsing
 - [ ] Autocompletion
 - [ ] Nested message suport
  - [X] Argument parsing
  - [ ] Documentation generation
  - [ ] Autocompletion support
 - [ ] Request execution
 - [ ] ProtoSet flag can be a remote URL
 - [ ] Streaming support
