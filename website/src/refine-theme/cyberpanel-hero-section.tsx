import React from "react";
import clsx from "clsx";
import { EnterpriseGetInTouchButton } from "./enterprise-get-in-touch-button";

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
                    CyberPanel
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
                Built on Docker for true per-user isolation, with a security track record and a fixed price that never changes.{" "}
            </p>
            <EnterpriseGetInTouchButton
                className={clsx(
                    "pl-4 landing-sm:pl-6 landing-md:pl-10",
                    "mt-6 landing-lg:mt-16",
                )}
            />
        </div>
    );
};
