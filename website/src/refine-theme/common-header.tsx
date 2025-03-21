import Link from "@docusaurus/Link";
import clsx from "clsx";
import { motion, useScroll, useTransform } from "framer-motion";
import React, { useState } from "react";

import { HamburgerIcon } from "./icons/hamburger";
import { OpenPanelLogoIcon } from "./icons/openpanel-logo";

import { openFigma } from "../utils/open-figma";
import { Menu } from "./common-header/menu";
import { MobileMenuModal } from "./common-header/mobile-menu-modal";
import { CommonThemeToggle } from "./common-theme-toggle";
import { TopAnnouncement } from "./top-announcement";

import { LandingGithubStarButton } from "./landing-github-star-button";

type Props = {
    hasSticky?: boolean;
    trackProgress?: boolean;
};

export const CommonHeader = ({ trackProgress }: Props) => {
    const [isModalOpen, setIsModalOpen] = useState(false);
    const { scrollYProgress } = useScroll();

    const progressPercentage = useTransform(
        scrollYProgress,
        [0.03, 0.95],
        ["0%", "100%"],
    );

    return (
        <>
            
            <header className={clsx("sticky", "top-0", "z-10")}>
                <div
                    className={clsx(
                        "absolute",
                        "top-0 left-0 right-0",
                        "backdrop-blur-[6px]",
                        "landing-md:backdrop-blur-[12px]",
                        "z-[-1]",
                        "bg-gray-0 dark:bg-gray-900",
                        "bg-opacity-80 dark:bg-opacity-80",
                        "pointer-events-none",
                    )}
                    style={{
                        top: "-20px",
                        bottom: "-80px",
                        WebkitMaskImage:
                            "linear-gradient(to bottom, transparent 0, transparent calc(0% + 20px), black calc(0% + 20px), black calc(100% - 80px), transparent calc(100% - 80px))",
                        maskImage:
                            "linear-gradient(to bottom, transparent 0, transparent calc(0% + 20px), black calc(0% + 20px), black calc(100% - 80px), transparent calc(100% - 80px))",
                    }}
                />
                <div
                    className={clsx(
                        "relative",
                        "z-[1]",
                        "p-4",
                        "landing-sm:px-8",
                        "landing-md:py-5",
                    )}
                >
                    <div
                        className={clsx(
                            "mx-auto",
                            "flex",
                            "items-center",
                            "justify-between",
                            "max-w-[896px]",
                            "landing-lg:max-w-[1200px]",
                        )}
                    >
                        <div
                            className={clsx(
                                "w-auto",
                                "landing-lg:w-[200px]",
                            )}
                        >
                            <Link to="/" class="no-underline flex items-center gap-2" onContextMenu={openFigma}>
                                <OpenPanelLogoIcon className="text-gray-900 dark:text-gray-0" />
                            
                <span
                    className={clsx(
                        "text-[22px]",
                        "text-gray-1000 dark:text-gray-0",
                        "text-base",
                        "font-bold",
                        "pr-5",
                    )}
                >
                    OpenPanel
                </span>


</Link>
                        </div>
                        <button
                            type="button"
                            className={clsx(
                                "text-gray-900 dark:text-gray-0",
                                "block landing-md:hidden",
                            )}
                            onClick={() => setIsModalOpen(true)}
                        >
                            <HamburgerIcon />
                        </button>
                        <div
                            className={clsx(
                                "hidden landing-md:flex",
                                "flex-1",
                                "items-center",
                                "gap-8",
                            )}
                        >
                            <Menu />
                        </div>
                        <div
                            className={clsx(
                                "hidden landing-md:flex",
                                "items-center",
                                "justify-end",
                                "gap-4",
                                "w-[130px]",
                                "landing-lg:w-[200px]",
                            )}
                        >
                            <LandingGithubStarButton />
                            <CommonThemeToggle />
                        </div>
                        <MobileMenuModal
                            isModalOpen={isModalOpen}
                            setIsModalOpen={setIsModalOpen}
                        />
                    </div>
                </div>
                {trackProgress && (
                    <div className={clsx("w-full", "h-0.5", "translate")}>
                        <motion.div
                            className={clsx("h-full", "bg-refine-blue")}
                            style={{ width: progressPercentage }}
                        />
                    </div>
                )}
            </header>
        </>
    );
};
