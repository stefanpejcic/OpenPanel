import React from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { DemoInstallCta } from "@site/src/refine-theme/demo-install-cta";

const DEMOS = [
    {
        name: "OpenPanel",
        lead: "The panel your customers log in to. Try adding a domain, installing WordPress or creating an email account and a database.",
        url: "https://demo.openpanel.com:2083/login",
        port: "2083",
        screenshot: "/img/openpanel-screenshots/dashboard/dashboard-window.png",
        alt: "OpenPanel dashboard",
        docs: "/docs/panel/intro/",
    },
    {
        name: "OpenAdmin",
        lead: "The admin side, for you as the host. Try creating a user and logging in as them, setting up a hosting plan, or changing the branding.",
        url: "https://demo.openpanel.com:2087/login",
        port: "2087",
        screenshot: "/img/openadmin-screenshots/001_dashboard-window.png",
        alt: "OpenAdmin dashboard",
        docs: "/docs/admin/intro/",
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
                            Both panels run on a real server with test data. The login is already filled in, so just click Log in.
                            Nothing to install or sign up for.
                        </p>
                    </section>

                    <section className="grid gap-12 lg:gap-10 lg:grid-cols-2">
                        {DEMOS.map((d) => (
                            <div key={d.name} className="flex flex-col gap-4">
                                <a
                                    href={d.url}
                                    className="block rounded-lg border border-gray-300 dark:border-gray-700 overflow-hidden no-underline"
                                >
                                    <img src={d.screenshot} alt={d.alt} className="block w-full aspect-[16/9] object-cover object-left-top dark:hidden" />
                                    <img
                                        src={d.screenshot.replace(/\.png$/, "_dark.png")}
                                        alt={d.alt}
                                        className="w-full aspect-[16/9] object-cover object-left-top hidden dark:block"
                                    />
                                </a>
                                <h2 className="text-2xl font-bold m-0 text-gray-900 dark:text-gray-0">
                                    {d.name}{" "}
                                    <span className="text-sm font-normal font-jetBrains-mono text-gray-500 dark:text-gray-400">
                                        demo.openpanel.com:{d.port}
                                    </span>
                                </h2>
                                <p className="text-base text-gray-600 dark:text-gray-400 m-0">{d.lead}</p>
                                <p className="text-sm m-0 flex flex-wrap gap-x-5 gap-y-1">
                                    <a
                                        href={d.url}
                                        className="font-semibold text-refine-react-light-link dark:text-refine-react-dark-link hover:underline"
                                    >
                                        Open the {d.name} demo
                                    </a>
                                    <a href={d.docs} className={"text-gray-600 dark:text-gray-400 hover:underline"}>
                                        Read the {d.name} docs
                                    </a>
                                </p>
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
