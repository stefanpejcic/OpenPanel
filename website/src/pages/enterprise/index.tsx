import Head from "@docusaurus/Head";
import Link from "@docusaurus/Link";
import clsx from "clsx";
import React from "react";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { SoftwareApplicationSchema } from "@site/src/refine-theme/software-application-schema";
import { EnterpriseHeroSection } from "@site/src/refine-theme/enterprise-hero-section";
import { EnterpriseGetSupport } from "@site/src/refine-theme/enterprise-get-support";
import { EnterpriseSecurity } from "@site/src/refine-theme/enterprise-secuity";
import { EnterpriseGetInTouchCta } from "@site/src/refine-theme/enterprise-get-in-touch-cta";
import { EnterpriseFlexibility } from "@site/src/refine-theme/enterprise-flexibility";
import { EnterpriseDataSource } from "@site/src/refine-theme/enterprise-data-source";
import { LandingTrustedByDevelopers } from "@site/src/refine-theme/landing-trusted-by-developers";
import { EnterpriseFaq } from "@site/src/refine-theme/enterprise-faq";
import { EnterpriseTable } from "@site/src/refine-theme/enterprise-table";
import { LandingFooter } from "@site/src/refine-theme/landing-footer";
import { ENTERPRISE_TRIAL_URL } from "@site/src/refine-theme/enterprise-get-in-touch-button";
import { PlanFeaturesList } from "@site/src/refine-theme/plan-features-list";

const Enterprise: React.FC = () => {
    const title = "OpenPanel Enterprise | Next Generation Hosting Panel";

    return (
        <>
            <Head>
                <html data-active-page="index" />
                <title>{title}</title>
                <meta property="og:title" content={title} />
            </Head>
            <SoftwareApplicationSchema
                name="OpenPanel Enterprise"
                description="OpenPanel Enterprise Edition provides robust user isolation and management features, designed for web hosting providers, all at a fixed price."
                url="https://openpanel.com/enterprise/"
                offers={{
                    price: "14.95",
                    priceCurrency: "EUR",
                    priceUnit: "per server per month",
                }}
            />
            <CommonLayout description="OpenPanel Enterprise Edition provides robust user isolation and management features, designed for web hosting providers, all at a fixed price.">
                <div className={clsx()}>
                    <CommonHeader />
                    <div
                        className={clsx(
                            "flex flex-col",
                            "gap-12 landing-sm:gap-20 landing-md:gap-28 landing-lg:gap-40",
                            "pb-12 landing-sm:pb-16 landing-md:pb-20 landing-lg:pb-40",
                            "mx-auto",
                        )}
                    >
                        <div>
                            <EnterpriseHeroSection
                                className={clsx(
                                    sectionWidth,
                                    sectionPadding,
                                    "h-auto landing-md:h-[432px]",
                                    "mt-4 landing-sm:mt-8 landing-lg:mt-8",
                                    "px-4 landing-sm:px-0",
                                    "landing-lg:pr-12",
                                )}
                            />
                            <LandingTrustedByDevelopers
                                className={clsx(
                                    sectionPadding,
                                    sectionWidth,
                                    "mt-12 landing-sm:mt-20 landing-md:mt-28 landing-lg:mt-10",
                                )}
                            />
                        </div>
                        <EnterpriseGetSupport
                            className={clsx(sectionPadding, sectionWidth)}
                        />
                        <EnterpriseSecurity
                            className={clsx(sectionPadding, sectionWidth)}
                        />
                        <EnterpriseGetInTouchCta
                            className={clsx(
                                sectionPadding,
                                sectionWidth,
                                "w-full landing-lg:max-w-[792px] mx-auto",
                            )}
                            question="Try Enterprise free for 14 days?"
                            subtext="No credit card required."
                            buttonLabel="Start free trial"
                            buttonHref={ENTERPRISE_TRIAL_URL}
                            eventName="trial_click_cta1"
                        />
                        <EnterpriseFlexibility
                            className={clsx(sectionPadding, sectionWidth)}
                        />
                        <EnterpriseDataSource
                            className={clsx(sectionPadding, sectionWidth)}
                        />
                        <EnterpriseGetInTouchCta
                            className={clsx(
                                sectionPadding,
                                sectionWidth,
                                "w-full landing-lg:max-w-[792px] mx-auto",
                            )}
                            question="Try Enterprise free for 14 days?"
                            subtext="No credit card required."
                            buttonLabel="Start free trial"
                            buttonHref={ENTERPRISE_TRIAL_URL}
                            eventName="trial_click_cta2"
                        />
                        <div className={clsx("w-full", "overflow-hidden")}>
                            <EnterpriseTable
                                className={clsx(sectionWidth, sectionPadding)}
                            />
                        </div>
                        <PlanFeaturesList
                            type="enterprise"
                            className={clsx(sectionPadding, sectionWidth)}
                        />
                        <div className={clsx(sectionPadding, sectionWidth)}>
                            <div
                                className={clsx(
                                    "pl-4 landing-sm:pl-6 landing-md:pl-10",
                                    "text-2xl landing-sm:text-[32px] landing-sm:leading-[40px]",
                                )}
                            >
                                <h2 className={clsx("dark:text-gray-0 text-gray-900")}>
                                    See how we{" "}
                                    <span
                                        className={clsx(
                                            "font-semibold",
                                            "dark:text-refine-cyan-alt dark:drop-shadow-[0_0_30px_rgba(71,235,235,0.25)]",
                                            "text-refine-blue drop-shadow-[0_0_30px_rgba(51,51,255,0.3)]",
                                        )}
                                    >
                                        compare
                                    </span>
                                    .
                                </h2>
                            </div>
                            <div
                                className={clsx(
                                    "grid",
                                    "grid-cols-2 landing-sm:grid-cols-4",
                                    "gap-4",
                                    "mt-8 landing-md:mt-20",
                                )}
                            >
                                {[
                                    {
                                        label: "vs cPanel",
                                        to: "/cpanel-alternative/",
                                    },
                                    {
                                        label: "vs Plesk",
                                        to: "/plesk-alternative/",
                                    },
                                    {
                                        label: "vs DirectAdmin",
                                        to: "/directadmin-alternative/",
                                    },
                                    {
                                        label: "vs CyberPanel",
                                        to: "/cyberpanel-alternative/",
                                    },
                                ].map((item) => (
                                    <Link
                                        key={item.to}
                                        to={item.to}
                                        className={clsx(
                                            "not-prose",
                                            "flex items-center justify-center text-center",
                                            "p-4 landing-sm:p-6",
                                            "dark:bg-landing-noise",
                                            "dark:bg-gray-800 bg-gray-50",
                                            "rounded-2xl",
                                            "text-sm landing-sm:text-base",
                                            "font-semibold",
                                            "text-gray-900 dark:text-gray-0",
                                            "transition-[filter]",
                                            "duration-150",
                                            "ease-in-out",
                                            "hover:brightness-110",
                                            "hover:!no-underline",
                                        )}
                                    >
                                        {item.label}
                                    </Link>
                                ))}
                            </div>
                        </div>
                        <EnterpriseFaq
                            className={clsx(
                                sectionPadding,
                                "px-4 landing-sm:px-10 landing-lg:px-0",
                                "w-full landing-lg:max-w-[792px] mx-auto",
                            )}
                        />
                        <EnterpriseGetInTouchCta
                            className={clsx(
                                sectionPadding,
                                sectionWidth,
                                "landing-lg:max-w-[792px]",
                            )}
                            question="Still deciding? Start your free trial today."
                            subtext="No credit card required."
                            buttonLabel="Start free trial"
                            buttonHref={ENTERPRISE_TRIAL_URL}
                            eventName="trial_click_cta3"
                        />
                    </div>
                    <LandingFooter />
                </div>
            </CommonLayout>
        </>
    );
};

const sectionPadding = clsx("px-2 landing-sm:px-0");

const sectionWidth = clsx(
    "mx-auto",
    "w-full",
    "max-w-[592px]",
    "landing-sm:max-w-[656px]",
    "landing-md:max-w-[896px]",
    "landing-lg:max-w-[1200px]",
);

export default Enterprise;
