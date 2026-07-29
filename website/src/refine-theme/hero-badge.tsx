import React from "react";
import clsx from "clsx";

type Props = {
    href: string;
    icon: React.ReactNode;
    children: React.ReactNode;
    className?: string;
};

export const HeroBadge = ({ href, icon, children, className }: Props) => {
    return (
        <a
            href={href}
            target="_blank"
            rel="noopener noreferrer"
            className={clsx(
                "self-start",
                "relative",
                "rounded-3xl",
                "p-px",
                "hover:no-underline",
                "w-auto",
                "bg-gray-200 dark:bg-gray-700",
                className,
            )}
            style={{ transform: "translateZ(0)" }}
        >
            <div
                className={clsx(
                    "absolute",
                    "inset-0",
                    "overflow-hidden",
                    "rounded-3xl",
                )}
                style={{ transform: "translateZ(0)" }}
            >
                <div
                    className={clsx(
                        "hidden dark:block",
                        "absolute",
                        "-top-8",
                        "-left-8",
                        "animate-github-stars-border",
                        "w-24",
                        "h-24",
                        "rounded-full",
                        "bg-refine-orange",
                        "opacity-40",
                        "blur-xl",
                    )}
                />
            </div>
            <div
                className={clsx(
                    "hidden dark:block",
                    "absolute",
                    "-left-3",
                    "-top-3",
                    "z-[0]",
                    "w-12",
                    "h-12",
                    "blur-lg",
                    "bg-refine-orange",
                    "rounded-full",
                    "opacity-[0.15]",
                    "dark:animate-github-stars-glow",
                )}
            />
            <div
                className={clsx(
                    "relative",
                    "z-[1]",
                    "rounded-[23px]",
                    "py-[7px]",
                    "pl-2",
                    "pr-4",
                    "flex",
                    "gap-2",
                    "items-center",
                    "justify-center",
                    "bg-gray-50 dark:bg-gray-900",
                    "dark:bg-landing-hero-github-stars-gradient",
                )}
            >
                {icon}
                <span
                    className={clsx(
                        "font-normal",
                        "text-xs",
                        "text-transparent",
                        "bg-clip-text",
                        "bg-landing-hero-github-stars-text-light",
                        "dark:bg-landing-hero-github-stars-text-dark",
                    )}
                >
                    {children}
                </span>
            </div>
        </a>
    );
};
