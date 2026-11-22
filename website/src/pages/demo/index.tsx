import React from "react";
import clsx from "clsx";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { DemoInstallCta } from "@site/src/refine-theme/demo-install-cta";

const DEMOS = [
    {
        name: "OpenPanel",
        who: "For website owners",
        lead: "The panel your customers use to run their websites.",
        url: "https://demo.openpanel.com:2083/login",
        port: "2083",
        screenshot: "/img/openpanel-screenshots/dashboard/dashboard-window.png",
        alt: "OpenPanel dashboard",
        tryThis: [
            "Add a domain and open the DNS zone editor",
            "Install WordPress or another app in one click",
            "Create email accounts and databases",
            "Browse files, backups and resource usage",
        ],
        docs: "/docs/panel/intro/",
        accent: "from-[#0FBDBD] to-[#26D97F]",
    },
    {
        name: "OpenAdmin",
        who: "For web hosts",
        lead: "The admin panel you use to run the server and your customers.",
        url: "https://demo.openpanel.com:2087/login",
        port: "2087",
        screenshot: "/img/openadmin-screenshots/001_dashboard-window.png",
        alt: "OpenAdmin dashboard",
        tryThis: [
            "Create a user and log in as them",
            "Set up hosting plans with limits",
            "Check services, firewall and security settings",
            "Change branding and OpenPanel settings",
        ],
        docs: "/docs/admin/intro/",
        accent: "from-[#FF9933] to-[#FF4C4D]",
    },
];

const Demo: React.FC = () => {
    return (
        <CommonLayout
            description="Try a live demo of OpenPanel and OpenAdmin online, explore the dashboard and admin panel before you install anything."
            image="/img/og/demo.png"
        >
            <Head>
                <title>Live Demo | OpenPanel</title>
                <meta property="og:title" content="Live Demo | OpenPanel" />
                <meta property="og:image:width" content="1200" />
                <meta property="og:image:height" content="630" />
                <meta property="og:image:type" content="image/png" />
                <html data-page="demo" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="not-prose flex flex-col gap-12 pt-8 lg:pt-16 pb-16 max-w-[928px] landing-lg:max-w-[1232px] w-full mx-auto px-4">
                    <section className="flex flex-col gap-4 max-w-[760px]">
                        <h1 className="text-3xl sm:text-5xl font-bold text-gray-900 dark:text-gray-0 m-0">
                            Try OpenPanel live
                        </h1>
                        <p className="text-base sm:text-lg text-gray-600 dark:text-gray-400 m-0">
                            Both panels run on a real server. The login details are already filled in, just click Log in and
                            look around, nothing to install or sign up for.
                        </p>
                    </section>

                    <section className="grid gap-6 lg:grid-cols-2">
                        {DEMOS.map((d) => (
                            <div
                                key={d.name}
                                className="rounded-2xl border border-gray-300 dark:border-gray-700 overflow-hidden flex flex-col"
                            >
                                <a
                                    href={d.url}
                                    className="block bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 aspect-[16/10] overflow-hidden"
                                >
                                    <img
                                        src={d.screenshot}
                                        alt={d.alt}
                                        className="w-full h-full object-cover object-left-top dark:hidden"
                                    />
                                    <img
                                        src={d.screenshot.replace(/\.png$/, "_dark.png")}
                                        alt={d.alt}
                                        className="w-full h-full object-cover object-left-top hidden dark:block"
                                    />
                                </a>
                                <div className="p-5 sm:p-6 flex flex-col gap-4 flex-1">
                                    <div className="flex flex-col gap-1">
                                        <span className="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">
                                            {d.who}
                                        </span>
                                        <h2 className="text-2xl font-bold m-0">
                                            <span className={clsx("text-transparent bg-clip-text bg-gradient-to-r", d.accent)}>
                                                {d.name}
                                            </span>
                                        </h2>
                                        <p className="text-sm text-gray-600 dark:text-gray-400 m-0">{d.lead}</p>
                                    </div>
                                    <div className="flex flex-col gap-2">
                                        <span className="text-sm font-semibold text-gray-900 dark:text-gray-100">Things to try</span>
                                        <ul className="m-0 p-0 list-none flex flex-col gap-1.5">
                                            {d.tryThis.map((t) => (
                                                <li key={t} className="flex gap-2 text-sm text-gray-700 dark:text-gray-300">
                                                    <span className="text-refine-blue dark:text-refine-cyan-alt font-bold">✓</span>
                                                    {t}
                                                </li>
                                            ))}
                                        </ul>
                                    </div>
                                    <div className="mt-auto flex flex-wrap items-center gap-3 pt-2">
                                        <a
                                            href={d.url}
                                            className="rounded-full px-5 py-2.5 text-sm font-semibold bg-refine-blue dark:bg-refine-cyan-alt text-gray-0 dark:text-gray-900 hover:brightness-110 no-underline"
                                        >
                                            Open {d.name} demo
                                        </a>
                                        <a
                                            href={d.docs}
                                            className="text-sm text-refine-react-light-link dark:text-refine-react-dark-link hover:underline"
                                        >
                                            {d.name} docs
                                        </a>
                                        <code className="ml-auto text-xs text-gray-500 dark:text-gray-400">
                                            demo.openpanel.com:{d.port}
                                        </code>
                                    </div>
                                </div>
                            </div>
                        ))}
                    </section>

                    <DemoInstallCta className="mt-4" />
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Demo;
