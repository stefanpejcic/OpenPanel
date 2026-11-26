import React from "react";
import Head from "@docusaurus/Head";
import Link from "@docusaurus/Link";
import clsx from "clsx";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { HeaderDiscordIcon } from "@site/src/refine-theme/icons/header-discord";
import { inProgress, plannedFeatures } from "@site/src/assets/roadmap";
import { HIGHLIGHTS } from "@site/src/components/changelog-overview";

const card = "rounded-2xl border border-gray-300 dark:border-gray-700 p-5 sm:p-6";
const h2 = "text-2xl sm:text-3xl font-bold text-gray-900 dark:text-gray-0 m-0";
const lead = "text-base text-gray-600 dark:text-gray-400 m-0";
const link = "text-refine-react-light-link dark:text-refine-react-dark-link hover:underline";
const primaryBtn =
    "inline-flex items-center gap-2 rounded-full px-5 py-2.5 text-sm font-semibold bg-refine-blue dark:bg-refine-cyan-alt text-gray-0 dark:text-gray-900 hover:brightness-110 no-underline";
const secondaryBtn =
    "inline-flex items-center gap-2 rounded-full px-5 py-2.5 text-sm font-semibold border border-gray-300 dark:border-gray-700 text-gray-900 dark:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-800 no-underline";

const Roadmap: React.FC = () => {
    return (
        <CommonLayout
            description="See what OpenPanel is building next - features in active development and what's planned, updated manually by our team."
            image="/img/og/roadmap.png"
        >
            <Head>
                <title>Roadmap | OpenPanel</title>
                <meta property="og:title" content="Roadmap | OpenPanel" />
                <meta property="og:image:width" content="1200" />
                <meta property="og:image:height" content="630" />
                <meta property="og:image:type" content="image/png" />
                <html data-page="roadmap" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="not-prose flex flex-col gap-14 pt-8 lg:pt-16 pb-32 max-w-[928px] landing-lg:max-w-[1232px] w-full mx-auto px-4">
                    <section className="flex flex-col gap-4">
                        <h1 className="text-3xl sm:text-5xl font-bold text-gray-900 dark:text-gray-0 m-0">Roadmap</h1>
                        <p className="text-base sm:text-lg text-gray-600 dark:text-gray-400 m-0 max-w-[720px]">
                            What we&apos;re building next. Nothing is set in stone, priorities can shift based on community feedback.
                        </p>
                    </section>

                    {inProgress.length > 0 && (
                        <section className="flex flex-col gap-5">
                            <div className="flex flex-col gap-2">
                                <h2 className={h2}>In progress</h2>
                                <p className={lead}>What our team is building right now, shipping in the next releases.</p>
                            </div>
                            <div className="grid gap-4 sm:grid-cols-2">
                                {inProgress.map((f) => (
                                    <div key={f.title} className={clsx(card, "flex gap-4")}>
                                        <div className="shrink-0">{f.icon}</div>
                                        <div className="flex flex-col gap-1.5">
                                            <span className="self-start rounded-full px-2 py-0.5 text-[11px] font-semibold bg-refine-blue/10 text-refine-blue dark:bg-refine-cyan-alt/10 dark:text-refine-cyan-alt">
                                                In progress
                                            </span>
                                            <h3 className="text-base font-semibold text-gray-900 dark:text-gray-0 m-0">{f.title}</h3>
                                            <p className="text-sm text-gray-600 dark:text-gray-400 m-0">{f.description}</p>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </section>
                    )}

                    <section className="flex flex-col gap-5">
                        <div className="flex flex-col gap-2">
                            <h2 className={h2}>Planned</h2>
                            <p className={lead}>Up next, in no particular order.</p>
                        </div>
                        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                            {plannedFeatures.map((f) => (
                                <div key={f.title} className={clsx(card, "flex items-center gap-4")}>
                                    <div className="shrink-0">{f.icon}</div>
                                    <h3 className="text-base font-semibold text-gray-900 dark:text-gray-0 m-0">{f.title}</h3>
                                </div>
                            ))}
                        </div>
                    </section>

                    <section className="flex flex-col gap-5">
                        <div className="flex flex-col gap-2">
                            <h2 className={h2}>Recently shipped</h2>
                            <p className={lead}>
                                The biggest changes from the last releases, see the{" "}
                                <Link to="/docs/changelog/intro/" className={link}>
                                    changelog
                                </Link>{" "}
                                for everything.
                            </p>
                        </div>
                        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                            {HIGHLIGHTS.map(([v, title, desc]) => (
                                <Link
                                    key={title}
                                    to={`/docs/changelog/${v}/`}
                                    className={clsx(card, "flex flex-col gap-1.5 no-underline hover:border-gray-400 dark:hover:border-gray-500")}
                                >
                                    <span className="self-start rounded-full px-2 py-0.5 text-[11px] font-semibold font-mono bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300">
                                        Shipped in {v}
                                    </span>
                                    <span className="text-base font-semibold text-gray-900 dark:text-gray-0">{title}</span>
                                    <span className="text-sm text-gray-600 dark:text-gray-400">{desc}</span>
                                </Link>
                            ))}
                        </div>
                    </section>

                    <section className={clsx(card, "flex flex-col sm:flex-row gap-4 sm:items-center sm:justify-between sm:!p-8")}>
                        <div className="flex flex-col gap-1">
                            <h2 className="text-xl font-bold text-gray-900 dark:text-gray-0 m-0">Want to shape the roadmap?</h2>
                            <p className="text-sm text-gray-600 dark:text-gray-400 m-0 max-w-[620px]">
                                Join our community and tell us what matters most to your hosting business, operator feedback drives
                                what we build next.
                            </p>
                        </div>
                        <div className="flex flex-wrap gap-3">
                            <Link to="https://github.com/stefanpejcic/OpenPanel/discussions" className={primaryBtn}>
                                Suggest a feature
                            </Link>
                            <Link to="https://discord.com/invite/7bNY8fANqF" className={secondaryBtn}>
                                <HeaderDiscordIcon className="w-4 h-4" /> Join Discord
                            </Link>
                        </div>
                    </section>
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Roadmap;
