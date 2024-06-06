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
        runOnYourLocalPath: "app-crm",
        liveDemo: "https://demo.openpanel.co/openpanel/",
        github: "https://github.com/refinedev/refine/tree/master/examples/app-crm",
        reactPlatform: "Vite",
        uiFramework: "Ant Design",
        dataProvider: "Nestjs-query",
        authProvider: "Custom",
        description: `
This  CRM app example, built with Refine, demonstrates a complete solution for enterprise-level CRM internal tool needs. It has a wide range of functionalities for real-use cases, which are extensively utilized by enterprise companies.

The app connected to GraphQL API through Refine's Nestjs-query data provider, and its user interface is developed using Ant Design, which Refine offers built-in UI framework support. 

We built this template to demonstrate how the Refine framework simplifies and speeds up development. It is a valuable resource, offering insights into Refine's best practices and modern development techniques.

The source code of the CRM app is also open-source; feel free to use or inspect it to discover how Refine works. Being production-ready, you can either build your own CRM internal tool using it as a code base or directly implement it as is.


### Key Features:


- **Dashboard**: Overview of CRM activities, key metrics, and customer interactions.
- **Calendar**: Manage appointments and events.
- **Scrumboard-Project Kanban**: Streamline project management and task tracking.
- **Sales Pipeline**: Visualize sales stages and track lead conversions.
- **Companies**: Organize business contacts with detailed profiles.
- **Contacts**: Efficient management of individual customer interactions.
- **Quotes**: Create, send, and manage customer quotes.
- **Administration**: Customize CRM settings, user roles, and permissions.


This CRM app template can be used in for various app requirements like B2B applications, internal tools, admin panel, dashboard and all CRUD applications, providing a comprehensive platform for: 

- Human Resource Management (HRM) Tools
- IT Service Management (ITSM) Tools
- Network Monitoring Tools
- Risk Management Tools
- Customer Support Tools
- Financial Planning Systems
- Customer Analytics Tools
- Inventory Management Systems
- Supply Chain Management Tools
- Retail Management Systems
- Business Intelligence Tools
- Electronic Health Record (EHR) Systems
- Patient Management Systems
- Health Information Exchange (HIE) Systems
- Pharmacy Management Systems
`,
    },
    {
        slug: "openpanel-premium-control-panel",
        title: "Enterprise edition",
        images: [
            "https://refine.ams3.cdn.digitaloceanspaces.com/templates/detail-finefoods-ant-design.jpg",
        ],
        runOnYourLocalPath: "--key=xxxxxxxx",
        liveDemo: "https://demo.openpanel.co/admin/",
        github: "https://github.com/refinedev/refine/tree/master/examples/finefoods-antd",
        reactPlatform: "Vite",
        uiFramework: "Ant Design",
        dataProvider: "Rest API",
        authProvider: "Custom",
        description: `
OpenPanel Enterprise edition offers advanced features for user isolation and management, suitable for web hosting providers.

### Pricing Options:

- **Monthly Plan**: $19.95 per month with a price-lock guarantee.
- **Yearly Plan**: $192 per year (20% savings).
- **Lifetime Plan**: $248.

The Lifetime Plan will be available only this year and is intended to help us accelerate the development of features for the Pro version.

### Premium features:

- **White-Labeling**: Options to completely brand the panel as your own.
- **Billing Integrations**: WHMCS, Blesta, HostBill, FossBilling.
- **Auto-Installers**: 50+ CMS and tools.
- **Additional Pro Features**: OpenLitespeed, Varnish, Emails, FTP, Clusters, Cloudflare DNS.
- **Migration Tools**: cPanel2OpenPanel and WordPress2OpenPanel.
- **Backup Storage**: 100GB of backup storage.
- **Real-Time Updates**: Immediate bug fixes as soon as they are available.
- **On-Site Support**: TeamViewer/AnyDesk support.

For hosting providers, we also offer white-label tutorials for your blog/knowledgebase and support ticket options. When clients have technical questions or issues related to OpenPanel, you can direct-message our developers for answers or resolutions with a 30-minute SLA.

Please note that not all features for the Pro version will be immediately available; it will take several months for all planned features to be implemented.
   },
];
