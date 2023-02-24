[<img src="docs/logo.svg">](https://github.com/checkr/openmock)

<p align="center">
    <a href="https://goreportcard.com/report/github.com/checkr/openmock" target="_blank">
        <img src="https://goreportcard.com/badge/github.com/checkr/openmock">
    </a>
    <a href="https://circleci.com/gh/checkr/openmock" target="_blank">
        <img src="https://circleci.com/gh/checkr/openmock.svg?style=shield">
    </a>
    <a href="https://godoc.org/github.com/checkr/openmock" target="_blank">
        <img src="https://img.shields.io/badge/godoc-reference-green.svg">
    </a>
</p>

# OpenMock

OpenMock is a Go service that can mock services in integration tests, staging environment, or anywhere.
The goal is to simplify the process of writing mocks in various channels.
Currently it supports the following channels:

- HTTP
- gRPC
- Kafka
- AMQP (e.g. RabbitMQ)

# Usage

Use it with docker.

```bash
$ docker run -it -p 9999:9999 -v $(pwd)/demo_templates:/data/templates checkr/openmock 
```

More complete openmock instance (e.g. redis) with docker-compose.

```bash
$ docker-compose up
```

Test it.

```bash
$ curl localhost:9999/ping
```

Dependencies.

- HTTP (native supported, thanks to https://echo.labstack.com/)
  - One can configure HTTP port, set env `OPENMOCK_HTTP_PORT=80`
- GRPC (supported through through HTTP/2 interface)
  - One can configure GRPC port, set env `OPENMOCK_GRPC_PORT=50051`
- Kafka (optional)
  - To enable mocking kafka, set env `OPENMOCK_KAFKA_ENABLED=true`.
  - One can also config the following kafka parameters, optionally with separate config for consumers and producers.  For example `OPENMOCK_KAFKA_SEED_BROKERS`, `OPENMOCK_KAFKA_PRODUCER_SEED_BROKERS`, and `OPENMOCK_KAFKA_CONSUMER_SEED_BROKERS`
    - `OPENMOCK_KAFKA_SEED_BROKERS`
    - `OPENMOCK_KAFKA_SASL_USERNAME`
    - `OPENMOCK_KAFKA_SASL_PASSWORD`
    - `OPENMOCK_KAFKA_TLS_ENABLED`
- AMQP (optional)
  - To enable mocking amqp, set env `OPENMOCK_AMQP_ENABLED=true`
  - One can also config `OPENMOCK_AMQP_URL`.
- NPM (development only)
  - Used in Makefile during swagger admin API server generation

# OpenMock Templates

Templates are YAML files that describe the behavior of OpenMock.

## Templates Directory

You can put any number of `.yaml` or `.yml` files in a directory, and then point
environment variable `OPENMOCK_TEMPLATES_DIR` to it. OpenMock
will recursively (including subdirectories) load all the YAML files. For example:

```sh
# OPENMOCK_TEMPLATES_DIR=./demo_templates

./demo_templates
├── amqp.yaml
├── files
│   └── colors.json
├── http.yaml
├── jsonrpc.yaml
├── kafka.yaml
└── payload_from_file.yaml
```

## Schema

OpenMock is configured a list of behaviors for it to follow. Each behavior is
identified by a key, and a kind:

```yaml
- key: respond-to-resource
  kind: Behavior
```

### Expect

It represents the channel to listen on and condition for the 
actions of the behavior to be performed. Available channels are:

- http
- kafka
- amqp
- grpc

For example, under what condition and from what channel should
we proceed with the actions.

```yaml
- key: no-op
  kind: Behavior
  expect:
    # Condition checks if we need to do the actions or not
    # It only proceeds if it evaluates to "true"
    condition: '{{.HTTPHeader.Get "X-Token" | eq "t1234"}}'
    # Use one (and only one) of the following channels - [http, kafka, amqp]
    http:
      method: GET
      path: /ping
    kafka:
      topic: hello_kafka_in
    amqp:
      exchange: exchange_1
      routing_key: key_in
      queue: key_in
```

### Actions

Actions are a series of functions to run. Availabe actions are:

- publish_amqp
- publish_kafka
- redis
- reply_http
- send_http
- reply_grpc
- sleep

```yaml
- key: every-op
  kind: Behavior
  expect:
    http:
      method: GET
      path: /ping
  actions:
    - publish_kafka:
        topic: hello_kafka_out
        payload: >
          {
            "kafka": "OK",
            "data": {}
          }
    - sleep:
        duration: 1s
    - reply_http:
        status_code: 200
        body: OK
        headers:
          Content-Type: text/html
```

The actions by default run in the order defined in the mock file; you can adjust this by adding an int 'order' value from lowest to highest number. The default value for 'order' is 0.

```yaml
- key: every-op
  kind: Behavior
  expect:
    http:
      method: GET
      path: /ping
  actions:
    - publish_kafka:
        topic: hello_kafka_out
        payload: >
          {
            "kafka": "OK",
            "data": {}
          }
    - sleep:
        duration: 1s
      # sleep first
      order: -1000
```

### Templates

Templates can be useful to assemble your payloads from parts

```yaml
- key: dog
  kind: Template
  template: >
    <animal>dog</animal>

- key: cat
  kind: Template
  template: >
    <animal>cat</animal>

# $ curl 0:9999/fred
# <human>   <name>fred</name>   <pets>     <animal>dog</animal>      <animal>cat</animal>    </pets> </human>
- key: get-freds-pets
  kind: Behavior
  expect:
    http:
      method: GET
      path: /fred
  actions:
    - reply_http:
        status_code: 200
        body: >
          <human>
            <name>fred</name>
            <pets>
              {{template "dog"}}
              {{template "cat"}}
            </pets>
          </human>
```

### Abstract Behaviors

Abstract Behaviors can be used to parameterize some data.

When an abstract behavior and a behavior extending it both have actions defined, all of them are run when the behavior matches.  Actions will run from lowest to highest value of the 'order' field; if this is the same for two actions the action defined earlier in the abstract behavior runs first, followed by actions in the concrete behavior.
Be aware that values with all digits will be interpreted into `int` type (YAML syntax), and it will fail the condition check given that some helper functions are returning `string` types. Pipe to `toString` before the comparison or alternatively put quotes around the values. See example in `abstract_behaviors.yml`.

```yaml
- key: fruit-of-the-day
  kind: AbstractBehavior
  values:
    fruit: potato
  expect:
    condition: '{{.HTTPQueryString | contains .Values.day}}'
    http:
      method: GET
      path: /fruit-of-the-day
  actions:
    - reply_http:
        status_code: 200
        body: '{"fruit": "{{.Values.fruit}}"}'

# $ curl 0:9999/fruit-of-the-day?day=monday
# {"fruit": "apple"}
- key: monday-fruit
  kind: Behavior
  extend: fruit-of-the-day
  values:
    day: monday
    fruit: apple

# $ curl 0:9999/fruit-of-the-day?day=tuesday
# {"fruit": "potato"}
- key: tuesday-fruit
  kind: Behavior
  extend: fruit-of-the-day
  values:
    day: tuesday
  actions: 
    # sleep then reply_http
    - sleep:
         duration: 1s
      order: -1000
```

### Dynamic templating

OpenMock leverages [https://golang.org/pkg/text/template/](https://golang.org/pkg/text/template/) to write dynamic templates. Specifically, it supports a lot of _Context_ and _Helper Functions_.

- Usage of `{{ expr }}`. One can put `{{ expr }}` inside three types of places:
  
  - `expect.condition`
  - `action.http.body`, `action.grpc.payload`, `action.kafka.payload`, `action.amqp.payload`
  - `action.http.body_from_file`, `action.http.body_from_binary_file`, `action.http.binary_file_name` ,`action.grpc.payload_from_file`, `action.kafka.payload_from_file`, `action.amqp.payload_from_file` (`{{ expr }}` will be in the file)

- Use Context inside `{{ expr }}`.
  
  ```bash
  .HTTPHeader      # type: http.Header; example: {{.HTTPHeader.Get "X-Token"}}
  .HTTPBody        # type: string;      example: {{.HTTPBody}}
  .HTTPPath        # type: string;      example: {{.HTTPPath}}
  .HTTPQueryString # type: string;      example: {{.HTTPQueryString}}
  
  .GRPCHeader      # type: string;      example: {{.GRPCHeader}}
  .GRPCPayload     # type: string;      example: {{.GRPCPayload}}
  .GRPCService     # type: string;      example: {{.GRPCService}}
  .GRPCMethod      # type: string;      example: {{.GRPCMethod}}
  
  .KafkaTopic      # type: string;      example: {{.KafkaTopic}}
  .KafkaPayload    # type: string;      example: {{.KafkaPayload}}
  
  .AMQPExchange    # type: string;      example: {{.AMQPExchange}}
  .AMQPRoutingKey  # type: string;      example: {{.AMQPRoutingKey}}
  .AMQPQueue       # type: string;      example: {{.AMQPQueue}}
  .AMQPPayload     # type: string;      example: {{.AMQPPayload}}
  ```

- Use helper functions inside `{{ expr }}`. We recommend pipeline format (`|`) of the functions.
  
  ```bash
  # Supported functions defined in ./template_helper.go
  
    - 
    - jsonPath    # doc: https://github.com/antchfx/xpath
    - gJsonPath   # doc: https://github.com/tidwall/gjson
    - xmlPath     # doc: https://github.com/antchfx/xpath
    - uuidv5      # uuid v5 sha1 hash
    - redisDo     # run redis commands. For example {{redisDo "RPUSH" "arr" "hi"}}
    - ...
  
  # Supported functions inherited from
  # https://github.com/Masterminds/sprig/blob/master/functions.go
  
    - replace
    - uuidv4
    - regexMatch
    - ...
  
  # Examples
  {{.HTTPHeader.Get "X-Token" | eq "t1234"}}
  {{.HTTPBody | jsonPath "user/first_name" | replace "A" "a" | uuidv5 }}
  {{.HTTPBody | gJsonPath "users.0.first_name" }}
  {{.HTTPBody | xmlPath "node1/node2/node3"}}
  ```

## Admin Interface

Openmock also by default provides an API on port 9998 to control the running instance.  See [api documentation](docs/api_docs/bundle.yaml).  You can serve the api documentation by getting [go-swagger](https://github.com/go-swagger/go-swagger) and running:

```
./swagger serve --host 0.0.0.0 --port 9997 docs/api_docs/bundle.yaml"
```

## Command Line Interface

Openmock has a command-line interface to help with certain tasks interacting with openmock instances. This is 
invoked with the `omctl` command.  This uses the [cobra](https://github.com/spf13/cobra) library to provide a discoverable CLI; run `omctl` for a list of commands / flags. 

### CLI: Directory

#### Push

Pushes a local openmock model from the file system to a remote instance.

```
# Adds templates from the ./demo_templates directory to the instance running on localhost.
omctl push --directory ./demo_templates --url http://localhost:9998
```

## Examples

### Example: Mock HTTP

```yaml
# demo_templates/http.yaml

# $ curl 0:9999/ping
# OK
- key: ping
  kind: Behavior
  expect:
    http:
      method: GET
      path: /ping
  actions:
    - reply_http:
        status_code: 200
        body: OK
        headers:
          Content-Type: text/html

# $ curl 0:9999/token -H X-Token:t1234 -H Y-Token:t1234
# OK
- key: header-token-200
  kind: Behavior
  expect:
    condition: '{{.HTTPHeader.Get "X-Token" | eq "t1234" | and (.HTTPHeader.Get "Y-Token" | eq "t1234")}}'
    http:
      method: GET
      path: /token
  actions:
    - reply_http:
        status_code: 200
        body: OK

# $ curl 0:9999/token
# Invalid X-Token
- key: header-token-401
  kind: Behavior
  expect:
    condition: '{{.HTTPHeader.Get "X-Token" | ne "t1234"}}'
    http:
      method: GET
      path: /token
  actions:
    - reply_http:
        status_code: 401
        body: Invalid X-Token
```

### Example: Mock HTTP and reply with binary file

```yaml
- key: get-pdf
  expect:
    http:
      method: GET
      path: /api/v1/:ClientID/pdf
    condition: '{{
      (.HTTPHeader.Get "Authorization" | contains "exp") | and
      (.HTTPHeader.Get "x-timestamp" | eq "" | not) 
    }}'
  actions:
    - reply_http:
        status_code: 200
        headers:
          Content-Type: application/pdf
        body_from_binary_file: ./data/example.pdf
        binary_file_name: example_pdf.pdf # optional file name
```

### 

### Example: Mock HTTP and reply with body from file

```yaml
- key: get-json
  expect:
    http:
      method: GET
      path: /api/v1/:ClientID/json
    condition: '{{
      (.HTTPHeader.Get "Authorization" | contains "exp") | and
      (.HTTPHeader.Get "x-timestamp" | eq "" | not) 
    }}'
  actions:
    - reply_http:
        status_code: 200
        headers:
          Content-Type: application/json
        body_from_file: ./data/example.json # only text files supported
```

### Example: Mock GRPC

```yaml
# demo_templates/grpc.yaml

- key: example_grpc
  expect:
    grpc:
      service: demo_protobuf.ExampleService
      method: ExampleMethod
  actions:
    - reply_grpc:
        payload_from_file: './files/example_grpc_response.json'
```

### Example: Mock Kafka

```yaml
# demo_templates/kafka.yaml

- key: test_kafka_1
  kind: Behavior
  expect:
    kafka:
      topic: hello_kafka_in
  actions:
    - publish_kafka:
        topic: hello_kafka_out
        payload: >
          {
            "kafka": "OK",
            "data": {}
          }

- key: test_kafka_2
  kind: Behavior
  expect:
    kafka:
      topic: hello_kafka_in_2
  actions:
    - publish_kafka:
        topic: hello_kafka_out
        payload_from_file: './files/colors.json' # the path is relative to OPENMOCK_TEMPLATES_DIR
```

If you started the example from docker-compose, you can test the above kafka mocks by using a kt docker container.

```bash
# Exec into the container
docker-compose exec kt bash

# Run some kt commands inside the container
# Notice that the container is within the docker-compose network, and it connects to "kafka:9092"

$ kt topic
$ echo '{"123":"hi"}' | kt produce -topic hello_kafka_in -literal
$ kt consume -topic hello_kafka_out -offsets all=newest:newest
```

### Example: Mock AMQP (e.g. RabbitMQ)

```yaml
# demo_templates/amqp.yaml

- key: test_amqp_1
  kind: Behavior
  expect:
    amqp:
      exchange: exchange_1
      routing_key: key_in
      queue: key_in
  actions:
    - publish_amqp:
        exchange: exchange_1
        routing_key: key_out
        payload: >
          {
            "amqp": "OK",
            "data": {}
          }

- key: test_amqp_2
  kind: Behavior
  expect:
    amqp:
      exchange: exchange_1
      routing_key: key_in
      queue: key_in
  actions:
    - publish_amqp:
        exchange: exchange_1
        routing_key: key_out
        payload_from_file: './files/colors.json'
```

### Example: Use Redis for stateful things (by default, OpenMock uses an in-memory miniredis)

```yaml
# demo_templates/redis.yaml

- key: hello_redis
  kind: Behavior
  expect:
    http:
      method: GET
      path: /test_redis
  actions:
    - redis:
      - '{{.HTTPHeader.Get "X-TOKEN" | redisDo "SET" "k1"}}'
      - '{{redisDo "RPUSH" "random" uuidv4}}'
      - '{{redisDo "RPUSH" "random" uuidv4}}'
      - '{{redisDo "RPUSH" "random" uuidv4}}'
    - reply_http:
        status_code: 200
        body: >
          {
            "k1": "{{redisDo "GET" "k1"}}",
            "randomStr": "{{redisDo "LRANGE" "random" 0 -1}}",
            "random": [
              {{ $arr := redisDo "LRANGE" "random" 0 -1 | splitList ";;" }}
              {{ range $i, $v := $arr }}
                {{if isLastIndex $i $arr}}
                  "{{$v}}"
                {{else}}
                  "{{$v}}",
                {{end}}
              {{end}}
            ]
          }

# To test
# curl localhost:9999/test_redis -H "X-TOKEN:t123"  | jq .
```

### Example: Send Webhooks

```yaml
# demo_templates/webhook.yaml

- key: webhooks
  kind: Behavior
  expect:
    http:
      method: GET
      path: /send_webhook_to_httpbin
  actions:
    - send_http:
        url: "https://httpbin.org/post"
        method: POST
        body: '{"hello": "world"}'
        headers:
          X-Token: t123
    - reply_http:
        status_code: 200
        body: 'webhooks sent'

# To test
# curl localhost:9999/send_webhook_to_httpbin
```

### Example: Use data in templates

```yaml
# demo_templates/http.yaml

- key: http-request-template
  kind: Template
  template: >
    { "http_path": "{{.HTTPPath}}", "http_headers": "{{.HTTPHeader}}" }

- key: color-template
  kind: Template
  template: >
    { "color": "{{.color}}" }

- key: teapot
  kind: AbstractBehavior
  expect:
    http:
      method: GET
      path: /teapot
  actions:
    - reply_http:
        status_code: 418
        body: >
          {
            "request-info": {{ template "http-request-template" . }},
            "teapot-info": {{ template "color-template" .Values }}
          }

# $ curl 0:9999/teapot
# {   "request-info": { "http_path": "/teapot", "http_headers": "map[Accept:[*/*] User-Agent:[curl/7.54.0]]" } ,   "teapot-info": { "color": "purple" }  }
- key: purple-teapot
  kind: Behavior
  extend: teapot
  values:
    color: purple
```

# Advanced pipeline functions

To enable advanced mocks, for example, your own encoding/decoding of the kafka messages,
one can develop by directly importing the `github.com/checkr/openmock` package, making a copy of the swagger-generated server main, and passing in a custom OpenMock.

For example:
(see [example](https://github.com/sesquipedalian-dev/openmock-custom-example/blob/master/main.go))

```go
package main

import (
  "github.com/checkr/openmock"
  "github.com/checkr/openmock/swagger_gen/restapi"
  "github.com/checkr/openmock/swagger_gen/restapi/operations"
  /// etc
)

func consumePipelineFunc(c openmock.Context, in []byte) (out []byte, error) {
  return decode(in), nil
}

func main() {
  // server set up copy & paste...

  // add our custom openmock functionality
  om := &openmock.OpenMock{}
  om.ParseEnv()
  om.KafkaConsumePipelineFunc = consumePipelineFunc

  server.ConfigureAPI(om)

  // rest of server set up copy & paste...
}
```

## GRPC Configuration Notes

OpenMock uses the APIv2 protobuf module (google.golang.org/protobuf). If your project uses the APIv1 protobuf module,
you can use https://github.com/golang/protobuf/releases/tag/v1.4.0 and convert your messages to be APIv2 compatible
with the `proto.MessageV2` method.

Please note that OpenMock expects the `payload` or `payload_from_file` for a reply_grpc action to be in the json
form of your `Response` protobuf message.  The request should be in the `Request` protobuf message format
as it is parsed into json to support `jsonPath` and `gJsonPath` operations.

Example configuration by directly importing the `github.com/checkr/openmock` package into a wrapper project.

```
func main() {
  // server set up copy & paste...

  // add our custom openmock functionality
  om := &openmock.OpenMock{}
  om.GRPCServiceMap = map[string]openmock.GRPCService{
      "demo_protobuf.ExampleService": {
          "ExampleMethod": openmock.RequestResponsePair{
              Request:  proto.MessageV2(&demo_protobuf.ExampleRequest{}),
              Response: proto.MessageV2(&demo_protobuf.ExampleResponse{}),
          },
      },
  }
  om.ParseEnv()
  server.ConfigureAPI(om)

  // rest of server set up copy & paste...
```

## Swagger

### Swagger files / directories:

```
Makefile                    # contains build process for swagger generation
swagger/                    # directory containing swagger definition, split 
                            # up into a few files
   index.yaml               # all the model definitions are in here
   health.yaml              # method definitions relating to e.g. /health
swagger_gen/                # directory where generated swagger files go
  restapi/
    configure_open_mock.go  # this file contains code further customized from the 
                            # generated code to hook an implementation into the API
                            # the makefiles makes sure it is preserved when 
                            # generating the other files
docs/
  api_docs/
    bundle.yaml             # combined swagger spec file, generated by Makefile
pkg/
  admin/                    # code implementing the handlers for the swagger API
```

### Install Swagger

```
brew tap go-swagger/go-swagger
brew install go-swagger
```

### Generate

* `make gen` - bundles the separate swagger files and generates swagger_gen
* `make build` - builds the executables `om` and `omctl`

### Run

`OPENMOCK_REDIS_TYPE=redis OPENMOCK_REDIS_URL=<redis Url, e.g. redis://localhost:6379> OPENMOCK_TEMPLATES_DIR=./demo_templates ./om --port 9998`
