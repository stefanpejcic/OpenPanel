import React from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";

const link = "text-refine-react-light-link dark:text-refine-react-dark-link hover:underline";

const MAIN = [
    {
        name: "OpenPanel",
        lead: "The user panel. For people who host websites on an OpenPanel server.",
        href: "/docs/panel/intro/",
        links: [
            ["Domains and DNS", "/docs/panel/domains/"],
            ["WordPress", "/docs/panel/applications/wordpress/"],
            ["Email accounts", "/docs/panel/emails/"],
            ["MySQL databases", "/docs/panel/mysql/databases/"],
            ["PHP versions", "/docs/panel/php/domains/"],
            ["File manager", "/docs/panel/files/"],
            ["Cron jobs", "/docs/panel/cronjobs/"],
            ["Backups", "/docs/panel/backups/"],
        ],
    },
    {
        name: "OpenAdmin",
        lead: "The admin panel. For whoever runs the server and creates the accounts.",
        href: "/docs/admin/intro/",
        links: [
            ["Install and first steps", "/docs/admin/intro/"],
            ["Users", "/docs/admin/accounts/users/"],
            ["Hosting plans", "/docs/admin/plans/hosting_plans/"],
            ["Import from cPanel", "/docs/admin/backups/cpanel/"],
            ["Firewall", "/docs/admin/security/firewall/"],
            ["WAF", "/docs/admin/security/waf/"],
            ["Notifications", "/docs/admin/notifications/"],
            ["Multi-server", "/docs/admin/system/multi-server/"],
        ],
    },
    {
        name: "OpenCLI",
        lead: "The opencli command. Every command with its options and example output.",
        href: "/docs/opencli/intro/",
        links: [
            ["All commands", "/docs/opencli/intro/"],
            ["Users", "/docs/opencli/user/"],
            ["Domains", "/docs/opencli/domains/"],
            ["PHP", "/docs/opencli/php/"],
            ["Backups", "/docs/opencli/backup/"],
            ["WAF", "/docs/opencli/waf/"],
            ["Hooks", "/docs/opencli/hooks/"],
            ["Updates", "/docs/opencli/update/"],
        ],
    },
];

const MORE = [
    ["How-to guides", "/docs/articles/intro/"],
    ["OpenAdmin API", "/docs/admin/api/"],
    ["Changelog", "/docs/changelog/2.0.28/"],
    ["Install command generator", "/install/"],
];

const Docs: React.FC = () => {
    return (
        <CommonLayout
            description="Browse the OpenPanel and OpenAdmin documentation: installation, API reference, terminal commands, and how-to guides for running OpenPanel hosting infrastructure."
            image="/img/og/docs.png"
        >
            <Head>
                <title>Documentation | OpenPanel</title>
                <meta property="og:title" content="Documentation | OpenPanel" />
                <meta property="og:image:width" content="1200" />
                <meta property="og:image:height" content="630" />
                <meta property="og:image:type" content="image/png" />
                <html data-page="docs" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="not-prose flex flex-col gap-12 pt-8 lg:pt-16 pb-24 max-w-[928px] landing-lg:max-w-[1232px] w-full mx-auto px-4">
                    <section className="flex flex-col gap-4 max-w-[760px]">
                        <h1 className="text-3xl sm:text-5xl font-bold text-gray-900 dark:text-gray-0 m-0">Documentation</h1>
                        <p className="text-base sm:text-lg text-gray-600 dark:text-gray-400 m-0">
                            Three parts, one for each side of OpenPanel. Installing for the first time? Start with the{" "}
                            <a href="/docs/admin/intro/" className={link}>
                                install guide
                            </a>
                            .
                        </p>
                    </section>

                    <section className="grid gap-10 lg:gap-12 lg:grid-cols-3 border-t border-gray-200 dark:border-gray-700 pt-8">
                        {MAIN.map((m) => (
                            <div key={m.name} className="flex flex-col gap-3">
                                <h2 className="text-xl font-bold m-0">
                                    <a href={m.href} className="no-underline text-gray-900 dark:text-gray-0 hover:underline">
                                        {m.name}
                                    </a>
                                </h2>
                                <p className="text-sm text-gray-600 dark:text-gray-400 m-0">{m.lead}</p>
                                <ul className="m-0 pl-0 list-none flex flex-col gap-2 text-sm">
                                    {m.links.map(([label, href]) => (
                                        <li key={href + label}>
                                            <a href={href} className={link}>
                                                {label}
                                            </a>
                                        </li>
                                    ))}
                                </ul>
                            </div>
                        ))}
                    </section>

                    <section className="border-t border-gray-200 dark:border-gray-700 pt-6 flex flex-wrap gap-x-8 gap-y-2 text-sm">
                        <span className="text-gray-600 dark:text-gray-400">Also:</span>
                        {MORE.map(([title, href]) => (
                            <a key={href} href={href} className={link}>
                                {title}
                            </a>
                        ))}
                    </section>

                    <p className="text-sm text-gray-600 dark:text-gray-400 m-0">
                        Can't find something? Ask in the{" "}
                        <a href="https://github.com/stefanpejcic/OpenPanel/discussions" className={link}>
                            GitHub Discussions
                        </a>{" "}
                        or try the{" "}
                        <a href="/demo/" className={link}>
                            live demo
                        </a>
                        .
                    </p>
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Docs;
