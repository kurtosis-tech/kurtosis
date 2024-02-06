// @ts-check

const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/vsDark');

/** @type {import('@docusaurus/types').Config} */
const config = {
	title: 'Kurtosis Docs',
	tagline: 'Next-gen developer experience for building, testing, and running distributed systems.',
	url: 'https://docs.kurtosis.com',
	baseUrl: '/',
	onBrokenLinks: 'throw',
	onBrokenMarkdownLinks: 'throw',
	favicon: 'img/favicon.ico',

	// GitHub pages deployment config.
	organizationName: 'kurtosis-tech',
	projectName: 'docs',

	i18n: {
		defaultLocale: 'en',
		locales: ['en'],
	},

	presets: [
		[
			'classic',
			/** @type {import('@docusaurus/preset-classic').Options} */
			({
				docs: {
					sidebarPath: require.resolve('./sidebars.js'),
					routeBasePath: '/',
					admonitions: {}, // Add this line to enable admonitions

					// TODO TODO Run Remark plugins through Docusaurus itself (right now we're running it via yarn and package.json)!! See https://docusaurus.io/docs/markdown-features/plugins#installing-plugins

				},
				theme: {
					customCss: require.resolve('./src/css/custom.css'),
				},
				gtag: {
					trackingID: 'G-9D2YD4C5FV',
				},
			}),
		],
	],

	themeConfig:
	/** @type {import('@docusaurus/preset-classic').ThemeConfig} */
		({
			announcementBar: {
				id: 'support_us',
				content:
					'<a target="_blank" rel="noopener noreferrer" href="https://github.com/kurtosis-tech/kurtosis">Support Kurtosis with a star on our Github repo!</a>',
				backgroundColor: '#1b1b1d',
				textColor: '#909294',
				isCloseable: false,
			},
			colorMode: {
				defaultMode: 'dark',
				disableSwitch: true,
				respectPrefersColorScheme: false,
			},
			navbar: {
				logo: {
					alt: 'Kurtosis',
					src: 'img/brand/kurtosis-logo-white-text.png',
					href: 'https://docs.kurtosis.com',
					target: '_self'
				},
				items: [
					{
						to: '/quickstart',
						position: 'left',
						label: 'Quickstart',
						activeBasePath: '/quickstart'
					},
					{
						to: '/cli',
						position: 'left',
						label: 'CLI',
						activeBasePath: '/cli'
					},
					{
						to: '/starlark-reference',
						position: 'left',
						label: 'Starlark',
						activeBasePath: '/sdk'
					},
					{
						href: 'https://www.kurtosis.com/release-notes',
						position: 'left',
						label: 'Release Notes',
					},
					{
						href: 'https://github.com/kurtosis-tech/kurtosis/issues/new?assignees=leeederek&labels=docs&template=docs-issue.yml',
						position: 'right',
						label: 'Report Docs Issue',
					},
				],
			},
			footer: {
				style: 'dark',
				links: [
					{
						title: 'Docs',
						items: [
							{
								label: 'Quickstart',
								to: '/quickstart',
							},
							{
								label: 'CLI',
								to: '/cli',
							},
							{
								label: 'Starlark',
								to: '/starlark-reference',
							},
							{
								label: 'Kurtosis for Web3',
								href: 'https://web3.kurtosis.com',
							},
						],
					},
					{
						title: 'Community',
						items: [
							{
								label: 'Discord',
								href: 'https://discord.gg/HUapYX9RvV',
							},
							{
								label: 'Twitter',
								href: 'https://twitter.com/KurtosisTech',
							},
							{
								label: 'GitHub',
								href: 'https://github.com/kurtosis-tech',
							},
						],
					},
					{
						title: 'Company',
						items: [
							{
								label: `Careers - We're Hiring`,
								href: 'https://www.kurtosis.com/careers',
							},
							{
								label: 'About Us',
								href: 'https://www.kurtosis.com/company',
							},
							{
								label: 'Blog',
								href: 'https://www.kurtosis.com/blog',
							},
						],
					},
				],
				copyright: `Copyright © ${new Date().getFullYear()} Kurtosis Technologies`,
			},
			prism: {
				theme: lightCodeTheme,
				darkTheme: darkCodeTheme,
				additionalLanguages: ['bash', 'shell-session'],
			},
			algolia: {
				appId: 'NTSX40VZB8',

				// Public API key, safe to commit
				apiKey: '4269c726c2fea4e6cddfeb9a21cd3d4e',

				indexName: 'kurtosis',

				contextualSearch: true,

				searchParameters: {},

				searchPagePath: 'search',
			},
		}),
	scripts: [
		{
			src: '/js/load-fullstory.js',
			async: false,
		}
	]
};

module.exports = config;
