import React from "react";
import clsx from "clsx";
import Translate from "@docusaurus/Translate";
import { CommonThemedImage } from "./common-themed-image";

export const EnterpriseFrequentUpdates = ({
    className,
}: {
    className?: string;
}) => {
    return (
        <div className={clsx("flex flex-col", "not-prose", className)}>
            <div
                className={clsx(
                    "pl-4 landing-sm:pl-6 landing-md:pl-10",
                    "text-2xl landing-sm:text-[32px] landing-sm:leading-[40px]",
                )}
            >
                <h2 className={clsx("dark:text-gray-0 text-gray-900")}>
                    <Translate id="enterprise.frequent_updates.title">Get frequent </Translate>
                    <span
                        className={clsx(
                            "font-semibold",
                            "dark:text-refine-cyan-alt dark:drop-shadow-[0_0_30px_rgba(71,235,235,0.25)]",
                            "text-refine-blue drop-shadow-[0_0_30px_rgba(51,51,255,0.3)]",
                        )}
                    >
                        <Translate id="enterprise.frequent_updates.updates">updates</Translate>
                    </span>
                    .
                </h2>
            </div>

            <div
                className={clsx(
                    "mt-8 landing-sm:mt-12 landing-md:mt-20",
                    "grid",
                    "grid-cols-1 landing-md:grid-cols-2",
                    "gap-8 landing-md:gap-12 landing-lg:gap-6",
                )}
            >
                <div
                    className={clsx(
                        "flex flex-col",
                        "dark:bg-landing-noise",
                        "dark:bg-gray-800 bg-gray-50",
                        "rounded-2xl landing-sm:rounded-3xl",
                    )}
                >
                    <CommonThemedImage
                        className={clsx("rounded-2xl landing-sm:rounded-3xl")}
                        srcDark="https://refine.ams3.cdn.digitaloceanspaces.com/enterprise/continuous-updates-dark.png"
                        srcLight="https://refine.ams3.cdn.digitaloceanspaces.com/enterprise/continuous-updates-light.png"
                    />
                    <div
                        className={clsx(
                            "flex flex-col",
                            "gap-2 landing-sm:gap-4",
                            "p-4 landing-sm:p-10",
                            "not-prose",
                        )}
                    >
                        <div
                            className={clsx(
                                "text-base landing-sm:text-2xl",
                                "dark:text-gray-300 text-gray-900",
                                "font-semibold",
                            )}
                        >
                            <Translate id="enterprise.frequent_updates.continuous.title">Continuous updates</Translate>
                        </div>
                        <div
                            className={clsx(
                                "text-base",
                                "dark:text-gray-400 text-gray-600",
                                "mt-2 landing-sm:mt-4",
                            )}
                        >
                            <Translate id="enterprise.frequent_updates.continuous.desc">Continuous OpenPanel updates with the latest features, bug fixes, and security patches.</Translate>
                        </div>
                    </div>
                </div>

                <div
                    className={clsx(
                        "flex flex-col",
                        "dark:bg-landing-noise",
                        "dark:bg-gray-800 bg-gray-50",
                        "rounded-2xl landing-sm:rounded-3xl",
                    )}
                >
                    <CommonThemedImage
                        className={clsx("rounded-2xl landing-sm:rounded-3xl")}
                        srcDark="https://refine.ams3.cdn.digitaloceanspaces.com/enterprise/container-updates-dark.png"
                        srcLight="https://refine.ams3.cdn.digitaloceanspaces.com/enterprise/container-updates-light.png"
                    />
                    <div
                        className={clsx(
                            "flex flex-col",
                            "gap-2 landing-sm:gap-4",
                            "p-4 landing-sm:p-10",
                            "not-prose",
                        )}
                    >
                        <div
                            className={clsx(
                                "text-base landing-sm:text-2xl",
                                "dark:text-gray-300 text-gray-900",
                                "font-semibold",
                            )}
                        >
                            <Translate id="enterprise.frequent_updates.container.title">Container Updates</Translate>
                        </div>
                        <div
                            className={clsx(
                                "text-base",
                                "dark:text-gray-400 text-gray-600",
                                "mt-2 landing-sm:mt-4",
                            )}
                        >
                            <Translate id="enterprise.frequent_updates.container.desc">In-container updates that automatically apply security patches, update the OS and services within user Docker containers without manual intervention.</Translate>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};
