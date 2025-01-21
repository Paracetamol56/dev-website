---
title: My journey with Arch Linux
description: A detailed walkthrough of a modern Arch Linux desktop configuration using Hyprland
tags:
  - linux
  - ricing
  - archlinux
release: 2025-01-21T21:17:00.000Z
listed: true
---

> This post is directly linked to my dotfiles repository, which you can find at [GitHub](https://github.com/Paracetamol56/dotfiles). If you're interested in exploring my configuration further, feel free to check it out!

As a Linux enthusiast, I've recently taken the step to open-source my dotfiles configuration. While this is very much a work in progress, I'm excited to share my current setup and the reasoning behind my choices. Let me walk you through the key components that make up my daily computing environment.

<img class="clear-left max-w-7xl" src="/img/my-journey-with-arch-linux/desktop.png" alt="My desktop" />

## 🏗️ The Foundation: Why **Arch Linux**?

After more than three years of daily use, **Arch Linux** has become more than just my operating system – it's a fundamental part of my computing philosophy. The manual installation process, while intimidating to some, gave me something invaluable: intimate knowledge of every package on my system. This transparency isn't just about control; it's about understanding my tools at a fundamental level.

### Rolling Release and Community Support

The rolling release model perfectly aligns with my desire to be on the bleeding edge. This is a crucial aspect that requires careful consideration before choosing **Arch**. Running the latest packages means accepting the responsibility of managing potential updates that could affect system stability. Fortunately, I personally haven't encountered major issues in my journey. But when issues do arise, they become opportunities to engage with one of the most knowledgeable and helpful communities in **Linux**.

### **AUR**: Expanding Software Availability

And let's not forget the game-changer that is the **[Arch User Repository (AUR)](https://aur.archlinux.org/)**. Having virtually any software at my fingertips through the **AUR** has transformed how I think about software availability on **Linux**.

## 🪟 Window Management: Embracing **Hyprland**

My transition to **[Hyprland](https://hyprland.org/)** represents more than just a change in window managers – it's a fundamental shift in how I think about desktops. After years of using **GNOME** both professionally and personally, I began to question the traditional desktop paradigm. Traditional desktop environments like **GNOME** inherit a decades-old metaphor where **Microsoft** attempted to translate physical office objects into virtual space: the desktop, windows that float and minimize, a recycle bin, document folders. While this made computers more accessible in the 1990s, we're no longer constrained by these skeuomorphic designs.

**Hyprland**, like **i3** and other tiling window managers, breaks free from these legacy concepts. Instead of mimicking a physical desk, it treats your screen space as what it truly is: a tree of virtual workspaces that can be organized hierarchically and navigated efficiently. This approach, combined with **Hyprland**'s modern features as a **Wayland** compositor, provides:

- A more logical and efficient use of screen real estate through tiling 📐
- Workspace management based on actual workflow rather than physical metaphors
- **GPU** acceleration for smooth performance 🎮
- Seamless animations that provide visual feedback without sacrificing efficiency

## 🖥️ Terminal-Centric Workflow

### **Neovim**: My Text Editing Powerhouse

At the heart of my terminal-centric workflow lies **[Neovim](https://neovim.io/)**, my primary tool for text editing and development. As a modern take on the venerable **Vim** editor, **Neovim** provides an incredibly powerful and extensible environment that has revolutionized how I interact with code and text. Its modal editing, combined with a rich ecosystem of plugins, makes text manipulation feel like speaking a natural language.

### **Kitty**: My Terminal Emulator

The choice of **[Kitty](https://sw.kovidgoyal.net/kitty/)** as my terminal emulator was straightforward and practical: it's a minimal, **GPU-accelerated** terminal that's officially recommended in the **Hyprland** documentation. No bells and whistles needed... It simply does its job of being a reliable terminal emulator.

My terminal environment is enhanced by a carefully chosen set of modern **TUI (Terminal User Interface)** applications:

- **[Broot](https://github.com/Canop/broot)** for efficient file navigation and tree exploration 🌲
- **[Bat](https://github.com/sharkdp/bat)** as a modern replacement for **cat**, offering syntax highlighting and **Git** integration
- **[Neofetch](https://github.com/dylanaraps/neofetch)** for system information display 🖼️
- **[Btop](https://github.com/aristocratos/btop)** for system monitoring and resource management 📊

These tools combine the power of traditional command-line utilities with modern user interfaces, making terminal work both efficient and visually appealing.

## 🎨 User Interface Elements

### **Waybar**: More Than Just a Status Bar

Inspired by the **[Hyprrice](https://github.com/sejjy/hyprrice)** project by **sejjy**, my **Waybar** configuration strikes a balance between information density and visual appeal. It displays essential system information while maintaining a clean look that doesn't distract from my work.

### **Rofi**: A Swiss Army Knife

While many know **[Rofi](https://github.com/davatorium/rofi)** primarily as an application launcher, in my setup it serves as a central hub for system interaction. This versatile tool has become an indispensable part of my workflow, handling numerous tasks through its menu system:

- Application launching with instant search capabilities 🔍
- Network management for both **WiFi** and **Bluetooth** connections 📶
- Clipboard history, which has proven to be an incredible productivity booster 📋
- System control through a custom power menu for **shutdown**, **restart**, and **lock** options

### **Catppuccin Mocha**: Aesthetic and Comfort

I decided to go with **[Catppuccin](https://github.com/catppuccin/catppuccin)** because I am really enjoying the colors and how rich the theme is, with nearly **400 ports** at this time. Additionally, I am part of the community as a **maintainer**. The visual theme of my setup centers around the **Catppuccin Mocha** color scheme, with blue and mauve accents. This wasn't just an aesthetic choice – the color palette provides excellent contrast and reduces eye strain during long coding sessions. The wallpapers, sourced from the **[orangci/walls-catppuccin-mocha](https://github.com/orangci/walls-catppuccin-mocha)** collection, complement this theme perfectly.

## 🔧 Work in Progress

This setup is constantly evolving. As I discover new tools and refine my workflow, I expect to make changes and improvements. The current configuration represents a snapshot of what works well for me right now, but I'm always open to trying new approaches and tools.

## 🌍 Sharing and Collaboration

I've made my **dotfiles** public on **GitHub** because I believe in the power of sharing and learning from the community. I also want to acknowledge the wonderful **[r/unixporn](https://www.reddit.com/r/unixporn/)** community on **Reddit**, which has been incredibly helpful in my **ricing** journey.

## 🔮 Looking Forward

As this project continues to evolve, I plan to:

- Refine my keybindings for even better efficiency ⌨️
- Explore additional **Wayland-native** applications
- Document more of my custom scripts and configurations

Additionally, I am actively thinking about replacing my OS with **[NixOS](https://nixos.org/)** because I find its concept very interesting.

For those interested in exploring my setup further, you can find all the configuration files in my **GitHub** repository. I welcome feedback and suggestions from the community as I continue to refine this setup. 🤝
