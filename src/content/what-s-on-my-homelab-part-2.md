---
title: What's on my homelab? PART 2 - 3D printing
description: 'Services and applications running on my homelab - Part 2: 3D printing'
tags:
  - homelab
  - self-hosting
  - iot
  - 3d-printing
  - klipper
  - moonraker
  - mainsail
listed: true
---

In this second part of the series, I'll be sharing the services related to 3D printing.

I own a 3D printer, and I've been using it for a few years now. I've been through a few different setups, and I'm pretty happy with the current one.

## Hardware

Again, I use a [Raspberry Pi 3 B+](https://www.raspberrypi.com/products/raspberry-pi-3-model-b-plus/) as the main server for my 3D printing setup. It's an older model and therefore less powerful than the Raspberry Pi 4 or 5, but it's still powerful enough to run all the different services and applications I need. In addition, I get mine second-hand, so it's even more affordable.

## Software

The software stack is pretty straightforward, I used [Mainsail OS](https://docs-os.mainsail.xyz/) as the operating system for the Raspberry Pi, which comes with a few pre-installed services:
- [Klipper](https://www.klipper3d.org/)
- [Moonraker](https://moonraker.readthedocs.io/en/latest/)
- [Mainsail](https://mainsail.xyz/)

### Klipper

I switched from Marlin to Klipper because I find it to be more flexible and easier to configure. The main advantage of Klipper is that it offloads the computation of the kinematics from the printer mainboard to a more powerful computer, such as a Raspberry Pi, which allows for faster and more accurate calculations. Even though my printer is not made for speed, this upgrade made a noticeable difference in the print quality at higher speeds.

### Moonraker

Moonraker is a web server that provides a REST API to control Klipper.

Since a REST API is very versatile, it allows me to control the printer from any device, such as my phone with [MobileRaker](https://mobileraker.com/) or via Mainsail. In addition, it enables me to integrate the printer with other services, such as Home Assistant or Node-RED.

![Home assistant integration](/img/what-s-on-my-homelab-part-2/home-assistant.png)

### Mainsail

Mainsail is a web-based user interface for Moonraker. It's a very user-friendly interface that allows me to control the printer, upload files, and monitor the print progress.

![Mainsail dashboard](/img/what-s-on-my-homelab-part-2/mainsail.png)

Mainsail comes with a powerful gcode visualizer that allows me to see the print progress in real-time.

![Gcode viewer in Mainsail](/img/what-s-on-my-homelab-part-2/mainsail-gcode-viewer.png)