import React from "react";
import clsx from "clsx";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";

const link = "text-refine-react-light-link dark:text-refine-react-dark-link hover:underline";

const MAIN = [
    {
        name: "OpenPanel",
        who: "For website owners",
        lead: "Using the panel: domains, websites, email, databases, files and backups.",
        href: "/docs/panel/intro/",
        screenshot: "/img/openpanel-screenshots/dashboard/dashboard-window.png",
        accent: "from-[#0FBDBD] to-[#26D97F]",
        links: [
            ["Domains", "/docs/panel/domains/"],
            ["WordPress", "/docs/panel/applications/wordpress/"],
            ["Email accounts", "/docs/panel/emails/"],
            ["MySQL databases", "/docs/panel/mysql/databases/"],
            ["Backups", "/docs/panel/backups/"],
        ],
    },
    {
        name: "OpenAdmin",
        who: "For web hosts",
        lead: "Running the server: install, users and plans, security, settings and multi-server.",
        href: "/docs/admin/intro/",
        screenshot: "/img/openadmin-screenshots/001_dashboard-window.png",
        accent: "from-[#FF9933] to-[#FF4C4D]",
        links: [
            ["Install & post-install steps", "/docs/admin/intro/"],
            ["Users", "/docs/admin/accounts/users/"],
            ["Hosting plans", "/docs/admin/plans/hosting_plans/"],
            ["Firewall", "/docs/admin/security/firewall/"],
            ["Multi-server", "/docs/admin/system/multi-server/"],
        ],
    },
    {
        name: "OpenCLI",
        who: "For the terminal",
        lead: "Every opencli command with its flags and example output, plus hooks for your own scripts.",
        href: "/docs/opencli/intro/",
        accent: "from-[#6366F1] to-[#A855F7]",
        links: [
            ["All commands", "/docs/opencli/intro/"],
            ["Users", "/docs/opencli/user/"],
            ["Domains", "/docs/opencli/domains/"],
            ["Hooks", "/docs/opencli/hooks/"],
            ["Updates", "/docs/opencli/update/"],
        ],
    },
];

const MORE = [
    ["How-to guides", "Step-by-step articles for common tasks and setups.", "/docs/articles/intro/"],
    ["OpenAdmin API", "REST API reference for integrations and automation.", "/docs/admin/api/"],
    ["Changelog", "What changed in every OpenPanel release.", "/docs/changelog/2.0.28/"],
    ["Install command", "Build the install command with your options.", "/install/"],
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

                <div className="not-prose flex flex-col gap-12 pt-8 lg:pt-16 pb-32 max-w-[928px] landing-lg:max-w-[1232px] w-full mx-auto px-4">
                    <section className="flex flex-col gap-4 max-w-[760px]">
                        <h1 className="text-3xl sm:text-5xl font-bold text-gray-900 dark:text-gray-0 m-0">Documentation</h1>
                        <p className="text-base sm:text-lg text-gray-600 dark:text-gray-400 m-0">
                            Pick the part you work with. New to OpenPanel? Start with the{" "}
                            <a href="/docs/admin/intro/" className={link}>
                                install guide
                            </a>
                            .
                        </p>
                    </section>

                    <section className="grid gap-6 lg:grid-cols-3">
                        {MAIN.map((m) => (
                            <div
                                key={m.name}
                                className="rounded-2xl border border-gray-300 dark:border-gray-700 overflow-hidden flex flex-col"
                            >
                                <a
                                    href={m.href}
                                    className="block bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 aspect-[16/9] overflow-hidden no-underline"
                                >
                                    {m.screenshot ? (
                                        <>
                                            <img src={m.screenshot} alt={`${m.name} dashboard`} className="w-full h-full object-cover object-left-top dark:hidden" />
                                            <img
                                                src={m.screenshot.replace(/\.png$/, "_dark.png")}
                                                alt={`${m.name} dashboard`}
                                                className="w-full h-full object-cover object-left-top hidden dark:block"
                                            />
                                        </>
                                    ) : (
                                        <pre className="m-0 h-full p-5 text-xs leading-relaxed bg-gray-900 !text-gray-100 overflow-hidden rounded-none whitespace-pre-wrap break-all">
                                            <span className="text-emerald-400"># </span>opencli user-add stefan ••••• stefan@example.com "Standard plan"
                                            {"\n"}[✔] Successfully added user stefan with password: •••••
                                            {"\n"}
                                            <span className="text-emerald-400"># </span>opencli domains-add example.com stefan
                                            {"\n"}Domain example.com added successfully
                                            {"\n"}
                                            <span className="text-emerald-400"># </span>opencli update --check
                                            {"\n"}
                                            {'{"status": "Up to date", "installed_version": "2.0.28", "latest_version": "2.0.28"}'}
                                        </pre>
                                    )}
                                </a>
                                <div className="p-5 sm:p-6 flex flex-col gap-4 flex-1">
                                    <div className="flex flex-col gap-1">
                                        <span className="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">{m.who}</span>
                                        <h2 className="text-2xl font-bold m-0">
                                            <a href={m.href} className="no-underline">
                                                <span className={clsx("text-transparent bg-clip-text bg-gradient-to-r", m.accent)}>{m.name}</span>
                                            </a>
                                        </h2>
                                        <p className="text-sm text-gray-600 dark:text-gray-400 m-0">{m.lead}</p>
                                    </div>
                                    <ul className="m-0 p-0 list-none flex flex-col gap-1.5">
                                        {m.links.map(([label, href]) => (
                                            <li key={href + label}>
                                                <a href={href} className={clsx("text-sm", link)}>
                                                    {label} →
                                                </a>
                                            </li>
                                        ))}
                                    </ul>
                                    <a
                                        href={m.href}
                                        className="mt-auto self-start rounded-full px-5 py-2.5 text-sm font-semibold bg-refine-blue dark:bg-refine-cyan-alt text-gray-0 dark:text-gray-900 hover:brightness-110 no-underline"
                                    >
                                        Open {m.name} docs
                                    </a>
                                </div>
                            </div>
                        ))}
                    </section>

                    <section className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
                        {MORE.map(([title, desc, href]) => (
                            <a
                                key={href}
                                href={href}
                                className="rounded-xl border border-gray-200 dark:border-gray-700 p-4 flex flex-col gap-1 no-underline hover:border-gray-400 dark:hover:border-gray-500"
                            >
                                <span className="text-sm font-semibold text-gray-900 dark:text-gray-100">{title} →</span>
                                <span className="text-sm text-gray-600 dark:text-gray-400">{desc}</span>
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
