/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

 Create as many sidebars as you want.
 */

// @ts-check

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
    main: [
        {
            "type": 'category',
            label: 'Get Started',
            collapsed: true,
            items: [
                {type: 'autogenerated', dirName: 'get-started'}
            ]
        },
        {
            "type": "link",
            "label": "Kurtosis for Web3",
            "href": "https://web3.kurtosis.com",
        },
        {
            type: 'category',
            label: 'Guides',
            collapsed: true,
            items: [
                {type: 'autogenerated', dirName: 'guides'}
            ]
        },
        {
            type: 'category',
            label: 'SDK Examples',
            collapsed: true,
            items: [
                'sdk-examples/go-sdk-example',
                'sdk-examples/ts-sdk-example'
            ]
        },
        {
            type: 'category',
            label: 'API Reference',
            collapsed: true,
            items: [
                'api-reference/engine-apic-reference',
                {
                    type: 'category',
                    label: 'Starlark Reference',
                    collapsed: true,
                    link: {type: 'doc', id: 'api-reference/starlark-reference/index'},
                    items: [
                        {type: 'autogenerated', dirName: 'api-reference/starlark-reference'}
                    ]
                },

            ]
        },
        {
            type: 'category',
            label: 'CLI Reference',
            collapsed: true,
            link: {type: 'doc', id: 'cli-reference/index'},
            items: [
                {type: 'autogenerated', dirName: 'cli-reference'}
            ]
        },
        {
            type: 'category',
            label: 'Advanced Concepts',
            collapsed: true,
            items: [
                {type: 'autogenerated', dirName: 'advanced-concepts'}
            ]
        },
        'code-examples',
        'faq',
        'best-practices',
        'roadmap',
        'changelog',
    ],
};

module.exports = sidebars;
