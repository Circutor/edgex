# EdgeX Foundry Support Scheduler Service

Go implementation of EdgeX Support Scheduler.

### Installation and Execution ###
To fetch the code and build the microservice execute the following:


```
cd $GOPATH/src
go get github.com/edgexfoundry/edgex-go
cd $GOPATH/src/github.com/edgexfoundry/edgex-go
# pull the 3rd party / vendor packages
make prepare
# build the microservice
make cmd/support-scheduler/support-scheduler
# get to the support logging microservice executable
cd cmd/support-scheduler
# run the microservice (may require other dependent services to run correctly)
./support-scheduler
```
To test, simple run:

```
cd $GOPATH/src
go get github.com/edgexfoundry/edgex-go
# pull the 3rd party / vendor packages
make prepare
# build the microservice
make cmd/support-scheduler/support-scheduler
# exectute the go test(s)
go test
```


## Community
- Chat: [https://edgexfoundry.slack.com](https://join.slack.com/t/edgexfoundry/shared_invite/enQtNDgyODM5ODUyODY0LWVhY2VmOTcyOWY2NjZhOWJjOGI1YzQ2NzYzZmIxYzAzN2IzYzY0NTVmMWZhZjNkMjVmODNiZGZmYTkzZDE3MTA)
- Mainling lists: https://lists.edgexfoundry.org/mailman/listinfo

## License
[Apache-2.0](LICENSE)