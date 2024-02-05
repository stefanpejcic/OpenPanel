import React from "react";
import { FooterDiscordIcon } from "./icons/footer-discord";
import { FooterGithubIcon } from "./icons/footer-github";
import { FooterLinkedinIcon } from "./icons/footer-linkedin";
import { FooterRedditIcon } from "./icons/footer-reddit";
import { FooterTwitterIcon } from "./icons/footer-twitter";
import { NewBadgeIcon } from "./icons/popover";

export const menuItems = [
    {
        label: "Resources",
        items: [
            {
                label: "Getting Started",
                href: "/docs/admin/intro/",
            },
            {
                label: "Tutorials",
                href: "/docs/tutorial/introduction/index/",
            },
            {
                label: "Blog",
                href: "/blog",
            },
        ],
    },
    {
        label: "Product",
        items: [
            {
                label: "BETA",
                icon: <NewBadgeIcon />,
                href: "/beta",
            },
            {
                label: "Products",
                href: "/templates",
            },
            {
                label: "Technology Stack",
                href: "/integrations",
            },
        ],
    },
    {
        label: "Company",
        items: [
            {
                label: "About",
                href: "/about",
            },
            {
                label: "Become a Partner",
                href: "/partner",
            },
            {
                label: "Contact Us",
                href: "#",
            },
        ],
    },
];

export const secondaryMenuItems = [
     {
         label: "Terms & Conditions",
         href: "/terms",
     },
    {
        label: "Privacy Policy",
        href: "/privacy-policy",
    },
    {
        label: "License",
        href: "/LICENSE",
    },
];

export const footerDescription = `OpenPanel - web hosting panel.`;

export const socialLinks = [
    {
        icon: FooterGithubIcon,
        href: "https://github.com/stefanpejcic",
    },
    {
        icon: FooterDiscordIcon,
        href: "https://discord.gg/",
    },
    {
        icon: FooterRedditIcon,
        href: "https://www.reddit.com/r/open-panel/",
    },
    {
        icon: FooterTwitterIcon,
        href: "https://twitter.com/",
    },
    {
        icon: FooterLinkedinIcon,
        href: "https://www.linkedin.com/company/openpanel/",
    },
];
