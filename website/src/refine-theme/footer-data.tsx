import React from "react";
import { FooterDiscordIcon } from "./icons/footer-discord";
import { FooterGithubIcon } from "./icons/footer-github";
import { FooterTwitterIcon } from "./icons/footer-twitter";
import { FooterLinkedinIcon } from "./icons/footer-linkedin";
import { FooterRedditIcon } from "./icons/footer-reddit";
import { NewBadgeIcon } from "./icons/popover";

export const menuItems = [
    {
        label: "Resources",
        items: [
            {
                label: "End-user Docs",
                href: "/docs/panel/intro/",
            },
            {
                label: "Admin Docs",
                href: "/docs/admin/intro/",
            },
            {
                label: "How-to Guides",
                href: "/docs/articles/intro/",
            },
            {
                label: "OpenCLI",
                href: "https://dev.openpanel.com/cli/",
            },            
            {
                label: "API Docs",
                href: "https://dev.openpanel.com/openadmin-api/",
            },
            {
                label: "Install Command",
                href: "/install",
            },
        ],
    },
    {
        label: "Product",
        items: [
            {
                label: "Features",
                href: "/features",
            },
            {
                label: "Live Preview",
                href: "/demo",
            },
            {
                label: "Roadmap",
                href: "/roadmap",
            },
            {
                label: "Changelog",
                href: "/docs/changelog/intro/",
            },
        ],
    },
    {
        label: "About",
        items: [
            {
                label: "Blog",
                href: "/blog",
            },
            {
                label: "Statistics",
                icon: <NewBadgeIcon />,
                href: "/statistics",
            },            
            {
                label: "Contact",
                href: "mailto:info@openpanel.com",
            },
        ],
    },
];

export const secondaryMenuItems = [
    {
        label: "Privacy Policy",
        href: "/privacy-policy",
    },
];

export const footerDescription = `OpenPanel - web hosting panel.`;

export const socialLinks = [
    {
        icon: FooterGithubIcon,
        href: "https://github.com/stefanpejcic/openpanel",
    },
    {
        icon: FooterDiscordIcon,
        href: "https://discord.com/invite/7bNY8fANqF",
    },
    {
        icon: FooterRedditIcon,
        href: "https://www.reddit.com/r/openpanelco/",
    },
    {
        icon: FooterTwitterIcon,
        href: "https://x.com/openpanel",
    },
    {
        icon: FooterLinkedinIcon,
        href: "https://www.linkedin.com/company/openpanel/",
    },
];
