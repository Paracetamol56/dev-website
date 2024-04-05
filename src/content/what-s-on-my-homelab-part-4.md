---
title: What's on my homelab? PART 4 - Everyday services
description: 'Services and applications running on my homelab - Part 4: Everyday services'
tags:
  - homelab
  - self-hosting
  - homepage
  - adguard-home
  - bitwarden
  - nextcloud
  - immich
  - n8n
  - opengist
listed: true
---

<div class="clear-left">

## Homepage

Let's start this fourth part with the front door of my homelab. [Homepage](https://gethomepage.dev/latest/) is a dashboard that I use to access all my services. It's a simple and customizable web interface that I use to access all my services. It's a great way to have a quick overview of what's running on my server and access all my services with a single click. I have been using [Homer](https://github.com/bastienwirtz/homer) and [Heimdall](https://heimdall.site/) in the past, but I recently switched to Homepage because it's ligtweight but heavily customizable. Perfect for my needs.

<img class="max-w-7xl" src="/img/what-s-on-my-homelab-part-4/homepage.png" alt="Homepage" />

</div>
<div class="clear-left">

## AdGuard Home

<img class="content-ignore float-left w-20 mr-4 mb-2" src="https://cdn.jsdelivr.net/gh/walkxcode/dashboard-icons/png/adguard-home.png" alt="AdGuard Home logo" />

[AdGuard Home](https://adguard.com/en/adguard-home/overview.html) is a network-wide software for blocking ads & tracking. After using [Pi-hole](https://pi-hole.net/) for a few years, I decided to switch to AdGuard Home because it offers more features and a better user interface. AdGuard Home is a DNS server that blocks ads, trackers, and malware domains on your network. It also has a web interface to manage the settings and view statistics. Obviously, it's far for perfect, and some ads still slip through, but it's a good start.

</div>
<div class="clear-left">

## Bitwarden

<img class="content-ignore float-left w-20 mr-4 mb-2" src="https://cdn.jsdelivr.net/gh/walkxcode/dashboard-icons/png/bitwarden.png" alt="Bitwarden logo" />

[Bitwarden](https://bitwarden.com/) is an open-source password manager. I use the Rust implementation of Bitwarden, which is called [Vaultwarden](https://github.com/dani-garcia/vaultwarden). After using the free tier of Bitwarden for a few years, I decided to self-host it because to remove my passwords from the cloud. Vaultwarden is of course compatible with the official Bitwarden clients: browser extensions, mobile and desktop apps.

</div>
<div class="clear-left">

## Nextcloud

<img class="content-ignore float-left w-20 mr-4 mb-2" src="https://cdn.jsdelivr.net/gh/walkxcode/dashboard-icons/png/nextcloud.png" alt="Nextcloud logo" />

[Nextcloud](https://nextcloud.com/) is a self-hosted file sync and share server with a powerful extension system. I use Nextcloud to store my files, contacts and calendars. It's a great alternative to Google Drive, OneDrive, or Dropbox especially if you care about privacy and if you mount a dedicated storage to your server.

</div>
<div class="clear-left">

## Immich

<img class="content-ignore float-left w-20 mr-4 mb-2" src="https://cdn.jsdelivr.net/gh/walkxcode/dashboard-icons/png/immich.png" alt="Immich logo" />

One of the best service I host... [Immich](https://immich.app/) is a photo and video gallery that I use to store and share my memories. In my opinion, it's the best alternative to Google Photos I've found so far. It's still under heavy development, but it's already a great service.

<img class="max-w-7xl" src="/img/what-s-on-my-homelab-part-4/immich-home.png" alt="Immich home" />

The killer feature of Immich is the map view. Pictures metadata are scanned to extract the GPS coordinates and display them on a map.

<img class="max-w-7xl" src="/img/what-s-on-my-homelab-part-4/immich-map.png" alt="Immich map" />

In addition, Immich has a great search engine powered by powerful ai algorithms. It can recognize objects, people, and places in your pictures.

<img class="max-w-7xl" src="/img/what-s-on-my-homelab-part-4/immich-search.png" alt="Immich search" />

</div>
<div class="clear-left">

## N8N

<img class="content-ignore float-left w-20 mr-4 mb-2" src="https://cdn.jsdelivr.net/gh/walkxcode/dashboard-icons/png/n8n.png" alt="N8N logo" />

Simple but powerful workflow automation tool. [N8N](https://n8n.io/) is a free and open-source alternative to Zapier, Integromat, or Microsoft Power Automate. It allows you to automate workflows between apps and services. I use it occasionally to automate tasks such as sending notifications or create small bots.

</div>
<div class="clear-left">

## Opengist

<img class="content-ignore float-left w-20 mr-4 mb-2" src="https://cdn.jsdelivr.net/gh/walkxcode/dashboard-icons/png/opengist.png" alt="Opengist logo" />

A simple pastebin service. [Opengist](https://opengist.app/) is a minimalist pastebin app that I use to share code snippets or text. It has nothing more than GitHub Gist or Pastebin, but it's self-hosted... and that's cool!

</div>