import Link from "@docusaurus/Link";
import React, { useEffect, useState } from "react";
import clsx from "clsx";

type Feature = {
    name: string;
    title: string;
    description: string;
    link: string;
    type: string;
    help_link?: string;
};

const FEATURES_URL =
    "https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openadmin/config/features.json";

const formatDescription = (description: string) => {
    const stripped = description.replace(/^if enabled,\s*/i, "");
    return stripped.charAt(0).toUpperCase() + stripped.slice(1);
};

export const PlanFeaturesList = ({
    type,
    className,
}: {
    type: "community" | "enterprise";
    className?: string;
}) => {
    const [features, setFeatures] = useState<Feature[]>([]);

    useEffect(() => {
        let cancelled = false;

        fetch(FEATURES_URL)
            .then((response) => response.json())
            .then((data: Feature[]) => {
                if (!cancelled) {
                    setFeatures(data.filter((feature) => feature.type === type));
                }
            })
            .catch((error) => {
                console.error("Failed to fetch features:", error);
            });

        return () => {
            cancelled = true;
        };
    }, [type]);

    if (features.length === 0) {
        return null;
    }

    return (
        <div className={clsx("flex flex-col", "not-prose", className)}>
            <div
                className={clsx(
                    "pl-4 landing-sm:pl-6 landing-md:pl-10",
                    "text-2xl landing-sm:text-[32px] landing-sm:leading-[40px]",
                )}
            >
                <h2 id="modules" className={clsx("dark:text-gray-0 text-gray-900")}>
                    Every{" "}
                    <span
                        className={clsx(
                            "font-semibold",
                            "dark:text-refine-cyan-alt dark:drop-shadow-[0_0_30px_rgba(71,235,235,0.25)]",
                            "text-refine-blue drop-shadow-[0_0_30px_rgba(51,51,255,0.3)]",
                        )}
                    >
                        {type === "enterprise" ? "Enterprise" : "Community"}
                    </span>{" "}
                    module.
                </h2>
                <p
                    className={clsx(
                        "mt-4 landing-sm:mt-6",
                        "text-base",
                        "dark:text-gray-400 text-gray-600",
                    )}
                >
                    {type === "enterprise" ? (
                        <>
                            Everything in{" "}
                            <Link
                                href="/community/#modules"
                                className={clsx(
                                    "dark:text-refine-cyan-alt text-refine-blue",
                                    "font-semibold",
                                )}
                            >
                                Community
                            </Link>
                            , plus these Enterprise-only features.
                        </>
                    ) : (
                        <>
                            These are the features included in the free
                            Community edition. Need more? See what&apos;s
                            included in{" "}
                            <Link
                                href="/enterprise/#modules"
                                className={clsx(
                                    "dark:text-refine-cyan-alt text-refine-blue",
                                    "font-semibold",
                                )}
                            >
                                Enterprise
                            </Link>
                            .
                        </>
                    )}
                </p>
            </div>
            <div
                className={clsx(
                    "grid",
                    "grid-cols-1 landing-sm:grid-cols-2 landing-lg:grid-cols-3",
                    "gap-4",
                    "mt-8 landing-md:mt-20",
                )}
            >
                {features.map((feature) => (
                    <a
                        key={feature.name}
                        href={feature.help_link}
                        target={feature.help_link ? "_blank" : undefined}
                        rel={feature.help_link ? "noopener noreferrer" : undefined}
                        className={clsx(
                            "flex flex-col",
                            "gap-1",
                            "p-4 landing-sm:p-6",
                            "dark:bg-landing-noise",
                            "dark:bg-gray-800 bg-gray-50",
                            "rounded-2xl",
                            "hover:!no-underline",
                            !feature.help_link && "pointer-events-none",
                        )}
                    >
                        <div
                            className={clsx(
                                "text-base",
                                "font-semibold",
                                "text-gray-900 dark:text-gray-0",
                            )}
                        >
                            {feature.title}
                        </div>
                        <div
                            className={clsx(
                                "text-sm",
                                "dark:text-gray-400 text-gray-600",
                            )}
                        >
                            {formatDescription(feature.description)}
                        </div>
                    </a>
                ))}
            </div>
        </div>
    );
};
