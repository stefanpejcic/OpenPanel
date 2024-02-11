import clsx from "clsx";
import React from "react";
export const LandingGithubStarButton = () => {

    return (
        <a
            href="/demo"
            rel="noreferrer"
            className={clsx(
                "flex gap-2 items-center",
                "font-normal",
                "text-sm leading-6",
                "text-gray-500 dark:text-gray-400",
                "hover:text-gray-400 dark:hover:text-gray-300",
                "hover:no-underline",
                "transition-colors",
                "duration-200",
                "ease-in-out",
            )}
        >

            <div className={clsx("flex items-center", "w-10 h-6")}>
                    <span>Demo</span>
            </div>
        </a>
    );
};
