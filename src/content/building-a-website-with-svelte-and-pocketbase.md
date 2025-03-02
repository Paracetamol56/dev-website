---
title: Building a Website with Svelte and PocketBase
description: How I built the ICAY website using Svelte, PocketBase. A look into my tech choices.
tags:
  - web-dev
  - devlog
release: 2025-03-02T12:00:00.000Z
listed: true
---

<script>
  import CodeBlock from '$lib/components/CodeBlock.svelte';
</script>

## Introduction

In 2024, I built the first version of the [ICAY](https://icay.saf-astronomie.fr/en/) (International Congress of the Astronomical Youth) website for that year's edition of the congress. Having previously experimented with Svelte, I was eager to use it again for this project. Its lightweight nature and efficient DOM management made it a perfect choice. For the 2025 edition, I expanded the stack by introducing [PocketBase](https://pocketbase.io/) to manage online registration. Overall, I enjoyed working with this stack, though there were some trade-offs.

## The Tech Stack

### Svelte: Simplicity and Performance

<img class="content-ignore float-left w-20 mr-4 mb-2" src="/img/building-a-website-with-svelte-and-pocketbase/svelte-logo.png" alt="Svelte logo" />

I chose [Svelte](https://svelte.dev/) 4 as the frontend framework because it provides a highly efficient way to manage the DOM without the overhead of a virtual DOM. The reactivity model feels intuitive, and its compilation-based approach results in lightweight and performant applications. When I first tried Svelte, I was impressed by how declarative and straightforward it made state management.

<div class="clear-left">

For this project, I used [SvelteKit](https://svelte.dev/docs/kit/introduction) with [static site generation](https://svelte.dev/docs/kit/adapter-static). This approach is particularly interesting because it allows for fast and efficient deployment while keeping the site lightweight. The website is hosted on a shared server with multiple sites and served using Apache. With the correct Apache configuration in the [.htaccess](https://opengist.matheo-galuba.com/mga/02681a9a5fe94ac1b6d863b805cc4e50) file, I even managed to make some routes dynamic and set up custom redirects, ensuring a seamless navigation experience for users.

</div>

### PocketBase: A Lightweight Backend Solution

For the 2025 edition, I needed a backend to handle online registrations, and [PocketBase](https://pocketbase.io/) turned out to be an excellent choice. It provides:

- A simple setup
- An integrated admin dashboard for managing data
- Built-in authentication and database management
- A lightweight yet powerful solution

<img class="clear-left max-w-7xl" src="/img/building-a-website-with-svelte-and-pocketbase/pocketbase-banner.png" alt="PocketBase backoffice" />

PocketBase is built with Go, which makes it verry fast and efficient. Since it runs as a single binary, deployment is straightforward, and its SQLite-based storage keeps things lightweight while still providing enough flexibility for ICAY's registration needs.

PocketBase eliminated the need for a separate backend framework, making development faster and more efficient. The built-in admin panel also allowed for quick content management without requiring custom backoffice tooling.

### I18N: Multilingual Support

Since the ICAY website caters to an international audience, I18N was crucial for supporting multiple languages. I implemented translations to ensure a seamless user experience in different locales.

### Markdown & File-Based CMS

The site’s content is managed using a file-based CMS with Markdown files, allowing for easy content editing and version control. While this approach worked well for basic content management, I found it lacking compared to Vue’s Nuxt Content, which offers built-in search and better indexing capabilities.

#### How It Works

- Extracts Route Parameters – lang and slug.
- Dynamically imports the Markdown file – It tries to load `/src/content/{lang}/{slug}.md`.
- Returns Content & Metadata:
  - page.default contains the Markdown content.
  - page.metadata holds frontmatter data.

<CodeBlock>

```ts
import { error } from '@sveltejs/kit';

export async function load({ params }) {
 try {
  const page = await import(`../../../content/${params.lang}/${params.slug}.md`);
  return {
   content: page.default, // The actual Markdown content
   meta: page.metadata // Any frontmatter metadata
  };
 } catch (e) {
  console.error(e);
  throw error(404, `Page not found`);
 }
}
```

</CodeBlock>

#### Extending Markdown

Beyond simple Markdown files, I extended Markdown with Svelte components. This allows embedding dynamic elements directly within content files. For example, the contact page includes a Svelte component for the contact form:

<CodeBlock>

```md
---
title: Contact
description: Contact the International Congress of the Astronomical Youth
---

<script>
  import ContactForm from '$lib/components/contact/ContactForm.svelte';
</script>

<p class="text-center max-w-xl mx-auto">
  To contact us, you can send an e-mail to the address
  <br>
  <a href="mailto:icay@saf-astronomie.fr">icay@saf-astronomie.fr</a>
  <br>
  or use the contact form below:
</p>

<br>
<br>

<ContactForm />
```

</CodeBlock>

This hybrid approach allows dynamic and interactive elements within Markdown content while keeping the benefits of a simple file-based CMS.

## Form Submission System

One of the features I’m particularly proud of is the form submission system. Instead of handling form states in each component separately, I built an abstract class called `FormState` and used Svelte stores to share data between components. This approach made:

- Form submission cleaner and reusable
- Error handling more structured
- Component communication seamless

<br>

By leveraging Svelte stores, form states remain centralized and reactive, improving the user experience and maintainability of the codebase.

Here’s the `FormState` abstract class:

<CodeBlock>

```ts
export default abstract class FormState<T> {
 data: T;
 loading: boolean;
 error: Error | null;
 success: boolean;

 constructor(data: T) {
  this.data = data;
  this.loading = false;
  this.error = null;
  this.success = false;
 }

 reset(): void {
  this.loading = false;
  this.error = null;
  this.success = false;
 }

 async submit(): Promise<void> {
  this.loading = true;
  this.error = null;
  try {
   await this.post();
   this.success = true;
  } catch (error: any) {
   console.error('Error while submitting form', error);
   this.success = false;
   if (error instanceof Error) {
    this.error = error;
   }
   this.handleError();
  }
  this.loading = false;
 }

 abstract post(): Promise<void>;
 handleError(): Promise<void> | void {}
}
```

</CodeBlock>

This class ensures that all forms share a consistent structure. Each form component uses its own child of FormState, defining its own post() method to handle submission logic. By storing the form state in a Svelte store, multiple components, such as modal dialogs or summary views, can reactively update as the form state changes.

## Conclusion

Working with Svelte and PocketBase for the ICAY 2025 edition was a rewarding experience. The combination of Svelte’s simplicity and reactivity with PocketBase’s lightweight yet powerful backend allowed me to build a solid and efficient platform for online registrations.

That said, the website is highly custom and self-made, which means it’s not easily manageable by the ICAY admins. Any modifications require my expertise, making it less accessible for direct content updates by non-technical users. Despite this, I’m really proud of what I was able to accomplish. Seeing everything come together smoothly, from multilingual support to dynamic content handling, reinforced my appreciation for this tech stack. I’m excited to refine it further and see how it evolves for future editions of ICAY!
