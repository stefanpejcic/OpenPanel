import React, { useState, useEffect } from 'react';
import clsx from 'clsx';
import { OrangeStarIcon } from './icons/orange-star';

export const LandingHeroGithubStars = () => {
    const [version, setVersion] = useState('1.7.0');

    useEffect(() => {
        const fetchVersion = async () => {
            try {
                const response = await fetch('https://usage-api.openpanel.org/latest_version');
                const data = await response.json();

                if (data && data.latest_version) {
                    setVersion(data.latest_version);
                }
            } catch (error) {
                console.error('Failed to fetch version:', error);
            }
        };

        fetchVersion();
    }, []);

    return (
        <a
            href={`/docs/changelog/${version}`}
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
                <OrangeStarIcon className="drop-shadow-none dark:drop-shadow-github-stars-glow" />
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
                    <span className="font-semibold">
                        latest version <span>{version}</span>
                    </span>{" "}
                    <span>- view the changelog</span>
                </span>
            </div>
        </a>
    );
};
