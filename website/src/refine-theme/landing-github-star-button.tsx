import clsx from "clsx";
import React from "react";
export const LandingGithubStarButton = () => {

    return (
        <a
            href="/demo"
            className={clsx(
                                "hover:!no-underline",
                            )}
                        >
                    <span className={clsx("text-base", "font-semibold")}>
                        Try Demo
                    </span>
                </a>
    );
};
