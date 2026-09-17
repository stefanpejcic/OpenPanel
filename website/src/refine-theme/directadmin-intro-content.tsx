import React from "react";
import clsx from "clsx";

export const DirectadminIntroContent = ({
    className,
}: {
    className?: string;
}) => {
    return (
        <div
            className={clsx(
                "flex flex-col",
                "gap-6 landing-sm:gap-8",
                "not-prose",
                className,
            )}
        >
            <h2
                className={clsx(
                    "text-2xl landing-sm:text-[32px] landing-sm:leading-[40px]",
                    "font-semibold",
                    "dark:text-gray-0 text-gray-900",
                )}
            >
                Why hosting providers choose OpenPanel as their DirectAdmin
                alternative
            </h2>
            <div
                className={clsx(
                    "flex flex-col",
                    "gap-4",
                    "text-base",
                    "dark:text-gray-400 text-gray-700",
                    "max-w-[792px]",
                )}
            >
                <p>
                    DirectAdmin licenses run $5-29/month per account, with
                    repeated undisclosed price hikes over the years. OpenPanel
                    is a{" "}
                    <span
                        className={clsx(
                            "font-semibold",
                            "dark:text-gray-0 text-gray-900",
                        )}
                    >
                        DirectAdmin alternative
                    </span>{" "}
                    built around one fixed price per server - €14.95/month,
                    for life - however many accounts you run.
                </p>
                <p>
                    DirectAdmin's user isolation is limited to an alpha
                    jailed shell or requires paid CloudLinux, and services
                    like PHP and MySQL are shared across accounts by default.
                    OpenPanel isolates every user and every service in its
                    own Podman container out of the box.
                </p>
                <p>
                    DirectAdmin exposes a REST API, but its MCP server for AI
                    agents is only a community, unofficial project. OpenPanel
                    ships a native MCP server so AI agents can manage
                    hosting through the same API humans use. There's no
                    automatic import tool on either side, so migrating
                    accounts over is a manual move.
                </p>
            </div>
        </div>
    );
};
