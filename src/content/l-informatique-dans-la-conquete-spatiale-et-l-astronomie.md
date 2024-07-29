---
title: L'informatique dans la conquête spatiale et l'astronomie
description: Annex to the talk "Computers in Space and Astronomy" given by Mathéo Galuba on the 19th of July 2024 at the observatory of Orion Astronomie du velay. 
tags:
  - computer-science
  - astronomy
  - talk
listed: false
---

<script>
  import CodeBlock from '$lib/components/CodeBlock.svelte';
  import Button from '$lib/components/Button.svelte';
  import { Projector, FileImage, FileCode2, FileText } from 'lucide-svelte';
  import SyncVSAsync from './computers-in-space-and-astronomy/SyncVSAsync.svelte'
  import MarsMissions from './computers-in-space-and-astronomy/MarsMissions.svelte'
</script>

<section id="files" class="flex gap-2 flex-wrap">
  <a
    class="flex items-center gap-1 rounded-md bg-ctp-mauve px-3 py-1
      font-semibold text-ctp-mantle content-ignore
      shadow-md shadow-ctp-crust transition-opacity"
    href="/files/l-informatique-dans-la-conquête-spatiale-et-l-astronomie/slide.html"
    target="_blank"
  >
    <Projector size="16" />
    Slide
  </a>
  <a
    class="flex items-center gap-1 rounded-md bg-ctp-mauve px-3 py-1
      font-semibold text-ctp-mantle content-ignore
      shadow-md shadow-ctp-crust transition-opacity"
    href="/files/l-informatique-dans-la-conquête-spatiale-et-l-astronomie/slide.pdf"
    download
  >
    <FileImage size="16" />
    Slide as PDF
  </a>
  <a
    class="flex items-center gap-1 rounded-md bg-ctp-mauve px-3 py-1
      font-semibold text-ctp-mantle content-ignore
      shadow-md shadow-ctp-crust transition-opacity"
    href="/files/l-informatique-dans-la-conquête-spatiale-et-l-astronomie/slide.md"
    download
  >
    <FileCode2 size="16" />
    Slide as Markdown
  </a>
  <a
    class="flex items-center gap-1 rounded-md bg-ctp-mauve px-3 py-1
      font-semibold text-ctp-mantle content-ignore
      shadow-md shadow-ctp-crust transition-opacity"
    href="/files/l-informatique-dans-la-conquête-spatiale-et-l-astronomie/presentation-conference-dark.pdf"
    download
  >
    <FileText size="16" />
    Presentation (Dark)
  </a>
  <a
    class="flex items-center gap-1 rounded-md bg-ctp-mauve px-3 py-1
      font-semibold text-ctp-mantle content-ignore
      shadow-md shadow-ctp-crust transition-opacity"
    href="/files/l-informatique-dans-la-conquête-spatiale-et-l-astronomie/presentation-conference-light.pdf"
    download
  >
    <FileText size="16" />
    Presentation (Light)
  </a>
</section>

## Introduction

Modèle de processeur utilisé pour l'exemple : [Intel Core i9-13900K](https://ark.intel.com/content/www/fr/fr/ark/products/230496/intel-core-i9-13900k-processor-36m-cache-up-to-5-80-ghz.html).

Article sur la [Loi de Moore](moore-s-law).

<details>
  <summary>
    Code source du programme utilisé pour démontrer la puissance de calcul d'un ordinateur moderne
  </summary>

<CodeBlock>

```c
#include <gmp.h>
#include <stdio.h>
#include <stdlib.h>
#include <time.h>

int main(int argc, char *argv[]) {
  // Get the limit from the first argument
  if (argc != 2) {
    printf("Wrong input.\n");
    printf("Usage: howfast <limit>\n");
    return EXIT_FAILURE;
  }

  long int limit = strtol(argv[1], NULL, 10);

  // Setup GMP arithmetic library
  mpz_t a;
  mpz_t b;
  mpz_t c;

  mpz_init_set_ui(a, 1);
  mpz_init_set_ui(b, 0);
  mpz_init(c);

  // Start timer
  const clock_t start = clock();

  // Fibonacci calculation
  for (long int i = 0; i < limit; ++i) {
    mpz_add(c, a, b);
    mpz_set(a, b);
    mpz_set(b, c);
  }

  // End timer
  const clock_t end = clock();

  // Print the result
  printf("Fibonacci number %ld: ", limit);
  mpz_out_str(stdout, 10, b);
  printf("\n");

  // Clean up
  mpz_clear(a);
  mpz_clear(b);
  mpz_clear(c);

  // Print time
  double time_taken = ((double)(end - start)) / CLOCKS_PER_SEC;
  printf("Time: %f seconds\n", time_taken);
  return EXIT_SUCCESS;
}
```

</CodeBlock>

</details>

## 1. Informatique et astronomie



### Observer l'Univers

Image très haute résolution de la galaxie d'Andromède :

<iframe
  src="https://esahubble.org/images/heic1502a/zoomable/"
  class="w-full max-w-2xl mx-auto aspect-video"
  title="Hubble Space Telescope"
  frameborder="0"
></iframe>

Source : [Hubble Space Telescope](https://esahubble.org/images/heic1502a/zoomable/)

### Comprendre l'Univers

Vidéo de la première simulation informatique d'une galaxie, réalisée en 1960 :

<video controls class="w-full max-w-2xl mx-auto">
  <track kind="captions">
  <source src="/files/l-informatique-dans-la-conquête-spatiale-et-l-astronomie/assets/23.mp4" type="video/mp4">
</video>

<br>

Vidéo d'une simulation moderne de formation d'une galaxie massive :

<video controls class="w-full max-w-2xl mx-auto">
  <track kind="captions">
  <source src="/files/l-informatique-dans-la-conquête-spatiale-et-l-astronomie/assets/24.mp4" type="video/mp4">
</video>

## 2. Informatique et exploration spatiale

### 2.1. L'histoire conjointe de l'informatique et de la course à l'espace 

<SyncVSAsync />

<MarsMissions />

### 2.2. Perspectives futures

