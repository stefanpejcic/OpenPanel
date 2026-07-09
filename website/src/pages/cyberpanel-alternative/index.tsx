import Head from "@docusaurus/Head";
import clsx from "clsx";
import React from "react";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { EnterpriseHeroSection } from "@site/src/refine-theme/cyberpanel-hero-section";
import { EnterpriseGetInTouchCta } from "@site/src/refine-theme/enterprise-get-in-touch-cta";
import { EnterpriseFaq } from "@site/src/refine-theme/cyberpanel-faq";
import { EnterpriseTable } from "@site/src/refine-theme/cyberpanel-table";
import { LandingFooter } from "@site/src/refine-theme/landing-footer";

const Enterprise: React.FC = () => {
    const title = "OpenPanel Enterprise | CyberPanel Alternative Web Hosting Panel";

    return (
        <>
            <Head>
                <html data-active-page="index" />
                <title>{title}</title>
                <meta property="og:title" content={title} />
            </Head>
            <CommonLayout description="OpenPanel Enterprise is a CyberPanel alternative.">
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
                                "h-auto landing-md:h-[432px]",
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
