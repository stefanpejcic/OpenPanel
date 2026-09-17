import Head from "@docusaurus/Head";
import Link from "@docusaurus/Link";
import clsx from "clsx";
import React from "react";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { SoftwareApplicationSchema } from "@site/src/refine-theme/software-application-schema";
import { EnterpriseHeroSection } from "@site/src/refine-theme/plesk-hero-section";
import { PleskIntroContent } from "@site/src/refine-theme/plesk-intro-content";
import { EnterpriseGetInTouchCta } from "@site/src/refine-theme/enterprise-get-in-touch-cta";
import { EnterpriseFaq } from "@site/src/refine-theme/plesk-faq";
import { EnterpriseTable } from "@site/src/refine-theme/plesk-table";
import { LandingFooter } from "@site/src/refine-theme/landing-footer";
import { ENTERPRISE_TRIAL_URL } from "@site/src/refine-theme/enterprise-get-in-touch-button";

const Enterprise: React.FC = () => {
    const title = "OpenPanel vs Plesk: Alternative Hosting Panel Compared";

    return (
        <>
            <Head>
                <html data-active-page="index" />
                <title>{title}</title>
                <meta property="og:title" content={title} />
            </Head>
            <SoftwareApplicationSchema
                name="OpenPanel"
                description="OpenPanel is a Plesk alternative hosting control panel with Podman-based per-user isolation, native ARM support, and a fixed price per server."
                url="https://openpanel.com/plesk-alternative/"
                offers={{
                    price: "14.95",
                    priceCurrency: "EUR",
                    priceUnit: "per server per month",
                }}
            />
            <CommonLayout description="Compare OpenPanel vs Plesk: true per-user Podman isolation, native ARM support, and one fixed €14.95/month price per server - no per-account licensing fees.">
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

                        <div className={clsx("w-full", "overflow-hidden")}>
                            <EnterpriseTable
                                className={clsx(sectionWidth, sectionPadding)}
                            />
                        </div>
                        <PleskIntroContent
                            className={clsx(
                                sectionPadding,
                                "px-4 landing-sm:px-10 landing-lg:px-0",
                                "w-full landing-lg:max-w-[792px] mx-auto",
                            )}
                        />
                        <div className={clsx(sectionPadding, sectionWidth)}>
                            <div
                                className={clsx(
                                    "pl-4 landing-sm:pl-6 landing-md:pl-10",
                                    "text-2xl landing-sm:text-[32px] landing-sm:leading-[40px]",
                                )}
                            >
                                <h2
                                    className={clsx(
                                        "font-semibold",
                                        "dark:text-gray-400 text-gray-600",
                                    )}
                                >
                                    See how OpenPanel compares to other
                                    hosting panels
                                </h2>
                            </div>
                            <div
                                className={clsx(
                                    "grid",
                                    "grid-cols-2 landing-sm:grid-cols-4",
                                    "gap-4",
                                    "mt-8 landing-md:mt-12",
                                )}
                            >
                                {[
                                    {
                                        label: "vs cPanel",
                                        to: "/cpanel-alternative/",
                                    },
                                    {
                                        label: "vs DirectAdmin",
                                        to: "/directadmin-alternative/",
                                    },
                                    {
                                        label: "vs CyberPanel",
                                        to: "/cyberpanel-alternative/",
                                    },
                                    {
                                        label: "OpenPanel Enterprise",
                                        to: "/enterprise/",
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
                            question="Try Enterprise free for 30 days?"
                            subtext="No credit card required."
                            buttonLabel="Start free trial"
                            buttonHref={ENTERPRISE_TRIAL_URL}
                            eventName="trial_click_cta_plesk"
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
