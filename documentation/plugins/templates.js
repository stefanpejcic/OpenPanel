"use strict";
Object.defineProperty(exports, "__esModule", { value: true });

async function RefineTemplates() {
    return {
        name: "docusaurus-plugin-refine-templates",
        contentLoaded: async (args) => {
            const { content, actions } = args;
            const { addRoute, createData } = actions;

            await Promise.all(
                content.map(async (data) => {
                    const json = await createData(
                        `templates-${data.slug}.json`,
                        JSON.stringify(data, null, 2),
                    );

                    addRoute({
                        path: `/product/${data.slug}`,
                        component:
                            "@site/src/components/templates-detail-page/index",
                        exact: true,
                        modules: {
                            content: json,
                        },
                    });
                }),
            );
        },
        loadContent: async () => {
            return templates;
        },
    };
}
exports.default = RefineTemplates;

const templates = [
    {
        slug: "openpanel-free-control-panel",
        title: "Community edition",
        images: [
            "/img/panel_cropped.png",
        ],
        runOnYourLocalPath: " ",
        github: "https://github.com/stefanpejcic/openpanel",
        reactPlatform: "Regular Features",
        uiFramework: "3 Users, 10 Websites",
        dataProvider: "No Billing Integrations",
        authProvider: "Discord & Forums",
        description: `
OpenPanel Community edition is a free hosting control panel for Debian and Ubuntu OS, suitable for VPS and private use.

### Key Features:


- **Resource Limiting**: Limit CPU, Memory, disk usage and port speed.
- **Multiple WebServers**: Each user has a private instance of Nginx or Apache webserver.
- **Select PHP versions**: User can install PHP versions, set them on domain-level, edit php.ini.
- **Advanced Configuration**: Edit MySQL or Nginx/Apache Configuration, Restart services, and more.
- **File Management**: Responsive File Manager with background uploads processing and bulk actions.
- **Autoinstallers**: WP Manager, NodeJS and Python Applications.
- **Usage Overview**: Detailed resource usage reports and beautiful domain access logs.
- **Detailed Logging**: Every user login and interation is logged.

The free version of OpenPanel includes more features than some paid alternatives, and we're proud of that.

We promise that OpenPanel will always have a free version. ✌️💯
`,
    },
    {
        slug: "openpanel-premium-control-panel",
        title: "Enterprise edition",
        images: [
            "/img/openpanel_admin.gif",
        ],
        //runOnYourLocalPath: "",
        liveDemo: "https://demo.openpanel.co/openadmin/",
        tutorial: "https://my.openpanel.co/products",
        reactPlatform: "Premium Features",
        uiFramework: "∞ Users & Websites",
        dataProvider: "WHMCS, HostBill",
        authProvider: "Hands-on Support",
        description: `
OpenPanel Enterprise edition offers advanced features for user isolation and management, suitable for web hosting providers.

### Pricing Options:

- [**Monthly Plan**: 14.95€ per month with a price-lock guarantee](https://my.openpanel.co/cart.php?a=add&pid=1&billingcycle=monthly&skipconfig=1)
- [**Yearly Plan**: 149.5€ per year (20% savings)](https://my.openpanel.co/cart.php?a=add&pid=1&billingcycle=annually&skipconfig=1)


### Premium features:

- **White-Labeling**: Options to completely brand the panel as your own.
- **Billing Integrations**: WHMCS, Blesta, HostBill, FossBilling.
- **Auto-Installers**: 50+ CMS and tools.
- **Additional Pro Features**: OpenLitespeed, Varnish, Emails, FTP, Clusters, Cloudflare DNS.
- **Migration Tools**: cPanel2OpenPanel and WordPress2OpenPanel.
- **Backup Storage**: 100GB of backup storage.
- **Real-Time Updates**: Immediate bug fixes as soon as they are available.
- **On-Site Support**: TeamViewer/AnyDesk support.

📢 For hosting providers, we also offer white-label tutorials for your blog/knowledgebase and support ticket options. When clients have technical questions or issues related to OpenPanel, you can direct-message our developers for answers or resolutions with a 30-minute SLA.

Please note that not all features for the Pro version will be immediately available; it will take several months for all planned features to be implemented.
`,
   },
];
