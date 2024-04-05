---
title: What's on my homelab? PART 1 - Home automation
description: 'Services and applications running on my homelab - Part 1: Home automation'
tags:
  - homelab
  - self-hosting
  - iot
  - home-automation
  - home-assistant
  - node-red
  - mosquitto
  - esphome
listed: true
---

First part of a series of articles where I'll be sharing the services and applications running on my homelab.
Home automation is one of the main use cases for my homelab. I use it to control all the different smart devices in my home, such as lights, and security cameras, power switches, or thermostats. I also use it to monitor energy consumption and to keep track of the weather.
One of the main reasons I decided to build my own home automation system is privacy. I didn't want to rely on cloud services to control my smart devices, and I wanted to have full control over my data. I also wanted to be able to customize my setup to fit my specific needs and to integrate all my different devices and services together.

## Hardware

I use a [Raspberry Pi 4](https://www.raspberrypi.com/products/raspberry-pi-4-model-b/) as the main server for my home automation setup. It's a small and affordable computer that's powerful enough to run all the different services and applications I need. I also use a few ESP8266 and ESP32 microcontrollers to build custom sensors and actuators.

**Specifications:**
- Raspberry Pi 4 Model B
- 4GB RAM
- 64GB microSD card flashed with Debian

## Software

<img class="content-ignore float-left w-20 mr-4 mb-2" src="https://user.oc-static.com/upload/2021/11/10/1636542639252_Moby-logo.png" alt="Docker logo" />
I use docker to run all the different services and applications on my Raspberry Pi. Docker is a great way to manage and run services in isolated containers, which makes it easy to keep everything organized and up to date.

[Docker official website](https://www.docker.com/)

<div class="clear-left">

### Home Assistant

<img class="content-ignore float-left w-20 mr-4 mb-2" src="https://cdn.jsdelivr.net/gh/walkxcode/dashboard-icons/png/home-assistant.png" alt="Home Assistant logo" />

[Home Assistant](https://www.home-assistant.io/) is the heart of my home automation setup. It's the central hub that connects all the different devices and services together.

In my opinion, Home Assistant is the best open-source home automation platform out there. It's easy to use, highly customizable, and has a large and active community. It supports a wide range of different devices and services, and it's easy to extend and customize with custom scripts and automations.

In addition to the core Home Assistant server, I also use a few add-ons to extend its functionality:
- [Node-RED](https://nodered.org/)
- [Mosquitto](https://mosquitto.org/)
- [ESPHome](https://esphome.io/)

</div>

<img class="clear-left max-w-7xl" src="/img/what-s-on-my-homelab-part-1/home-assistant.png" alt="Home Assistant dashboard" />

<div class="clear-left">

#### Installation notes

Home Assistant comes in many different flavors, the recommended way to install it is using the Home Assistant OS image, which is a pre-configured operating system with Home Assistant and all its dependencies pre-installed. However, I prefer to use the Home Assistant Supervised installation method, which allows me to run Home Assistant on top of a regular Linux distribution, such as Debian, and to install other services and applications alongside it.

</div>
<div class="clear-left">

### Node-RED

Node-RED is a visual programming tool that I use to create automations and scripts for Home Assistant. It's a great way to create complex automations without having to write a lot of code.

<img class="max-w-7xl" src="/img/what-s-on-my-homelab-part-1/node-red.png" alt="Node-RED editor" />

</div>
<div class="clear-left">

### Mosquitto

<img class="content-ignore float-left w-20 mr-4 mb-2" src="https://cdn.jsdelivr.net/gh/walkxcode/dashboard-icons/png/eclipse-mosquitto.png" alt="Mosquitto logo" />

Mosquitto is an open source MQTT broker that I use to connect all my IoT devices together. It's an efficient way to send messages between different devices and services. I use it to send sensor data from my custom sensors to Home Assistant.

*MQTT ([Message Queuing Telemetry Transport](https://en.wikipedia.org/wiki/MQTT)) is a lightweight, publish-subscribe, machine to machine network protocol for message queue/message queuing service. It is designed for connections with remote locations that have devices with resource constraints or limited network bandwidth, such as in the Internet of Things (IoT).*

</div>
<div class="clear-left">

### ESPHome

<img class="content-ignore float-left w-20 mr-4 mb-2" src="https://cdn.jsdelivr.net/gh/walkxcode/dashboard-icons/png/esphome.png" alt="ESPHome logo" />

ESPHome is a system to program ESP8266/ESP32 boards by simple yet powerful `.yaml` configuration files and control them remotely through the Home Assistant API.

</div>
