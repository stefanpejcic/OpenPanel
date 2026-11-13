import React from "react";
import Link from "@docusaurus/Link";
import clsx from "clsx";

import { MENU_ITEMS, NavbarItemType, NavbarPopoverItemType } from "./constants";
import {
    GithubStarIcon,
    GithubIcon,
    DiscordIcon,
    CompareIcon,
} from "../icons/popover";
import { MenuItem } from "./menu-item";
import { NavbarItem } from "./navbar-item";
import { NavbarPopoverItem } from "./navbar-popover-item";

const MailIcon = ({ className }: { className?: string }) => (
    <svg
        width="20"
        height="20"
        viewBox="0 0 20 20"
        fill="none"
        aria-hidden="true"
        className={className}
    >
        <rect x="2.5" y="4.5" width="15" height="11" rx="2" stroke="currentColor" strokeWidth="1.5" />
        <path d="m3.5 6 6.5 5 6.5-5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
);

const footerStrip = clsx(
    "border-t dark:border-gray-700 border-gray-300",
    "dark:bg-gray-800 bg-gray-100",
    "flex items-center",
    "py-4 px-7",
);

const PopoverFooter = ({ type }: { type: NavbarPopoverItemType["footer"] }) => {
    switch (type) {
        case "github":
            return (
                <Link
                    to="https://github.com/stefanpejcic/openpanel/"
                    className="no-underline"
                >
                    <div className={footerStrip}>
                        <GithubStarIcon />
                        <div className="ml-4 dark:text-gray-400 text-gray-600">
                            If you like OpenPanel, don’t forget to star us on
                            GitHub!
                        </div>
                    </div>
                </Link>
            );
        case "community":
            return (
                <div className={clsx(footerStrip, "justify-between")}>
                    <div className="dark:text-gray-400 text-gray-600">
                        Join the party!
                    </div>
                    <div className="flex gap-4">
                        <Link
                            to="https://github.com/stefanpejcic/openpanel/"
                            className="no-underline hover:text-inherit"
                        >
                            <GithubIcon className="dark:text-gray-400 text-gray-500" />
                        </Link>
                        <Link to="https://discord.openpanel.com/">
                            <DiscordIcon />
                        </Link>
                    </div>
                </div>
            );
        case "contact":
            return (
                <Link
                    to="mailto:info@openpanel.com"
                    className="no-underline hover:no-underline"
                >
                    <div className={clsx(footerStrip, "justify-between gap-4")}>
                        <div className="dark:text-gray-400 text-gray-600">
                            Not sure which fits? Talk to us.
                        </div>
                        <MailIcon className="shrink-0 dark:text-gray-400 text-gray-500" />
                    </div>
                </Link>
            );
        case "trial":
            return (
                <Link to="/trial" className="no-underline hover:no-underline">
                    <div className={clsx(footerStrip, "justify-between gap-4")}>
                        <div className="dark:text-gray-400 text-gray-600">
                            Try Enterprise free for 30 days, no credit card.
                        </div>
                        <span className="font-semibold dark:text-refine-cyan-alt text-refine-blue">
                            Start trial →
                        </span>
                    </div>
                </Link>
            );
        default:
            return null;
    }
};

const QuickLinks = ({
    quickLinks,
}: {
    quickLinks: NonNullable<NavbarPopoverItemType["quickLinks"]>;
}) => (
    <div className="flex items-start p-4">
        <div className="shrink-0">
            <CompareIcon />
        </div>
        <div className="ml-2">
            <div className="font-semibold text-gray-900 dark:text-white">
                {quickLinks.title}
            </div>
            <div className="mt-1 flex flex-wrap gap-1.5">
                {quickLinks.links.map((l) => (
                    <Link
                        key={l.to}
                        to={l.to}
                        className={clsx(
                            "no-underline hover:no-underline",
                            "text-xs",
                            "rounded-full px-2 py-0.5",
                            "border dark:border-gray-700 border-gray-200",
                            "text-gray-600 dark:text-gray-300",
                            "hover:bg-gray-50 dark:hover:bg-gray-800",
                        )}
                    >
                        {l.label}
                    </Link>
                ))}
            </div>
        </div>
    </div>
);

export const Menu: React.FC = () => {
    return (
        <>
            {MENU_ITEMS.map((item) => {
                if (item.isPopover) {
                    const twoCols = item.columns === 2;
                    return (
                        <NavbarPopoverItem
                            key={`navbar-${item.label}`}
                            item={item}
                        >
                            <div
                                className={clsx(
                                    "grid gap-4",
                                    twoCols ? "grid-cols-2 w-[672px]" : "w-[336px]",
                                    "p-4",
                                    "dark:bg-gray-900 bg-white",
                                )}
                            >
                                {item.items.map((subItem) => (
                                    <MenuItem
                                        key={subItem.label}
                                        item={subItem}
                                    />
                                ))}
                                {item.quickLinks && (
                                    <QuickLinks quickLinks={item.quickLinks} />
                                )}
                            </div>
                            <PopoverFooter type={item.footer} />
                        </NavbarPopoverItem>
                    );
                }

                return (
                    <NavbarItem
                        key={`navbar-${item.label}`}
                        item={item as NavbarItemType}
                    />
                );
            })}
        </>
    );
};
