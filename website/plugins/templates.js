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
        liveDemo: "https://demo.openpanel.org/openadmin/",
        tutorial: "https://my.openpanel.com/cart.php?a=add&pid=1",
        reactPlatform: "Premium Features",
        uiFramework: "∞ Users & Websites",
        dataProvider: "WHMCS, HostBill",
        authProvider: "Hands-on Support",
        description: `
OpenPanel Enterprise edition offers advanced features for user isolation and management, suitable for web hosting providers.

### Pricing Options:

- [**Monthly Plan**: 14.95€ per month](https://my.openpanel.com/cart.php?a=add&pid=1&billingcycle=monthly&skipconfig=1)
- [**Yearly Plan**: 149.50€ per year (20% savings)](https://my.openpanel.com/cart.php?a=add&pid=1&billingcycle=annually&skipconfig=1)

Monthly/yearly options come with a price-lock guarantee, ensuring that future price increases will not affect you. You will always pay the price you started with.

### Premium features:

- **White-Labeling**: Options to completely brand the panel as your own.
- **Billing Integrations**: WHMCS, FossBilling.
- **Auto-Installers**: WordPress, Mautic, nodeJS, Python..
- **Additional Pro Features**: MariaDB support, Emails, FTP..
- **Migration Tools**: cPanel account 2 OpenPanel import
- **Real-Time Updates**: Immediate bug fixes as soon as they are available.
- **On-Site Support**: TeamViewer/AnyDesk support.

`,
   },
];
