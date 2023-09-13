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
            label: 'Starlark API Reference',
            collapsed: true,
            link: {type: 'doc', id: 'starlark-reference/index'},
            items: [
                {type: 'autogenerated', dirName: 'starlark-reference'}
            ]
        },
        'runtime-api-reference',
        'code-examples',
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
            label: 'Explanations',
            collapsed: true,
            items: [
                {type: 'autogenerated', dirName: 'explanations'}
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
            label: 'Concepts Reference',
            collapsed: true,
            items: [
                {type: 'autogenerated', dirName: 'concepts-reference'}
            ]
        },
        'faq',
        'best-practices',
        'roadmap',
        'changelog',
    ],
};

module.exports = sidebars;
