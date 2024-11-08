import clsx from "clsx";
import React from "react";
export const LandingGithubStarButton = () => {

    return (
        <a
            href="https://my.openpanel.com/clientarea.php"
            className={clsx(
                                "hover:!no-underline",
                            )}
                        >
                    <span className={clsx("text-base", "font-semibold")}>
                        Account Login
                    </span>
                </a>
    );
};
