/**
 * Copyright (c) 2017-present, Facebook, Inc.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */

require("dotenv").config();

const redirectJson = require("./redirects.json");
const tutorialData = require("./tutorial-units");

/** @type {import('@docusaurus/types/src/index').DocusaurusConfig} */
const siteConfig = {
    title: "OpenPanel",
    tagline: 'Unparalleled support. Effortless website hosting. Continuous feature development.',
    url: "https://openpanel.com",
    baseUrl: "/",
    onBrokenLinks: 'ignore',
    projectName: "openpanel",
    organizationName: "stefanpejcic",
    trailingSlash: true,
    favicon: "img/favicon.svg",
    scripts: [
      'https://platform.twitter.com/widgets.js',
    ],
    headTags: [
        {
            tagName: "script",
            attributes: {
                async: "true",
                src: "https://www.googletagmanager.com/gtag/js?id=AW-18287005046",
            },
        },
        {
            tagName: "script",
            attributes: {},
            innerHTML: `
                window.dataLayer = window.dataLayer || [];
                function gtag(){dataLayer.push(arguments);}
                gtag('js', new Date());
                gtag('config', 'AW-18287005046');
            `,
        },
    ],
    presets: [
        [
            "@docusaurus/preset-classic",
            {
                docs: Boolean(process.env.DISABLE_DOCS)
                    ? false
                    : {
                          path: "./docs",
                          sidebarPath: require.resolve("./sidebars.js"),
                          editUrl:
                              "https://github.com/stefanpejcic/openpanel/tree/master/website",
                          showLastUpdateAuthor: true,
                          showLastUpdateTime: true,
                          disableVersioning:
                              process.env.DISABLE_VERSIONING === "true",
                          versions: {
                              current: {
                                  label: "1.7.65",
                              },
                          },
                          lastVersion: "current",
                          admonitions: {
                              tag: ":::",
                              keywords: [
                                  "additional",
                                  "note",
                                  "tip",
                                  "info-tip",
                                  "info",
                                  "caution",
                                  "danger",
                                  "sourcecode",
                                  "create-example",
                                  "simple",
                              ],
                          },
                          exclude: ["**/**/_*.md"],
                          // Adding the sidebarItemsGenerator with reverse logic
                          async sidebarItemsGenerator({defaultSidebarItemsGenerator, ...args}) {
                              const sidebarItems = await defaultSidebarItemsGenerator(args);

                              // Reverse items only for the 'Changelog' category
                              const modifiedSidebarItems = sidebarItems.map((item) => {
                                  if (item.type === 'category' && item.label === 'Changelog') {
                                      return { ...item, items: item.items.reverse() };
                                  }
                                  return item;
                              });

                              return modifiedSidebarItems;
                          },
                      },
                blog: false,
                theme: {
                    customCss: [
                        require.resolve("./src/refine-theme/css/colors.css"),
                        require.resolve("./src/refine-theme/css/fonts.css"),
                        require.resolve("./src/refine-theme/css/custom.css"),
                        require.resolve("./src/css/custom.css"),
                        require.resolve("./src/css/split-pane.css"),
                        require.resolve("./src/css/demo-page.css"),
                    ],
                },
                gtag: {
                    trackingID: ["G-5YBHWWDTL4", "AW-18287005046"],
                },
                sitemap: {
                    ignorePatterns: ["**/_*.md"],
                },
            },
        ],
    ],
    plugins: [
        [
            "@docusaurus/plugin-client-redirects",
            {
                redirects: redirectJson.redirects,
                createRedirects(existingPath) {
                    if (existingPath.includes("/api-reference/core/")) {
                        return [
                            existingPath.replace(
                                "/api-reference/core/",
                                "/api-references/",
                            ),
                        ];
                    }
                    return undefined; // Return a falsy value: no redirect created
                },
            },
        ],
        [
            "docusaurus-plugin-copy",
            {
                id: "Copy Workers",
                path: "static/workers",
                context: "workers",
                include: ["**/*.{js}"],
            },
        ],
        [
            "docusaurus-plugin-generate-llms-txt",
            {
                outputFile: "llms-full.txt",
            },
        ],
        "./plugins/llms-pages-plugin.js",
        [
            "@easyops-cn/docusaurus-search-local",
            {
                hashed: true,
                indexBlog: false,
                indexPages: false,
                ignoreFiles: [/\/changelog\//],
            },
        ],
        async function tailwindcss() {
            return {
                name: "docusaurus-tailwindcss",
                configurePostCss(postcssOptions) {
                    postcssOptions.plugins.push(require("tailwindcss"));
                    postcssOptions.plugins.push(require("autoprefixer"));
                    return postcssOptions;
                },
            };
        },
        "./plugins/docgen.js",
        "./plugins/checklist.js",
        ...(process.env.DISABLE_BLOG
            ? []
            : [
                  [
                      "./plugins/blog-plugin.js",
                      {
                          blogTitle: "Blog",
                          blogDescription:
                              "A resource for OpenPanel, front-end ecosystem, and web development",
                          routeBasePath: "/blog",
                          postsPerPage: 12,
                          blogSidebarTitle: "All posts",
                          blogSidebarCount: 0,
                          feedOptions: {
                              type: "all",
                              copyright: `Copyright © ${new Date().getFullYear()} OpenPanel.`,
                          },
                      },
                  ],
              ]),
        "./plugins/clarity.js",
        "./plugins/example-redirects.js",
        "./plugins/og-images.js",
    ],
    themeConfig: {
        prism: {
            theme: require("prism-react-renderer/themes/github"),
            darkTheme: require("prism-react-renderer/themes/vsDark"),
            magicComments: [
                // Remember to extend the default highlight class name as well!
                {
                    className: "theme-code-block-highlighted-line",
                    line: "highlight-next-line",
                    block: { start: "highlight-start", end: "highlight-end" },
                },
                {
                    className: "code-block-hidden",
                    line: "hide-next-line",
                    block: { start: "hide-start", end: "hide-end" },
                },
                {
                    className: "theme-code-block-added-line",
                    line: "added-line",
                    block: { start: "added-start", end: "added-end" },
                },
                {
                    className: "theme-code-block-removed-line",
                    line: "removed-line",
                    block: { start: "removed-start", end: "removed-end" },
                },
            ],
        },
        image: "img/openpanel_social.png",
        metadata: [
            {
                name: "keywords",
                content:
                    "openpanel, openadmin, open panel, open admin, open hosting panel, open control panel",
            },
        ],
        navbar: {
            logo: {
                alt: "OpenPanel",
                src: "img/svg/openpanel_logo.svg",
            },
            items: [
                { to: "/blog", label: "Blog", position: "left" },
                {
                    type: "docsVersionDropdown",
                    position: "right",
                    dropdownActiveClassDisabled: true,
                },
                {
                    href: "https://github.com/stefanpejcic/openpanel",
                    position: "right",
                    className: "header-icon-link header-github-link",
                },
                {
                    href: "https://discord.com/invite/7bNY8fANqF",
                    position: "right",
                    className: "header-icon-link header-discord-link",
                },
                {
                    href: "https://x.com/openpanel",
                    position: "right",
                    className: "header-icon-link header-twitter-link",
                },
            ],
        },
        footer: {
            logo: {
                alt: "OpenPanel",
                src: "/img/svg/openpanel_logo.svg",
            },
            links: [
                {
                    title: "Resources",
                    items: [
                        {
                            label: "OpenPanel Docs",
                            to: "/docs/panel/intro/",
                        },
                        {
                            label: "OpenAdmin Docs",
                            to: "/docs/admin/intro/",
                        },
                        {
                            label: "Blog",
                            to: "/blog",
                        },
                    ],
                },
                {
                    title: "Product",
                    items: [
                        {
                            label: "Features",
                            to: "/features",
                        },
                        {
                            label: "Live Preview",
                            to: "/demo",
                        },
                        {
                            label: "Roadmap",
                            to: "/roadmap",
                        },
                    ],
                },
                {
                    title: "Company",
                    items: [
                        {
                            label: "About",
                            to: "/about",
                        },
                        {
                            label: "Support",
                            to: "/support",
                        },
                    ],
                },
                {
                    title: "__LEGAL",
                    items: [
                        {
                            label: "License",
                            to: "/LICENSE",
                        },
                        {
                            label: "Privacy",
                            to: "/privacy-policy",
                        },
                        {
                            label: "info@openpanel.com",
                            to: "mailto:info@openpanel.com",
                        },
                    ],
                },
                {
                    title: "__SOCIAL",
                    items: [
                        {
                            href: "https://github.com/stefanpejcic/openpanel",
                            label: "github",
                        },
                        {
                            href: "https://discord.com/invite/7bNY8fANqF",
                            label: "discord",
                        },
                        {
                            href: "https://www.reddit.com/r/openpanelco/",
                            label: "reddit",
                        },
                        {
                            href: "https://x.com/openpanel",
                            label: "twitter",
                        },
                        {
                            href: "https://www.linkedin.com/company/openpanel/",
                            label: "linkedin",
                        },
                    ],
                },
            ],
        },
        docs: {
            sidebar: {
                autoCollapseCategories: false,
            },
        },
        colorMode: {
            defaultMode: "dark",
        },
    },
    customFields: {
        /** Footer Fields */
        footerDescription:
            '<strong style="font-weight:700;">OpenPanel</strong> is a next generation hosting panel for more secure and provacy focused hosting.',
        contactTitle: "Contact",
        contactDescription: [
            "OpenPanel, LLC",
            "131 Continental Dr, Newark, Delaware",
        ],
        contactEmail: "info@openpanel.com",
        /** ---- */
        /** Live Preview */
        LIVE_PREVIEW_URL:
            process.env.LIVE_PREVIEW_URL ?? "http://localhost:3030/preview",
        /** ---- */
        tutorial: tutorialData,
    },
    webpack: {
        jsLoader: (isServer) => ({
            loader: require.resolve("swc-loader"),
            options: {
                jsc: {
                    parser: {
                        syntax: "typescript",
                        tsx: true,
                    },
                    target: "es2017",
                },
                module: {
                    type: isServer ? "commonjs" : "es6",
                },
            },
        }),
    },
};

module.exports = siteConfig;
