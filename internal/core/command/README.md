# EdgeX Foundry Core Command Service
[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE)

Command Service is the conduit for other services to trigger action on devices/sensors via there managing device services. The service provides an API to get the list of commands that can be issued for all devices or a single device. Commands are divided into to groups for each device: Gets and Puts. Get commands are issued to a device/sensor get a current value for a particular attribute on the device (like the current temperature offered by a thermostat sensor, or like the on/off status of a light). Put commands are issued to a device/sensor to change the current state or status of a device or one of its attributes (like setting the speed in RPMs of a motor or setting the brightness of a dimmer light).

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
make cmd/core-command/core-command
# get to the command microservice executable
cd cmd/core-command
# run the microservice (may require other dependent services to run correctly)
./core-command
```


## Community
- Chat: [https://edgexfoundry.slack.com](https://join.slack.com/t/edgexfoundry/shared_invite/enQtNDgyODM5ODUyODY0LWVhY2VmOTcyOWY2NjZhOWJjOGI1YzQ2NzYzZmIxYzAzN2IzYzY0NTVmMWZhZjNkMjVmODNiZGZmYTkzZDE3MTA)
- Mainling lists: https://lists.edgexfoundry.org/mailman/listinfo

## License
[Apache-2.0](LICENSE)

