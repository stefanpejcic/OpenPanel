import Link from "@docusaurus/Link";
import clsx from "clsx";
import React from "react";

import { HeaderGithubIcon } from "../icons/header-github";

type GitHubStarProps = {
    isPermanentDark?: boolean;
};

export const GitHubStar: React.FC<GitHubStarProps> = ({ isPermanentDark }) => {

    return (
        <Link
            className="flex items-center no-underline"
            to="https://community.openpanel.org/"
        >
            <HeaderGithubIcon
                className={clsx(
                    "text-gray-500 dark:gray-400",
                    isPermanentDark && "!text-white",
                )}
            />
            <div
                className={clsx(
                    "flex items-center",
                    "w-10 h-6",
                    "ml-2",
                    "text-sm font-medium ",
                    "text-gray-500 dark:text-gray-400",
                    isPermanentDark && "!text-white",
                )}
            >
            <span className="tabular-nums">Get Support</span>
            </div>
        </Link>
    );
};
