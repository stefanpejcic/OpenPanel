import clsx from "clsx";
import React, { FC } from "react";
import { EnterpriseGetInTouchButton } from "./enterprise-get-in-touch-button";

type Props = {
    className?: string;
    question?: string;
    subtext?: string;
    buttonLabel?: string;
    buttonHref?: string;
    eventName?: string;
};

export const EnterpriseGetInTouchCta: FC<Props> = (props) => {
    return (
        <div className={clsx(props.className)}>
            <div
                className={clsx(
                    "not-prose",
                    "flex flex-col landing-md:flex-row",
                    "items-center",
                    "justify-between",
                    "gap-4 landing-sm:gap-6",
                    "py-6 pr-6 pl-6 landing-md:pl-12",
                    "rounded-2xl landing-md:rounded-full",
                    "dark:bg-gray-800 bg-gray-50",
                    "dark:bg-enterprise-cta-dark dark:landing-md:bg-enterprise-cta-dark-md",
                    "bg-enterprise-cta-light landing-md:bg-enterprise-cta-light-md",
                )}
            >
                <div className={clsx("flex flex-col", "gap-1")}>
                    <h2
                        className={clsx(
                            "text-sm landing-sm:text-2xl",
                            "dark:text-gray-400 text-gray-600",
                        )}
                    >
                        {props.question ?? "Ready to try OpenPanel Enterprise?"}
                    </h2>
                    {props.subtext && (
                        <p
                            className={clsx(
                                "text-xs landing-sm:text-sm",
                                "dark:text-gray-500 text-gray-500",
                            )}
                        >
                            {props.subtext}
                        </p>
                    )}
                </div>
                <EnterpriseGetInTouchButton
                    label={props.buttonLabel}
                    href={props.buttonHref}
                    eventName={props.eventName}
                />
            </div>
        </div>
    );
};
