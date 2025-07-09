import React from "react";
import clsx from "clsx";

type Props = {
    href: string;
};

export const TwitterButton = ({ href }: Props) => {
    return (
        <a
            href="/support"
            className={clsx(
                "no-underline",
                "flex items-center justify-center gap-2",
                "rounded-lg",
                "w-[144px] h-[48px]",
                "border border-gray-200 dark:border-gray-700",
            )}
        >
            <span className={clsx("text-gray-700 dark:text-gray-500")}>
                Get Support
            </span>
        </a>
    );
};
