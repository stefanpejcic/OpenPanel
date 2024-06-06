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
        slug: "openpanel-community",
        title: "OpenPanel",
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
        slug: "react-admin-panel-ant-design",
        title: "B2B Internal tool with Ant Design",
        images: [
            "https://refine.ams3.cdn.digitaloceanspaces.com/templates/detail-finefoods-ant-design.jpg",
        ],
        runOnYourLocalPath: "finefoods-antd",
        liveDemo: "https://example.admin.refine.dev",
        github: "https://github.com/refinedev/refine/tree/master/examples/finefoods-antd",
        reactPlatform: "Vite",
        uiFramework: "Ant Design",
        dataProvider: "Rest API",
        authProvider: "Custom",
        description: `
This example of a B2B React admin panel, built with Refine, provides a comprehensive solution for the needs of enterprise-level internal tools. It features a full range of functionalities typical in products used by enterprise companies.

The admin panel connects to a REST API using a Simple REST data provider. Its user interface is developed with Ant Design, which Refine offers built-in UI framework support.

We built this template to showcase the efficiency and ease of using the Refine framework in development. It is a valuable resource, offering insights into Refine's best practices and modern development techniques. As it's ready for production, you can use this template as a foundation to build your own React admin panel or implement it as it is.

The template is open-source, so you can freely use or examine it to understand how Refine works.


### Key Features:

- **Dashboard**: Get an overview of food ordering activities, track product performance, and view insightful charts.
- **Orders**: Manage, track, and filter all customer orders.
- **Users**: Administer customer and courier accounts and data.
- **Stores**: View and manage the list of participating stores.
- **Categories**: Categorize and organize menu items and store types.
- **Couriers**: Monitor and manage courier activity and interactions.
- **Reviews**: Handle customer feedback, review ratings, and respond to comments.

This admin panel template can be used in for various app requirements like B2B applications, internal tools, admin panel, dashboard and all CRUD applications, providing a comprehensive platform for managing order interactions, restaurant management, and sales processes.`,
    },
];
