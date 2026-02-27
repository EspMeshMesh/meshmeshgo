# Firmware build and upload

```bash
mkdir esphome-mm
cd esphome-mm
# Create a new python virtual env inside the folder
python3 -m venv venv
# Active the virtual env
source venv/bin/activate
# Install esphome using pip
pip install esphome
```

Create a config file for your device like the example below and the relative
secrets.yaml file to store your passowords

```yaml
external_components:
  - source: github://EspMeshMesh/esphome-meshmesh@main
    components: [ meshmesh, network, socket ]

esphome:
  name: wroom32s3
  comment: test wroom32s3

esp32:
  board: esp32-s3-devkitc-1
  framework:
    type: esp-idf

logger:
  level: DEBUG
  baud_rate: 115200

api:
  reboot_timeout: 900s

ota:
  platform: meshmesh

meshmesh:
  channel: 3
  baud_rate: 0
  tx_buffer_size: 0
  password: !secret meshmesh_password

mdns:
  disabled: True
```

Connect the device to the USB portm compile and upload the firmware using the 
following esphome command. Change se name of the config file and the name of the serial port to fit your needs.

```bash
esphome run ./wroom32s3.yaml  --device /dev/ttyACM0  
```

Now the node is ready to be disocvered inside the mesh network.