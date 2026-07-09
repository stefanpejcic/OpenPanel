import Link from "@docusaurus/Link";
import React from "react";
import clsx from "clsx";
import {
    EnterpriseGetInTouchButton,
    ENTERPRISE_PURCHASE_URL,
    ENTERPRISE_TRIAL_URL,
    fireGoogleAdsConversion,
} from "./enterprise-get-in-touch-button";
import { CommonThemedImage } from "./common-themed-image";

export const EnterpriseHeroSection = ({
    className,
}: {
    className?: string;
}) => {
    return (
        <div
            className={clsx(
                "flex flex-col",
                "landing-md:grid landing-md:grid-cols-12",
                "not-prose",
                className,
            )}
        >
            <div className={clsx("flex flex-col", "col-start-1 col-end-8")}>
                <h1
                    className={clsx(
                        "max-w-xl landing-md:max-w-[408px] landing-lg:max-w-non landing-lg:whitespace-nowrap",
                        "text-[32px] leading-[40px] landing-sm:text-[56px] landing-sm:leading-[72px]",
                        "tracking-tight",
                        "text-start",
                        "pl-4 landing-sm:pl-6 landing-md:pl-10",
                        "dark:text-gray-0 text-gray-900",
                        "landing-lg:pt-8",
                    )}
                >
                    OpenPanel{" "}
                    <span
                        className={clsx(
                            "font-semibold",
                            "text-transparent bg-clip-text bg-gradient-to-r text-gradient-to-r",
                            "from-[#0FBDBD] to-[#26D97F]",
                        )}
                    >
                        Enterprise
                    </span>
                    .
                </h1>
                <p
                    className={clsx(
                        "max-w-[446px]",
                        "mt-6",
                        "pl-4 landing-sm:pl-6 landing-md:pl-10",
                        "dark:text-gray-400 text-gray-600",
                    )}
                >
                    Robust user isolation and management for web hosting
                    providers, at a fixed{" "}
                    <span
                        className={clsx(
                            "font-semibold",
                            "dark:text-gray-0 text-gray-900",
                        )}
                    >
                        €14.95/server/month
                    </span>{" "}
                    — unlimited accounts, no per-user fees.
                </p>
                <div
                    className={clsx(
                        "flex flex-col",
                        "gap-3",
                        "pl-4 landing-sm:pl-6 landing-md:pl-10",
                        "mt-6 landing-lg:mt-16",
                    )}
                >
                    <div
                        className={clsx(
                            "flex flex-col landing-sm:flex-row",
                            "items-start landing-sm:items-center",
                            "gap-3 landing-sm:gap-6",
                        )}
                    >
                        <EnterpriseGetInTouchButton
                            label="Start free 30-day trial"
                            href={ENTERPRISE_TRIAL_URL}
                            eventName="trial_click_hero"
                        />
                        <Link
                            href={ENTERPRISE_PURCHASE_URL}
                            target="_self"
                            rel="noopener noreferrer"
                            onClick={() => {
                                if (
                                    typeof window !== "undefined" &&
                                    typeof window.gtag !== "undefined"
                                ) {
                                    window.gtag(
                                        "event",
                                        "purchase_click_hero",
                                    );
                                }
                                fireGoogleAdsConversion();
                            }}
                            className={clsx(
                                "text-sm font-semibold",
                                "dark:text-gray-0 text-gray-900",
                                "underline underline-offset-4",
                            )}
                        >
                            or purchase a license directly →
                        </Link>
                    </div>
                    <p
                        className={clsx(
                            "text-xs",
                            "dark:text-gray-500 text-gray-500",
                        )}
                    >
                        No credit card required · Cancel anytime
                    </p>
                </div>
            </div>
            <div
                className={clsx(
                    "flex",
                    "justify-end",
                    "col-start-8",
                    "col-end-13",
                    "mt-12 landing-sm:mt-16 landing-md:mt-0",
                )}
            >
                <CommonThemedImage
                    className={clsx(
                        "landing-md:h-[360px] landing-md:w-[326px]",
                        "landing-md:h-[360px] landing-md:w-[326px]",
                    )}
                    srcDark="/img/hero.png"
                    srcLight="/img/hero.png"
                    alt="OpenPanel Enterprise dashboard preview"
                />
            </div>
        </div>
    );
};
