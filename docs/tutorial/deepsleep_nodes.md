# Deep Sleep Nodes

## Overview

EspHome devices with the deep_sleep component support allows to put the device in an low power mode for a specified period of time. In this way such node can be powered by batteries and or solar panels.

## Usage

Deep sleep mode is automatically enabled when the deep_sleep component is present in the configuration file. The EnableZeroConf option must be enabled in the meshmeshgo HUB to allow HomeAssistant to reconnect to the node as soon as it exits low-power mode.

## Example

```yaml
external_components:
  - source: github://EspMeshMesh/esphome-meshmesh@main

esphome:
  name: deep-sleep-node
  friendly_name: Depp Sleep Node
  platformio_options:
    board_build.flash_mode: dio

esp32:
  board: esp32-c3-devkitm-1
  variant: ESP32C3
  framework:
    type: esp-idf

deep_sleep:
  id: deep_sleep_1
  run_duration: 20s
  sleep_duration: 5min

logger:
  level: DEBUG
  hardware_uart: USB_CDC

mdns:
  disabled: True

socket:
  implementation: meshmesh_esp32

meshmesh:
  baud_rate: 0
  rx_buffer_size: 0
  tx_buffer_size: 0
  password: !secret meshmesh_password
  channel: 3
  use_starpath: True

switch:
  - platform: template
    id: deep_sleep_disable
    name: "Disable deep_sleep"
    restore_mode: ALWAYS_OFF
    optimistic: true
    turn_on_action:
      - deep_sleep.prevent: deep_sleep_1
  - platform: template
    id: deep_sleep_enable
    name: "Enable deep_sleep"
    restore_mode: ALWAYS_OFF
    optimistic: true
    turn_on_action:
      - deep_sleep.enter: deep_sleep_1

api:
  reboot_timeout: 0s

ota:
  platform: esphome

sensor:
  - platform: uptime
    name: "Uptime"
    id: uptime_sensor
  - platform: dht
    pin: GPIO4
    model: DHT22
    temperature:
      id: temperature
      name: "Temperature"
    humidity:
      id: humidity
      name: "Humidity"
    update_interval: 60s

light:
  - platform: status_led
    name: "Status LED"
    id: esp_status_led
    icon: "mdi:alarm-light"
    pin:
      number: GPIO8
      inverted: true
    restore_mode: ALWAYS_OFF
```