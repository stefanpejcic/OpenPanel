import Link from "@docusaurus/Link";
import { ArrowRightIcon } from "@site/src/refine-theme/icons/arrow-right";
import { LandingRainbowButton } from "@site/src/refine-theme/landing-rainbow-button";
import clsx from "clsx";
import React from "react";

const text = "Struggling with hosting panels?";
const description =
    "Boost your hosting service: Try OpenPanel Community edition and elevate your clients experience to new heights!";
const image =
    "https://openpanel.com/img/panel/v1/dashboard/dashboard.png";

export const BannerSidebar = ({ shouldShowBanner }) => {
    React.useEffect(() => {
        if (
            typeof window !== "undefined" &&
            typeof window.gtag !== "undefined" &&
            shouldShowBanner
        ) {
            window.gtag("event", "view_banner", {
                banner_name: "banner-sidebar",
                banner_text: text,
                banner_description: description,
                banner_image: image,
            });
        }
    }, [shouldShowBanner]);

    return (
        <div
            className={clsx(
                "flex",
                "flex-col",
                "gap-6",
                "py-6",
                "px-4",
                "rounded-2xl",
                "bg-banner-examples-sider-purple",
                "not-prose",
            )}
        >
            <Link
                to={"https://openpanel.com/demo?ref=banner-sidebar"}
                target="_blank"
                rel="noopener noreferrer"
                className={clsx(
                    "flex",
                    "w-full h-auto xl:h-[152px]",
                    "flex-shrink-0",
                    "rounded-md",
                    "overflow-hidden",
                )}
            >
                <img src={image} alt={"refine App screenshot"} loading="lazy" />
            </Link>

            <h2 className={clsx("text-2xl font-semibold", "text-gray-0")}>
                {text}
            </h2>
            <p className={clsx("text-base", "text-gray-100")}>{description}</p>

            <LandingRainbowButton
                className={clsx("w-max")}
                buttonClassname={clsx("!px-4", "!py-2")}
                href={"https://openpanel.com/demo?ref=banner-sidebar"}
                target="_blank"
                rel="noopener noreferrer"
            >
                <div
                    className={clsx("text-gray-900", "text-base", "font-bold")}
                >
                    Try online
                </div>
                <ArrowRightIcon className={clsx("ml-2", "w-4", "h-4")} />
            </LandingRainbowButton>
        </div>
    );
};
