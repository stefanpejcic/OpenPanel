import clsx from "clsx";
import React from "react";

type Props = {
    className?: string;
};

export const CommonGithubStarButton = ({ className }: Props) => {

    return (
        <a
            href="https://community.openpanel.co/"
            target="_blank"
            rel="noreferrer"
            className={clsx(
                "text-sm",
                "text-gray-500 dark:text-gray-400",
                "rounded-[32px]",
                "border border-solid border-gray-300 dark:border-gray-700",
                "flex gap-2 items-center",
                "py-2 pl-2.5 pr-4",
                "no-underline",
                className,
            )}
        >
	<svg xmlns="http://www.w3.org/2000/svg" class="icon icon-tabler icon-tabler-message-2-heart" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M8 9h8" /><path d="M8 13h3.5" /><path d="M10.5 19.5l-1.5 -1.5h-3a3 3 0 0 1 -3 -3v-8a3 3 0 0 1 3 -3h12a3 3 0 0 1 3 3v4" /><path d="M18 22l3.35 -3.284a2.143 2.143 0 0 0 .005 -3.071a2.242 2.242 0 0 0 -3.129 -.006l-.224 .22l-.223 -.22a2.242 2.242 0 0 0 -3.128 -.006a2.143 2.143 0 0 0 -.006 3.071l3.355 3.296z" /></svg>
            <div className={clsx("flex items-center", "min-w-[76px] h-6")}>
                    <span
                        className={clsx(
                            "tabular-nums text-gray-800 dark:text-gray-100",
                        )}
                    >
			Get Support
                    </span>
            </div>
        </a>
    );
};
