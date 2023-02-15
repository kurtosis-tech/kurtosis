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
      colorMode: {
        defaultMode: 'dark',
        disableSwitch: true,
        respectPrefersColorScheme: false,
      },
      navbar: {
        logo: {
          alt: 'Kurtosis',
          src: 'img/brand/kurtosis-logo-white-text.png',
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
            to: '/sdk',
            position: 'left',
            label: 'SDK',
            activeBasePath: '/sdk'
          },
          {
            href: 'https://github.com/kurtosis-tech/docs/issues/new',
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
                label: 'SDK',
                to: '/sdk',
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
      },
    }),
};

module.exports = config;
