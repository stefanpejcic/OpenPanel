import Link from "@docusaurus/Link";
import React from "react";
import clsx from "clsx";
import {
    EnterpriseGetInTouchButton,
    ENTERPRISE_PURCHASE_URL,
    ENTERPRISE_TRIAL_URL,
    gtagReportConversion,
} from "./enterprise-get-in-touch-button";

export const EnterpriseHeroSection = ({
    className,
}: {
    className?: string;
}) => {
    return (
        <div
            className={clsx(
                "flex flex-col",
                "not-prose",
                className,
            )}
        >
            <div className={clsx("flex flex-col")}>
                <h1
                    className={clsx(
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
                        "mt-4 landing-lg:mt-6",
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
                            onClick={(e) => {
                                e.preventDefault();
                                if (
                                    typeof window !== "undefined" &&
                                    typeof window.gtag !== "undefined"
                                ) {
                                    window.gtag(
                                        "event",
                                        "purchase_click_hero",
                                    );
                                }
                                gtagReportConversion(ENTERPRISE_PURCHASE_URL);
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
        </div>
    );
};
