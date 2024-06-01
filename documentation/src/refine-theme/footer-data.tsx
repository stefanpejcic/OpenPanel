import React from "react";
import { FooterDiscordIcon } from "./icons/footer-discord";
import { FooterGithubIcon } from "./icons/footer-github";
import { FooterLinkedinIcon } from "./icons/footer-linkedin";
import { FooterRedditIcon } from "./icons/footer-reddit";
import { NewBadgeIcon } from "./icons/popover";

export const menuItems = [
    {
        label: "Resources",
        items: [
            {
                label: "Documentation",
                href: "/docs/admin/intro/",
            },
            {
                label: "Support Forums",
                href: "https://community.openpanel.co/",
            },
            {
                label: "License Verification",
                href: "/verify",
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
        label: "Product",
        items: [
//            {
//                label: "Enterprise edition",
//                icon: <NewBadgeIcon />,
//                href: "/beta",
//            },
            //{
            //    label: "Products",
            //    href: "/templates",
            //},
            {
                label: "Features",
                href: "/features",
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
                label: "Blog",
                href: "/blog",
            },
            {
                label: "Become a Partner",
                href: "mailto:info@openpanel.co",
            },
            {
                label: "Contact Us",
                href: "mailto:info@openpanel.co",
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
        icon: FooterLinkedinIcon,
        href: "https://www.linkedin.com/company/openpanel/",
    },
];
