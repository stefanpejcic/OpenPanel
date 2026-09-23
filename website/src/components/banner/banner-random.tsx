import React from "react";
import { BannerImageWithText } from "./banner-image-with-text";
import { BannerExamples } from "./banner-examples";
import BrowserOnly from "@docusaurus/BrowserOnly";

const data = [
    {
        title: "Give every customer their own VPS, minus the VPS bill",
        description:
            "OpenPanel Enterprise isolates each user in their own container with CPU, RAM and disk limits, so one noisy site never takes down the rest. One flat price per server, unlimited accounts, from €12.46/mo.",
        image: {
            src: "/img/openadmin-screenshots/001_dashboard-window_dark.png",
            alt: "OpenAdmin dashboard with an OpenPanel Enterprise license",
            href: "https://openpanel.com/enterprise/?ref=banner-admin-panel",
        },
        button: {
            text: "Start 30-day free trial",
            href: "https://openpanel.com/trial?ref=banner-admin-panel",
        },
        bannerName: "banner-twitter",
    },
    /*    {
        title: "Save developer hours!",
        description:
            "An open-source, industry-standard codebase designed for building enterprise-grade internal tools, admin panels, and CRUD apps.",
        image: {
            src: "/banners/banner-save-hours.png",
            alt: "Illustration about time is gold",
            href: "https://github.com/refinedev/refine?ref=banner-save-hours",
        },
        button: {
            text: "Learn more",
            href: "https://github.com/refinedev/refine?ref=banner-save-hours",
        },
        bannerName: "banner-save-hours",
    },
    {
        description:
            "refine is ranked among the top 3 rapidly growing React frameworks in the ecosystem.",
        image: {
            src: "/banners/banner-oss-insight.png",
            alt: "Photo about refine ranking on OSS Insight website",
            href: "https://github.com/refinedev/refine?ref=banner-oss-insight",
        },
        button: {
            text: "Learn more",
            href: "https://github.com/refinedev/refine?ref=banner-oss-insight",
        },
        bannerName: "banner-oss-insight",
    }, */
];

// +1 for BannerExamples
const random = Math.floor(Math.random() * (data.length + 1));

export const BannerRandom = () => {
    // when random is equal to data.length, we will show BannerExamples
    if (random === data.length) {
        return <BrowserOnly>{() => <BannerExamples />}</BrowserOnly>;
    }

    return (
        <BrowserOnly>
            {() => <BannerImageWithText {...data[random]} />}
        </BrowserOnly>
    );
};
