import clsx from "clsx";
import React, { FC } from "react";
import {
    HostkeyIcon,
    AltusHostIcon,
    DigitalOceanIcon,
    UnlimitedIcon,
} from "../components/landing/icons";

type Props = {
    className?: string;
};

export const LandingTrustedByDevelopers: FC<Props> = ({ className }) => {
    return (
        <div className={clsx(className, "w-full")}>
            <div
                className={clsx(
                    "not-prose",
                    "relative",
                    "w-full",
                    "p-4 landing-md:p-10",
                    "dark:bg-landing-trusted-by-developers-dark bg-landing-trusted-by-developers",
                    "dark:bg-gray-800 bg-gray-50",
                    "rounded-2xl landing-sm:rounded-3xl",
                )}
            >
                <p
                    className={clsx(
                        "whitespace-nowrap",
                        "px-0 landing-sm:px-6 landing-lg:px-0",
                        "text-base landing-sm:text-2xl",
                        "dark:text-gray-400 text-gray-600",
                    )}
                >
                    Trusted by major hosting brands
                </p>
                <div
                    className={clsx(
                        "grid",
                        "grid-cols-3 landing-lg:grid-cols-6",
                        "min-h-[160px] landing-lg:min-h-[80px]",
                        "justify-center",
                        "items-center",
                        "mt-6",
                    )}
                >
                    {list.slice(0, 6).map((item) => (
                        <div
                            key={item.id}
                            className={clsx(
                                "max-w-[187px] w-full",
                                "overflow-hidden",
                            )}
                        >
                            <a
                                href={item.href}
                                target="_blank"
                                rel="noopener noreferrer"
                                className={clsx(
                                    "animate-opacity-reveal",
                                    "flex",
                                    "items-center",
                                    "justify-center",
                                    "max-w-[187px]",
                                )}
                            >
                                {item.icon}
                            </a>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};

type IList = {
    icon: React.ReactNode;
    id: number;
    href: string;
}[];

const list: IList = [
    { icon: <HostkeyIcon />, id: 1, href: "https://hostkey.com/apps/hosting-control-panels/openpanel/" },
    { icon: <AltusHostIcon />, id: 2, href: "https://altushost.com" },
    { icon: <DigitalOceanIcon />, id: 3, href: "https://digitalocean.com/" },
    { icon: <UnlimitedIcon />, id: 4, href: "https://unlimited.rs/" },
];
