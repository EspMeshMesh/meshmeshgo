# Running with Docker

This guide explains how to run the meshmeshgo program in Docker. The app requires access to:
1. Serial device (e.g., /dev/ttyUSB0)


## 1. Pull the prebuilt Docker Image

From the project root (where the Dockerfile is located):

```bash
docker pull ghcr.io/espmeshmesh/amd64-meshmeshgo-addon:latest
```


## 2. Run with `docker run`

> NB This command will tell docker to use the current working directory as storage for persistent files like 

```bash
docker run --rm -it --network host --privileged -v $(pwd):/data/ --device /dev/ttyUSB0 amd64-meshmeshgo-addon:latest

```

If you need to change some config values create the an options.json file in your persistent folder and modify the corresponding value.

> If there is no options.json in the persistent data folder a default one will be created. 

```json
{
  "WantHelp": false,
  "DataFolder": "",
  "SerialPortName": "/dev/ttyUSB0",
  "SerialPortBaudRate": 460800,
  "SerialIsEsp8266": false,
  "SerialShouldRetry": true,
  "SerialResetOnInit": false,
  "VerboseLevel": 0,
  "TargetNode": 0,
  "DebugNodeAddr": "",
  "RestBindAddress": ":4040",
  "RpcBindAddress": "",
  "BindAddress": "dynamic",
  "BindPort": 6053,
  "BasePortOffset": 20000,
  "SizeOfPortsPool": 10000,
  "EnableZeroconf": false
}
```


