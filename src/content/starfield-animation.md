---
title: Starfield animation
description: Implementing a starfield animation in C unsing Raylib
tags:
  - c
  - raylib
  - graphics
  - creative-coding
listed: true
---

<script>
  import CodeBlock from '$lib/components/CodeBlock.svelte';
</script>

Let's implement a starfield animation in C using Raylib.
This is a very simple project, useful to learn how to use a new library or to get started with graphics programming and creative coding.

<iframe src="/files/starfield-animation/starfield-animaiton.html" title="Starfield simulation" class="aspect-video mx-auto w-full max-w-3xl"></iframe>

<div class="mt-4 flex justify-center">
  <a
    target="_blank"
    href="https://paracetamol56.itch.io/starfield-animation"
    class="content-ignore flex items-center gap-1 rounded-md bg-ctp-mauve px-3 py-1
      font-semibold text-ctp-mantle
      shadow-md shadow-ctp-crust transition-opacity
      hover:opacity-80 active:opacity-60"
  >
    <span>Run on Itch.io</span>
    <svg class="square-4 fill-current" role="img" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><title>Itch.io</title><path d="M3.13 1.338C2.08 1.96.02 4.328 0 4.95v1.03c0 1.303 1.22 2.45 2.325 2.45 1.33 0 2.436-1.102 2.436-2.41 0 1.308 1.07 2.41 2.4 2.41 1.328 0 2.362-1.102 2.362-2.41 0 1.308 1.137 2.41 2.466 2.41h.024c1.33 0 2.466-1.102 2.466-2.41 0 1.308 1.034 2.41 2.363 2.41 1.33 0 2.4-1.102 2.4-2.41 0 1.308 1.106 2.41 2.435 2.41C22.78 8.43 24 7.282 24 5.98V4.95c-.02-.62-2.082-2.99-3.13-3.612-3.253-.114-5.508-.134-8.87-.133-3.362 0-7.945.053-8.87.133zm6.376 6.477a2.74 2.74 0 0 1-.468.602c-.5.49-1.19.795-1.947.795a2.786 2.786 0 0 1-1.95-.795c-.182-.178-.32-.37-.446-.59-.127.222-.303.412-.486.59a2.788 2.788 0 0 1-1.95.795c-.092 0-.187-.025-.264-.052-.107 1.113-.152 2.176-.168 2.95v.005l-.006 1.167c.02 2.334-.23 7.564 1.03 8.85 1.952.454 5.545.662 9.15.663 3.605 0 7.198-.21 9.15-.664 1.26-1.284 1.01-6.514 1.03-8.848l-.006-1.167v-.004c-.016-.775-.06-1.838-.168-2.95-.077.026-.172.052-.263.052a2.788 2.788 0 0 1-1.95-.795c-.184-.178-.36-.368-.486-.59-.127.22-.265.412-.447.59a2.786 2.786 0 0 1-1.95.794c-.76 0-1.446-.303-1.948-.793a2.74 2.74 0 0 1-.468-.602 2.738 2.738 0 0 1-.463.602 2.787 2.787 0 0 1-1.95.794h-.16a2.787 2.787 0 0 1-1.95-.793 2.738 2.738 0 0 1-.464-.602zm-2.004 2.59v.002c.795.002 1.5 0 2.373.953.687-.072 1.406-.108 2.125-.107.72 0 1.438.035 2.125.107.873-.953 1.578-.95 2.372-.953.376 0 1.876 0 2.92 2.934l1.123 4.028c.832 2.995-.266 3.068-1.636 3.07-2.03-.075-3.156-1.55-3.156-3.025-1.124.184-2.436.276-3.748.277-1.312 0-2.624-.093-3.748-.277 0 1.475-1.125 2.95-3.156 3.026-1.37-.004-2.468-.077-1.636-3.072l1.122-4.027c1.045-2.934 2.545-2.934 2.92-2.934zM12 12.714c-.002.002-2.14 1.964-2.523 2.662l1.4-.056v1.22c0 .056.56.033 1.123.007.562.026 1.124.05 1.124-.008v-1.22l1.4.055C14.138 14.677 12 12.713 12 12.713z"/></svg>
  </a>
</div>

## Inspiration

I got the idea for this project from [this video](https://www.youtube.com/watch?v=17WoOqgXsRM) by [The Coding Train](https://www.youtube.com/user/shiffman).
In the video, Daniel Shiffman implements a starfield animation in JavaScript using processing.
I thought it would be a fun project to implement the same animation in C, the language of gods.

## Set up

First of all, we need a bit of boilerplate code to set up the development environment, libraries, build system, etc.
Fortunately, Raylib provides a very simple and easy-to-use [repository template](https://github.com/raylib-extras/game-premake) that we can use to get started quickly.
It uses [Premake](https://premake.github.io/) as a build system (an other template using CMake is also available if you prefer).

Once you created your repository on GitHub using the template, clone it and download the raylib library as a git submodule:

<CodeBlock>

```bash
git clone git@git.matheo-galuba.com:mga/starfield-animaiton.git
cd starfield-animaiton
git submodule add git@github.com:raysan5/raylib.git raylib-master
```

</CodeBlock>

Then, you can build the project using the provided `premake-VirtualStudio.bat` or `premake-mingw.bat` scripts. Or directly using the premake binary file for your platform.

The last step is to get rid of the example code provided in the template.

Files to delete:

- `resources/*`
- `game/src/screens_ending.c`
- `game/src/screens_gameplay.c`
- `game/src/screens_logo.c`
- `game/src/screens_options.c`
- `game/src/screens_title.c`
- `game/src/screens.h`

In the same way, all the content of `/extras` can be deleted.

Last but not least, the `game/src/raylib_game.c` (which I renamed `main.c`) should be heavily modified to remove all the existing screen logic.
We only need to keep the main loop and the rendering logic.

## Implementation

### The Star struct

The first thing we need is a struct to represent a star.

Each star will have a position that will be randomly generated. I chose to define the position as a 3D vector, `x` and `y` are the initial position of the star and will not change until the star leaves the screen, `z` is the depth of the star and will continuously decrease to simulate the star moving towards the screen.

Next I added `sx` and `sy` to store the screen position of the star, 

Optionally, I added `prevX`, `prevY`, and `prevZ` to store the previous position of the star. This is useful to draw a line from the previous position to the current position of the star.

Finally, I added a `radius` to store the size of the star.

<CodeBlock>

```c
// star.h

#ifndef STAR_H
#define STAR_H
typedef struct Star
{
	int x;
	int y;
	int z;

	int sx;
	int sy;

	int prevX;
	int prevY;
	int prevZ;

	float radius;
} Star;

#endif // STAR_H
```

</CodeBlock>

### The star update function

The basic idea of the star animation is to use linear interpolation to calculate the current state of the star based on its z position.

For example the radius of the star will evolve from 0 to a maximum (4 in this case) as the star moves towards the screen.

<CodeBlock>

```c
// star.c

#include "star.h"

void updateStar(Star* star, const int screenWidth, const int screenHeight, float deltaTime)
{
	if (screenWidth == 0 || screenHeight == 0)
		return;

	star->z -= (int)(600.0f * deltaTime); // Decrease the z position of the star (we use a multiplier to control the speed of the animation)

    if (star->z < 1)
    {
      // Case where the star leaves the screen: reset its position
      star->x = getRandInt(-screenWidth, screenWidth);
      star->y = getRandInt(-screenHeight, screenHeight);
      star->z = max(screenWidth, screenHeight);

      star->prevZ = star->z;
    }

    // Calculate the radius of the star based on its z position
    star->radius = map(star->z, 0.0f, screenWidth, 4.0f, 0.0f);

    // Calculate the screen position of the star based on its z position
    star->sx = lerp(0.0f, screenWidth, (float)star->x / (float)star->z);
    star->sy = lerp(0.0f, screenHeight, (float)star->y / (float)star->z);

    // Calculate the previous screen position of the star based on its previous z position
    star->prevX = lerp(0.0f, screenWidth, (float)star->x / (float)star->prevZ);
	  star->prevY = lerp(0.0f, screenHeight, (float)star->y / (float)star->prevZ);
    star->prevZ = star->z;
}
```

</CodeBlock>

Obviously, we need to write a few helper functions to make this code work:

<CodeBlock>

```c
// star.c

#include <stdlib.h>

int getRandInt(int min, int max)
{
	return min + rand() / (RAND_MAX / (max - min));
}

float lerp(float a, float b, float t)
{
	return a + t * (b - a);
}

float map(float value, float x1, float y1, float x2, float y2)
{
	return x2 + (value - x1) * (y2 - x2) / (y1 - x1);
}

#ifndef max // Some compilers already have a max function
int max(int a, int b)
{
	return a > b ? a : b;
}
#endif
```

</CodeBlock>

Note that the `getRandInt` requires to set the seed of the random number generator at the beginning of the program:

<CodeBlock>

```c
// main.c

#include <stdio.h>
#include <stdlib.h>
#include <time.h>

int main()
{
    // Seed the random number generator
    srand((int)time(NULL));

    // ...
}
```

</CodeBlock>

### The star initialization function

We also probably want a function to initialize the star struct with random values:

<CodeBlock>

```c
// star.c

void initStar(Star* star)
{
	star->x = getRandInt(-GetScreenWidth(), GetScreenWidth());
	star->y = getRandInt(-GetScreenHeight(), GetScreenHeight());
	star->z = getRandInt(0, max(GetScreenWidth(), GetScreenHeight()));

	star->sx = 0;
	star->sy = 0;

	star->prevX = 0;
	star->prevY = 0;

	star->radius = 0.0f;
}
```

</CodeBlock>

### Star creation

Now that we defined what is a star and how to update it, we can create an array of stars and update them in the main loop.

The number of stars has no resons to be dynamic, so I used a macro to define the number of stars at compile time.

<CodeBlock>

```c
// main.c

#include "raylib.h"
#incldue "star.h"

#define STAR_COUNT 5000

Star stars[STAR_COUNT];

int main(void)
{
  // Initialization
  //---------------------------------------------------------
  for (int i = 0; i < STAR_COUNT; i++)
  {
    initStar(&stars[i]);
  }

  // ...
}
```

</CodeBlock>

### The main loop

Finally, we need to update and draw the stars in the main loop.

<CodeBlock>

```c
static void UpdateDrawFrame(void)
{
  // Update
  //----------------------------------------------------------------------------------
  for (int i = 0; i < STAR_COUNT; i++)
  {
    updateStar(&stars[i], s_screenWidth, s_screenHeight, GetFrameTime());
  }


  // Draw
  //----------------------------------------------------------------------------------
  BeginDrawing();
	ClearBackground(BLACK);

  for (int i = 0; i < STAR_COUNT; i++)
  {
    DrawCircle(stars[i].sx, stars[i].sy, stars[i].radius, WHITE);
    DrawLineEx((Vector2){stars[i].prevX, stars[i].prevY}, (Vector2){stars[i].sx, stars[i].sy}, stars[i].radius, WHITE);
  }

	// DrawFPS(10, 10);
  EndDrawing();
  //----------------------------------------------------------------------------------
}
```

</CodeBlock>

### Centering the starfield

Right now, the result looks like this:

![Starfield animation not centered](/img/starfield-animation/not-centered.png)

We never cared about centering the starfield origin, so the stars look like they are moving from the origin.

The simpler way to center the starfield is to use a Raylib 2D camera and offset the drawing of the stars by half the screen size.

<CodeBlock>

```c
Camera2D camera = {
	{WINDOW_WIDTH / 2.0f, WINDOW_HEIGHT / 2.0f},
    {0.0f, 0.0f},
    0.0f,
    1.0f
};

// Update and draw game frame
static void UpdateDrawFrame(void)
{
  // ...

  // Draw
  //----------------------------------------------------------------------------------
  BeginDrawing();
  ClearBackground(BLACK);

  BeginMode2D(camera);

  // ...
}
```

</CodeBlock>

And voilà! The starfield is now centered:

![Starfield animation centered](/img/starfield-animation/centered.png)

## Downloadable content

- Download the full source code [here](https://git.matheo-galuba.com/mga/starfield-animaiton).
- Download the compiled version for Windows [here](https://git.matheo-galuba.com/mga/starfield-animaiton/releases).
- Embed the animation in your website using the following code:

<CodeBlock>

```html
<iframe src="https://dev.matheo-galuba.com/files/starfield-animation/starfield-animaiton.html" title="Starfield simulation"></iframe>
```

</CodeBlock>

## Conclusion

For further improvements, here is a list of ideas:

- Control the speed of the animation with the mouse wheel
- Make the window resizable
- Export for web using Emscripten
- Enable multisampling to improve the rendering quality
- Add a background image

I hope you enjoyed this tutorial and that you learned something new.

## Ressources

- [Raylib](https://www.raylib.com/)
- [The Coding Train - Starfield animation](https://www.youtube.com/watch?v=17WoOqgXsRM)
- [Raylib game template](https://github.com/raylib-extras/game-premake)
- My implementation of the starfield animation on [Itch.io](https://paracetamol56.itch.io/starfield-animation)
- My implementation [source code](https://git.matheo-galuba.com/mga/starfield-animaiton)
```
