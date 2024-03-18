---
title: What's on my homelab? PART 3 - Media
description: 'Services and applications running on my homelab - Part 3: Media'
tags:
  - homelab
  - self-hosting
  - media
  - plex
  - overseerr
  - sonarr
  - radarr
  - deluge
  - jackett
listed: true
---

After home automation and 3D printing, let's talk about media services. I've been using [Plex](https://www.plex.tv/) for a few years now, and I've been very happy with it. However, I've been adding a few more services to my stack to improve the experience and automate the process of adding and managing content.

## What's Plex?

<img src="https://cdn.jsdelivr.net/gh/walkxcode/dashboard-icons/png/plex-alt.png" alt="Plex logo" width="100px" align="left" style="margin-top: 0; margin-right: 1rem;"/>

[Plex](https://www.plex.tv/) is a media server that allows you to stream your media content to any device. It's a very user-friendly application that automatically organizes your media content and fetches metadata such as posters, summaries, and trailers from the internet. I also use it to share my library with friends and family. The only weak point of Plex is that it requires a subscription to unlock some features, such as offline sync or mobile streaming.

![Plex](/img/what-s-on-my-homelab-part-3/plex.png)

## Upgrading Plex

Without any upgrade, Plex requires you to manually manage the content to your library. I have been doing this for a few years, and it's a tedious process. I've been looking for a way to automate this process, and I found the following services:
- [Overseerr](https://overseerr.dev/)
- [Sonarr](https://sonarr.tv/)
- [Radarr](https://radarr.video/)
- [Deluge](https://deluge-torrent.org/)
- [Jackett](https://github.com/Jackett/Jackett)

### Overseerr

<img src="https://cdn.jsdelivr.net/gh/walkxcode/dashboard-icons/png/overseerr.png" alt="Overseerr logo" width="100px" align="left" style="margin-top: 0; margin-right: 1rem;"/>

[Overseerr](https://overseerr.dev/) is a web interface for managing requests for your media library. It allows users to request content, and it integrates with Sonarr and Radarr to automate the process of adding new content to your library. Before installing Overseerr, I regularly visited websites such as [IMDb](https://www.imdb.com/) or [The Movie Database](https://www.themoviedb.org/) to check for new releases and add them to my library. Overseerr is user-friendly and enables my friends and family to explore new content and request movies or TV shows that they want to watch.

Overseeer also has a notification system that can send emails or push notifications to users when their requested content is available.

### Sonarr & Radarr

<img src="https://cdn.jsdelivr.net/gh/walkxcode/dashboard-icons/png/sonarr.png" alt="Sonarr logo" width="50px" align="left" style="margin-top: 0; margin-right: 1rem;"/>
<img src="https://cdn.jsdelivr.net/gh/walkxcode/dashboard-icons/png/radarr.png" alt="Radarr logo" width="50px" align="left" style="margin-top: 0; margin-right: 1rem;"/>

[Sonarr](https://sonarr.tv/) and [Radarr](https://radarr.video/) are applications that automatically download TV shows and movies. They can be configured to search for new content and add it to your download client, such as Deluge. They are similar but, Sonarr is for TV shows and Radarr is for movies.

### Deluge

<img src="https://cdn.jsdelivr.net/gh/walkxcode/dashboard-icons/png/deluge.png" alt="Deluge logo" width="100px" align="left" style="margin-top: 0; margin-right: 1rem;"/>

[Deluge](https://deluge-torrent.org/) is a lightweight, cross-platform BitTorrent client.

It comes with a web interface to manage torrents, manage the download queue, and monitor the download progress.

**Torrent is a peer-to-peer file sharing protocol used for distributing large amounts of data between multiple users without the need for a central server, which means no single authority has control over the data. This also means that the data could be unavailable if there are no peers sharing it.**

### Jackett

[Jackett](https://github.com/Jackett/Jackett) works as a proxy server: it translates queries from apps (Sonarr or Radarr) into tracker-site-specific HTTP queries, parses the HTML response, then sends results back to the requesting software.

## The automation process

The process to request, download and stream new content is as follows:

1. A user requests a new TV show or movie on Overseerr
2. Overseerr sends the request to Sonarr or Radarr depending on the type of content
3. Sonarr or Radarr searches for the content on the internet using Jackett as an interface to the different torrent trackers
4. Once the content is found, Sonarr or Radarr sends the request to Deluge to download the content
5. Once the download is complete, Sonarr or Radarr moves the content to the appropriate folder
6. Plex scans this folder and adds the content to the library
7. Plex handles the streaming to any device

