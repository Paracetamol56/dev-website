---
title: Temperature history
description: How temperatures have changed through the year since 1940, anywhere on Earth
tags:
  - climate
  - data-science
  - data-visualization
release: 2026-10-18T12:00:00.000Z
listed: true
---

<script>
  import TemperatureHistory from './temperature-history/TemperatureHistory.svelte';
</script>

Global warming is often summed up in a single number, but nobody lives in a global average. Search any place below to see how its temperatures have changed through the year since 1940.

The data comes from [ERA5](https://cds.climate.copernicus.eu/datasets/reanalysis-era5-single-levels), a reconstruction of past weather across the whole planet by the European Copernicus programme.

Each line of the radar chart is a year, running clockwise from January: the further out, the warmer. Drag the timeline handles to pick the years to compare, or press play to watch them appear one by one.

<TemperatureHistory />

## Resources

- [ERA5 on the Copernicus Climate Data Store](https://cds.climate.copernicus.eu/datasets/reanalysis-era5-single-levels)
- [Hersbach et al. (2020), The ERA5 global reanalysis](https://doi.org/10.1002/qj.3803)
- [Open-Meteo historical weather API](https://open-meteo.com/en/docs/historical-weather-api), which serves ERA5 for this page
- [Ed Hawkins' climate spiral](https://www.climate-lab-book.ac.uk/spirals/), the inspiration for the radar chart
