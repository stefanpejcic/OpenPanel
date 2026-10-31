import React from "react";
import clsx from "clsx";

export const EnhanceIntroContent = ({
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
                Why hosting providers choose OpenPanel as their Enhance
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
                    Enhance bills every hosted website, starting at $0.15 per
                    website per month with a $10 minimum, so the bill grows
                    with every customer you sign up. A server with 1,000
                    websites costs $150/month on Enhance. OpenPanel is an{" "}
                    <span
                        className={clsx(
                            "font-semibold",
                            "dark:text-gray-0 text-gray-900",
                        )}
                    >
                        Enhance alternative
                    </span>{" "}
                    with one fixed price per server - €14.95/month, no matter
                    how many accounts and websites it hosts.
                </p>
                <p>
                    Both panels isolate websites in containers and manage
                    several servers from one place. OpenPanel goes further and
                    gives every user their own web server, PHP and database
                    service, lets users run their own containers, and runs on
                    Ubuntu, Debian, AlmaLinux, Rocky Linux and CentOS on x86_64
                    or ARM, where Enhance runs on Ubuntu only.
                </p>
                <p>
                    OpenPanel also comes with a forever free Community Edition,
                    so you can try the panel on your own servers before you pay
                    anything, and the Enterprise price stays locked for as long
                    as you keep the license.
                </p>
            </div>
        </div>
    );
};
