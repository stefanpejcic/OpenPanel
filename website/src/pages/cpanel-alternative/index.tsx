import Head from "@docusaurus/Head";
import Link from "@docusaurus/Link";
import clsx from "clsx";
import React from "react";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { EnterpriseHeroSection } from "@site/src/refine-theme/cpanel-hero-section";
import { EnterpriseGetInTouchCta } from "@site/src/refine-theme/enterprise-get-in-touch-cta";
import { EnterpriseFaq } from "@site/src/refine-theme/cpanel-faq";
import { EnterpriseTable } from "@site/src/refine-theme/cpanel-table";
import { LandingFooter } from "@site/src/refine-theme/landing-footer";
import { LandingTrustedByDevelopers } from "@site/src/refine-theme/landing-trusted-by-developers";
import { PlanFeaturesList } from "@site/src/refine-theme/plan-features-list";
import { ENTERPRISE_TRIAL_URL } from "@site/src/refine-theme/enterprise-get-in-touch-button";

const Enterprise: React.FC = () => {
    const title = "OpenPanel vs cPanel/WHM: Alternative Hosting Panel Compared";

    return (
        <>
            <Head>
                <html data-active-page="index" />
                <title>{title}</title>
                <meta property="og:title" content={title} />
            </Head>
            <CommonLayout description="Compare OpenPanel vs cPanel/WHM: Docker-based per-user isolation, native ARM support, and a fixed €14.95/month price per server - no per-account fees or hikes.">
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
                        <EnterpriseHeroSection
                            className={clsx(
                                sectionWidth,
                                sectionPadding,
                                "h-auto",
                                "mt-4 landing-sm:mt-8 landing-lg:mt-8",
                                "px-4 landing-sm:px-0",
                                "landing-lg:pr-12",
                            )}
                        />
                        <LandingTrustedByDevelopers
                            className={clsx(sectionPadding, sectionWidth)}
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
                            question="Try Enterprise free for 30 days?"
                            subtext="No credit card required."
                            buttonLabel="Start free trial"
                            buttonHref={ENTERPRISE_TRIAL_URL}
                            eventName="trial_click_cta_cpanel"
                        />
                        <div
                            className={clsx(
                                sectionPadding,
                                sectionWidth,
                                "text-center",
                            )}
                        >
                            <Link
                                to="mailto:info@openpanel.com"
                                className={clsx(
                                    "text-sm",
                                    "dark:text-gray-400 text-gray-600",
                                )}
                            >
                                Have questions first? Contact us at
                                info@openpanel.com
                            </Link>
                        </div>
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
