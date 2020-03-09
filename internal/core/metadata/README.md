# EdgeX Foundry Core Metadata Service
[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

Metadata retains and provides access to the knowledge about the devices and sensors connected to EdgeX and how to communicate with them. More specifically,it manages information about the devices and sensors connected to, and operated by, EdgeX Foundry, knows the type, and organization of data reported by the devices and sensors, and it knows how to command the devices and sensors.  This service may also hold and manage other configuration metadata used by other services on the gateway – such as clean up schedules, hardware configuration (Wi-Fi connection info, MQTT queues, etc.). Non-device metadata may need to be held in a different database and/or managed by another service – depending on implementation.

# Install and Deploy Native #

### Prerequisites ###
Serveral EdgeX Foundry services depend on ZeroMQ for communications by default.  The easiest way to get and install ZeroMQ is to use or follow the following setup script:  https://gist.github.com/katopz/8b766a5cb0ca96c816658e9407e83d00.

**Note**: Setup of the ZeroMQ library is not supported on Windows plaforms.

### Installation and Execution ###
To fetch the code and build the microservice execute the following:

```
cd $GOPATH/src
go get github.com/edgexfoundry/edgex-go
cd $GOPATH/src/github.com/edgexfoundry/edgex-go
# pull the 3rd party / vendor packages
make prepare
# build the microservice
make cmd/core-metadata/core-metadata
# get to the core metadata microservice executable
cd cmd/core-metadata
# run the microservice (may require other dependent services to run correctly)
./core-metadata
```


## Community
- Chat: [https://edgexfoundry.slack.com](https://join.slack.com/t/edgexfoundry/shared_invite/enQtNDgyODM5ODUyODY0LWVhY2VmOTcyOWY2NjZhOWJjOGI1YzQ2NzYzZmIxYzAzN2IzYzY0NTVmMWZhZjNkMjVmODNiZGZmYTkzZDE3MTA)
- Mainling lists: https://lists.edgexfoundry.org/mailman/listinfo

## License
[Apache-2.0](LICENSE)

