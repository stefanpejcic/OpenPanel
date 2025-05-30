import React, { useState, useEffect } from 'react';
import clsx from 'clsx';
import { OrangeStarIcon } from './icons/orange-star';

export const LandingHeroGithubStars = () => {
    const [version, setVersion] = useState('1.2.9'); // Default version

    useEffect(() => {
        // Function to fetch the latest version tag from Docker Hub
        const fetchVersion = async () => {
            try {
                const response = await fetch('https://hub.docker.com/v2/repositories/openpanel/openpanel-ui/tags');
                const data = await response.json();
                
                // Extract tags, filter out 'latest', sort, and get the latest numeric version
                const tags = data.results.map(tag => tag.name);
                const filteredTags = tags
                    .filter(tag => /^\d+\.\d+\.\d+$/.test(tag))  // Only valid version tags
                    .sort((a, b) => {
                        // Compare versions numerically
                        const [majorA, minorA, patchA] = a.split('.').map(Number);
                        const [majorB, minorB, patchB] = b.split('.').map(Number);
                        
                        if (majorA !== majorB) return majorB - majorA;
                        if (minorA !== minorB) return minorB - minorA;
                        return patchB - patchA;
                    });

                // Get the latest version from the sorted list
                const latestVersion = filteredTags.length > 0 ? filteredTags[0] : '1.3.4';  // new verzija tag
                setVersion(latestVersion); // Update the version state
            } catch (error) {
                console.error('Failed to fetch version:', error);
            }
        };

        fetchVersion();
    }, []); // After initial render

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
            style={{
                transform: "translateZ(0)",
            }}
        >
            <div
                className={clsx(
                    "absolute",
                    "inset-0",
                    "overflow-hidden",
                    "rounded-3xl",
                    "",
                )}
                style={{
                    transform: "translateZ(0)",
                }}
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
                    <span className={clsx("font-semibold")}>
                        OpenPanel <span>{version}</span>
                    </span> {" "}
                    <span>is out</span>
                </span>
            </div>
        </a>
    );
};
