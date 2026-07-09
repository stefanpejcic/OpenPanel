import Link from "@docusaurus/Link";
import React from "react";
import clsx from "clsx";
import {
    EnterpriseGetInTouchButton,
    ENTERPRISE_PURCHASE_URL,
    ENTERPRISE_TRIAL_URL,
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
                OpenPanel -{" "}
                <span
                    className={clsx(
                        "font-semibold",
                        "dark:text-refine-cyan-alt dark:drop-shadow-[0_0_30px_rgba(71,235,235,0.25)]",
                        "text-refine-blue drop-shadow-[0_0_30px_rgba(51,51,255,0.3)]",
                    )}
                >
                    cPanel
                </span>{" "}
                Alternative
            </h1>
            <p
                className={clsx(
                    "max-w-[446px]",
                    "mt-6",
                    "pl-4 landing-sm:pl-6 landing-md:pl-10",
                    "dark:text-gray-400 text-gray-600",
                )}
            >
                Built on Docker for true per-user isolation, with a fixed{" "}
                <span
                    className={clsx(
                        "font-semibold",
                        "dark:text-gray-0 text-gray-900",
                    )}
                >
                    €14.95/server/month
                </span>{" "}
                that never changes: no surprises, no per-account billing.
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
                        eventName="trial_click_hero_cpanel"
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
                                    "purchase_click_hero_cpanel",
                                );
                            }
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
    );
};
