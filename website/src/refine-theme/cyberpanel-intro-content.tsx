import React from "react";
import clsx from "clsx";

export const CyberpanelIntroContent = ({
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
                Why hosting providers choose OpenPanel as their CyberPanel
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
                    CyberPanel's core is free, but its Enterprise tier and
                    OpenLiteSpeed licensing can run up to $97/month per
                    account, with CloudLinux as an extra cost on top for
                    isolation. OpenPanel is a{" "}
                    <span
                        className={clsx(
                            "font-semibold",
                            "dark:text-gray-0 text-gray-900",
                        )}
                    >
                        CyberPanel alternative
                    </span>{" "}
                    built around one fixed price per server - €14.95/month,
                    license and isolation included.
                </p>
                <p>
                    CyberPanel runs without LVE isolation by default and
                    shares PHP and MySQL across accounts unless you add
                    CloudLinux, and it doesn't support ARM (AArch64) servers
                    at all. OpenPanel isolates every user and every service
                    in its own Podman container out of the box, on x86_64 or
                    ARM.
                </p>
                <p>
                    CyberPanel has no automatic import tool, so switching
                    means migrating accounts by hand. OpenPanel supports the
                    same Ubuntu, AlmaLinux, and Rocky Linux distributions
                    CyberPanel runs on, so rebuilding your hosting
                    environment doesn't mean changing your OS.
                </p>
            </div>
        </div>
    );
};
