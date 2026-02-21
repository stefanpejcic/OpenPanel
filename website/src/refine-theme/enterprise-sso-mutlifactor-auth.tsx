import React from "react";
import clsx from "clsx";
import Translate from "@docusaurus/Translate";
import { CommonThemedImage } from "./common-themed-image";

export const EnterpriseSSOMultifactorAuth = ({
    className,
}: {
    className?: string;
}) => {
    return (
        <div className={clsx("flex flex-col", "not-prose", className)}>
            <div
                className={clsx(
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
                        srcDark="https://refine.ams3.cdn.digitaloceanspaces.com/enterprise/sign-in-dark.png"
                        srcLight="https://refine.ams3.cdn.digitaloceanspaces.com/enterprise/sign-in-light.png"
                    />
                    <div
                        className={clsx(
                            "flex flex-col",
                            "gap-4",
                            "p-4 landing-sm:p-10",
                            "not-prose",
                        )}
                    >
                        <h2
                            className={clsx(
                                "text-base landing-sm:text-2xl",
                                "dark:text-gray-300 text-gray-900",
                                "font-semibold",
                            )}
                        >
                            <Translate id="enterprise.sso_mfa.brand.title">Match your Brand</Translate>
                        </h2>
                        <p
                            className={clsx(
                                "text-base",
                                "dark:text-gray-400 text-gray-600",
                            )}
                        >
                            <Translate id="enterprise.sso_mfa.brand.desc">Customize the OpenPanel interface to match your brand.</Translate>
                        </p>
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
                        srcDark="https://refine.ams3.cdn.digitaloceanspaces.com/enterprise/multifactor-dark.png"
                        srcLight="https://refine.ams3.cdn.digitaloceanspaces.com/enterprise/multifactor-light.png"
                    />
                    <div
                        className={clsx(
                            "flex flex-col",
                            "gap-2 landing-sm:gap-4",
                            "p-4 landing-sm:p-10",
                            "not-prose",
                        )}
                    >
                        <h2
                            className={clsx(
                                "text-base landing-sm:text-2xl",
                                "dark:text-gray-300 text-gray-900",
                                "font-semibold",
                            )}
                        >
                            <Translate id="enterprise.sso_mfa.mfa.title">Multifactor Authentication</Translate>
                        </h2>
                        <p
                            className={clsx(
                                "text-base",
                                "dark:text-gray-400 text-gray-600",
                            )}
                        >
                            <Translate id="enterprise.sso_mfa.mfa.desc">Enable or enforce MFA for your clients.</Translate>
                        </p>
                    </div>
                </div>
            </div>
        </div>
    );
};
