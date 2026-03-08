// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	site: 'https://dcsg.github.io',
	base: '/archway',
	integrations: [
		starlight({
			title: 'Archway',
			tagline: 'Architecture-aware service composer',
			social: [{ icon: 'github', label: 'GitHub', href: 'https://github.com/dcsg/archway' }],
			editLink: {
				baseUrl: 'https://github.com/dcsg/archway/edit/main/website/',
			},
			sidebar: [
				{
					label: 'Getting Started',
					items: [
						{ label: 'Introduction', slug: 'getting-started/introduction' },
						{ label: 'Installation', slug: 'getting-started/installation' },
						{ label: 'Quick Start', slug: 'getting-started/quickstart' },
					],
				},
				{
					label: 'Concepts',
					items: [
						{ label: 'How It Works', slug: 'concepts/how-it-works' },
						{ label: 'Architectures', slug: 'concepts/architectures' },
						{ label: 'Capabilities', slug: 'concepts/capabilities' },
						{ label: 'Bootstrap Pattern', slug: 'concepts/bootstrap' },
					],
				},
				{
					label: 'Reference',
					items: [
						{ label: 'Capabilities Matrix', slug: 'reference/capabilities-matrix' },
						{ label: 'CLI Commands', slug: 'reference/cli' },
						{ label: 'archway.yaml', slug: 'reference/archway-yaml' },
					],
				},
				{
					label: 'Guides',
					items: [
						{ label: 'Building a REST API', slug: 'guides/rest-api' },
						{ label: 'Adding Capabilities', slug: 'guides/adding-capabilities' },
					],
				},
			],
		}),
	],
});
