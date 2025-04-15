import clsx from "clsx";
import React from "react";
export const LandingGithubStarButton = () => {

    return (
        <a
            href="https://my.openpanel.com/index.php?rp=/store/openpanel/enterprise-license"
            target="_blank"
            className={clsx(
                                "hover:!no-underline",
                            )}
                        >
                    <span className={clsx("text-base", "font-semibold")}>
                        Enterprise
                    </span>
                </a>
    );
};
