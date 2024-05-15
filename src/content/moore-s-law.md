---
title: Moore's Law
description: Moore's Law in numbers
tags:
  - computer-science
  - data-science
  - data-visualization
listed: true
---

<script> 
  import Chart from './moore-s-law/Chart.svelte';
</script>

The Moore's Law is an observation made by Gordon Moore, co-founder of Intel, in 1965 :
> The number of transistors in a dense integrated circuit doubles approximately every two years.
His observation has held true for decades and has been a driving force behind the exponential growth of computing power.

The chart below shows how the number of transistors, the gate size, the frequency or TDP have evolved over the years from 2000 to 2024.

#### Transistor count

Transistor count is a pretty good measure of the computing power of a chip. The more transistors, the more complex the chip can be and the more operations it can perform per second. Transistor count and frequency directly determine the performance of a chip measured in FLOPS (Floating Point Operations Per Second).

$FLOPS = TransistorCount * Frequency$

#### Gate size

The gate is the part of the transistor that controls the flow of electricity. The gate size determines how many transistors can be packed into a given area. The smaller the gate size, the more transistors can be packed into a chip, and the more powerful the chip can be. The gate size is measured in nanometers (nm) and has been shrinking over the years as technology has advanced. Nowadays, the gate size is around 7nm (a silicon atom is about 0.2nm so the gate size is about 35 atoms wide).

The constant shrinking of the gate size has been a major factor in the exponential growth of computing power and it shows the incredible progress that has been made in the field of semiconductor manufacturing and photolithography.

#### Frequency

The frequency is the number of clock cycles per second (measured in Hertz) at which a chip operates. As shown in the chart, the frequency has been increasing until the early 2000s and has been relatively stable since then due to physical limitations such as power consumption and heat dissipation. That's why chip manufacturers decided to put multiple cores on a single chip to increase performance.

#### TDP

TDP stands for Thermal Design Power and is a measure of the maximum amount of heat that a chip can generate under load. The TDP is measured in watts and is an important factor to consider when designing a cooling solution for a chip.

As you can imagine, the TDP is an important constraint to take into account for consumer electronics. Especially for mobile devices and embedded systems.

<Chart />

### Data source

The data used in this chart comes from [Yifan Sun](https://sarchlab.org/syifan) who created a similar [chart](https://chip-dataset.vercel.app/) and published his dataset.

### Ressources

- [Wikipedia article on Moore's Law](https://en.wikipedia.org/wiki/Moore%27s_law)
- [WikiChip](https://en.wikichip.org/wiki/moore%27s_law)
- [Yifan Sun's research](https://arxiv.org/abs/1911.11313)
