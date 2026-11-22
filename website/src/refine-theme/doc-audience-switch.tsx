import Link from "@docusaurus/Link";
import { useLocation } from "@docusaurus/router";
import clsx from "clsx";
import React from "react";

// User / Admin / CLI switch on the panel, admin and opencli docs only, like Virtualizor's End User / Admin toggle
const SECTIONS = [
    { key: "panel", label: "User", to: "/docs/panel/intro/" },
    { key: "admin", label: "Admin", to: "/docs/admin/intro/" },
    { key: "opencli", label: "CLI", to: "/docs/opencli/intro/" },
];

export const useDocAudience = (): string | undefined => {
    const { pathname } = useLocation();
    return SECTIONS.find((s) => pathname.startsWith(`/docs/${s.key}/`))?.key;
};

export const DocAudienceSwitch = ({ className }: { className?: string }) => {
    const current = useDocAudience();
    if (!current) {
        return null;
    }

    return (
        <div
            role="tablist"
            aria-label="Documentation for"
            className={clsx(
                "grid grid-cols-3",
                "p-1",
                "rounded-full",
                "border border-gray-300 dark:border-gray-700",
                "text-sm font-semibold",
                className,
            )}
        >
            {SECTIONS.map((s) => {
                const active = s.key === current;
                return (
                    <Link
                        key={s.key}
                        to={s.to}
                        role="tab"
                        aria-selected={active}
                        className={clsx(
                            "text-center",
                            "py-1.5 px-2",
                            "whitespace-nowrap",
                            "rounded-full",
                            "no-underline hover:no-underline",
                            "transition-colors duration-200 ease-in-out",
                            // same active look as the TOC item next to it
                            active &&
                                "bg-refine-blue-2-light dark:bg-refine-blue-2 dark:bg-opacity-10 text-refine-react-light-link dark:text-refine-react-dark-link",
                            !active &&
                                "text-refine-600 dark:text-refine-300 hover:bg-gray-100 dark:hover:bg-gray-700",
                        )}
                    >
                        {s.label}
                    </Link>
                );
            })}
        </div>
    );
};
