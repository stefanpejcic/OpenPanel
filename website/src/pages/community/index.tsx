import Head from "@docusaurus/Head";
import Link from "@docusaurus/Link";
import clsx from "clsx";
import React from "react";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { SoftwareApplicationSchema } from "@site/src/refine-theme/software-application-schema";
import { EnterpriseHeroSection } from "@site/src/refine-theme/community-hero-section";
import { EnterpriseGetInTouchCta } from "@site/src/refine-theme/enterprise-get-in-touch-cta";
import { LandingTrustedByDevelopers } from "@site/src/refine-theme/landing-trusted-by-developers";
import { EnterpriseFaq } from "@site/src/refine-theme/community-faq";
import { EnterpriseTable } from "@site/src/refine-theme/enterprise-table";
import { LandingFooter } from "@site/src/refine-theme/landing-footer";
import { PlanFeaturesList } from "@site/src/refine-theme/plan-features-list";
import { HeaderDiscordIcon } from "@site/src/refine-theme/icons/header-discord";
import { FooterGithubIcon } from "@site/src/refine-theme/icons/footer-github";
import { MailIcon } from "@site/src/refine-theme/icons/mail";

const Enterprise: React.FC = () => {
    const title = "OpenPanel Community Edition | Free Web Hosting Panel";

    return (
        <>
            <Head>
                <html data-active-page="index" />
                <title>{title}</title>
                <meta property="og:title" content={title} />
            </Head>
            <SoftwareApplicationSchema
                name="OpenPanel Community"
                description="OpenPanel Community edition is a free web hosting control panel, suitable for VPS and private use."
                url="https://openpanel.com/community/"
                offers={{
                    price: "0",
                    priceCurrency: "EUR",
                    priceUnit: "free",
                }}
            />
            <CommonLayout description="OpenPanel Community edition is a free web hosting control panel, suitable for VPS and private use.">
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
                        <PlanFeaturesList
                            type="community"
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
                        />
                        <div className={clsx(sectionPadding, sectionWidth)}>
                            <div
                                className={clsx(
                                    "not-prose",
                                    "flex flex-col items-center text-center",
                                    "gap-4",
                                    "px-6 py-12 landing-sm:px-20 landing-sm:py-16",
                                    "dark:bg-landing-noise",
                                    "dark:bg-gray-800 bg-gray-50",
                                    "rounded-2xl landing-sm:rounded-3xl",
                                )}
                            >
                                <h2
                                    className={clsx(
                                        "text-2xl landing-sm:text-[32px] landing-sm:leading-[40px]",
                                        "tracking-tight",
                                        "font-semibold",
                                        "p-0",
                                        "dark:text-gray-0 text-gray-900",
                                    )}
                                >
                                    Join the OpenPanel community.
                                </h2>
                                <p
                                    className={clsx(
                                        "max-w-lg",
                                        "text-base",
                                        "dark:text-gray-400 text-gray-600",
                                    )}
                                >
                                    Get help, share feedback, or contribute on
                                    GitHub - OpenPanel is built in the open, with
                                    our users.
                                </p>
                                <div
                                    className={clsx(
                                        "mt-4",
                                        "flex flex-col landing-xs:flex-row",
                                        "items-center justify-center",
                                        "gap-4",
                                        "w-full landing-xs:w-auto",
                                    )}
                                >
                                    <Link
                                        to="https://discord.com/invite/7bNY8fANqF"
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        className={clsx(
                                            "w-full landing-xs:w-auto",
                                            "!text-gray-0 dark:!text-gray-900",
                                            "bg-refine-blue dark:bg-refine-cyan-alt",
                                            "transition-[filter]",
                                            "duration-150",
                                            "ease-in-out",
                                            "hover:brightness-110",
                                            "hover:!no-underline",
                                            "rounded-3xl",
                                            "py-3 px-8",
                                            "flex items-center justify-center gap-2",
                                            "text-base font-semibold",
                                        )}
                                    >
                                        <HeaderDiscordIcon width={20} height={20} />
                                        Join Discord
                                    </Link>
                                    <Link
                                        to="https://github.com/stefanpejcic/openpanel"
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        className={clsx(
                                            "w-full landing-xs:w-auto",
                                            "dark:text-refine-cyan-alt text-refine-blue",
                                            "dark:bg-refine-cyan-alt/10 bg-refine-blue/10",
                                            "transition-[filter]",
                                            "duration-150",
                                            "ease-in-out",
                                            "hover:brightness-110",
                                            "hover:!no-underline",
                                            "rounded-3xl",
                                            "py-3 px-8",
                                            "flex items-center justify-center gap-2",
                                            "text-base font-semibold",
                                        )}
                                    >
                                        <FooterGithubIcon width={20} height={20} />
                                        GitHub
                                    </Link>
                                </div>
                                <Link
                                    to="mailto:info@openpanel.com"
                                    className={clsx(
                                        "mt-2",
                                        "flex items-center gap-2",
                                        "text-sm",
                                        "dark:text-gray-400 text-gray-600",
                                        "hover:!no-underline",
                                    )}
                                >
                                    <MailIcon width={16} height={14} />
                                    info@openpanel.com
                                </Link>
                            </div>
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
