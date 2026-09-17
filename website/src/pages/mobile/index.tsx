import React from "react";
import Head from "@docusaurus/Head";
import Link from "@docusaurus/Link";
import clsx from "clsx";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { OpenPanelLogoIcon } from "@site/src/refine-theme/icons/openpanel-logo";
import { MobileFaq } from "@site/src/refine-theme/mobile-faq";

const DOWNLOAD_URL =
    "https://github.com/stefanpejcic/openpanel-mobile-app/releases/latest/download/OpenPanel.apk";
const REPO_URL = "https://github.com/stefanpejcic/openpanel-mobile-app";

const screenshots = [
    "/img/mobile/screenshot-empty-list.jpg",
    "/img/mobile/screenshot-add-server.jpg",
    "/img/mobile/screenshot-server-list.jpg",
    "/img/mobile/screenshot-dashboard.jpg",
];

const DownloadIcon = () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
        <path
            d="M12 3v12m0 0-4-4m4 4 4-4M5 21h14"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
        />
    </svg>
);

const features = [
    {
        title: "Multiple servers, one app",
        description:
            "Add every OpenPanel account and OpenAdmin server you manage, and switch between them instantly.",
    },
    {
        title: "OpenPanel & OpenAdmin, both supported",
        description:
            "One app for both products - pick the type when you add a server, or let the app detect it from the port.",
    },
    {
        title: "Credentials stay on your device",
        description:
            "Logins are stored in your phone's secure keychain, not on a third-party server.",
    },
    {
        title: "See what's reachable at a glance",
        description:
            "Each saved server shows a live green or red status dot, so you know what needs attention before you open it.",
    },
];

const steps = [
    "Download the APK above on your phone.",
    'Open the downloaded file - Android will warn it\'s from outside the Play Store the first time, tap "Settings" and allow installs from this source.',
    'Open the app, tap "+ Add server", and enter your panel\'s URL, username and password.',
];

const Mobile: React.FC = () => {
    const title = "OpenPanel Mobile App | Manage your servers from your phone";
    const description =
        "Download the free OpenPanel mobile app for Android to manage your OpenPanel accounts and OpenAdmin servers on the go.";

    return (
        <CommonLayout description={description}>
            <Head>
                <title>{title}</title>
                <meta property="og:title" content={title} />
                <html data-page="mobile" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div
                    className={clsx(
                        sectionWidth,
                        "flex flex-col",
                        "gap-16 landing-sm:gap-20 landing-md:gap-28",
                        "px-2 landing-sm:px-0",
                        "pt-8 landing-sm:pt-12 landing-lg:pt-20",
                        "pb-16 landing-sm:pb-24 landing-md:pb-32",
                        "mx-auto",
                        "not-prose",
                    )}
                >
                    {/* Hero */}
                    <div
                        className={clsx(
                            "flex flex-col items-center text-center",
                            "gap-6",
                            "px-4 landing-md:px-10",
                        )}
                    >
                        <OpenPanelLogoIcon
                            className={clsx(
                                "w-16 h-16 landing-sm:w-20 landing-sm:h-20",
                                "text-gray-900 dark:text-gray-0",
                            )}
                        />
                        <h1
                            className={clsx(
                                "text-3xl landing-sm:text-[40px] landing-sm:leading-[48px]",
                                "font-semibold",
                                "tracking-tight",
                                "p-0",
                                "dark:text-gray-0 text-gray-900",
                            )}
                        >
                            OpenPanel, in your pocket.
                        </h1>
                        <p
                            className={clsx(
                                "max-w-xl",
                                "text-base landing-sm:text-lg",
                                "dark:text-gray-400 text-gray-600",
                            )}
                        >
                            Manage your OpenPanel accounts and OpenAdmin servers from your
                            phone - add every server you run, check their status, and log
                            in without typing a URL each time.
                        </p>
                        <div
                            className={clsx(
                                "flex flex-col landing-xs:flex-row",
                                "items-center",
                                "gap-4",
                                "mt-2",
                            )}
                        >
                            <Link
                                to={DOWNLOAD_URL}
                                className={clsx(
                                    "!text-gray-0 dark:!text-gray-900",
                                    "bg-refine-blue dark:bg-refine-cyan-alt",
                                    "transition-[filter]",
                                    "duration-150",
                                    "ease-in-out",
                                    "hover:brightness-110",
                                    "hover:!no-underline",
                                    "rounded-3xl",
                                    "py-3 px-8",
                                    "flex items-center justify-center gap-2",
                                    "text-base font-semibold",
                                )}
                            >
                                <DownloadIcon />
                                Download for Android
                            </Link>
                        </div>
                        <p
                            className={clsx(
                                "text-sm",
                                "dark:text-gray-500 text-gray-500",
                            )}
                        >
                            Free & open source, built straight from{" "}
                            <Link to={REPO_URL} target="_blank" rel="noopener noreferrer">
                                GitHub
                            </Link>{" "}
                            - requires Android 7.0 or newer.
                        </p>
                    </div>

                    {/* Screenshots */}
                    <div
                        className={clsx(
                            "flex",
                            "gap-4 landing-sm:gap-6",
                            "overflow-x-auto",
                            "px-4 landing-md:px-10",
                            "snap-x snap-mandatory",
                            "[&::-webkit-scrollbar]:hidden",
                            "[-ms-overflow-style:none] [scrollbar-width:none]",
                        )}
                    >
                        {screenshots.map((src) => (
                            <img
                                key={src}
                                src={src}
                                alt="OpenPanel mobile app screenshot"
                                className={clsx(
                                    "snap-center",
                                    "flex-shrink-0",
                                    "w-[220px] landing-sm:w-[260px]",
                                    "rounded-2xl landing-sm:rounded-3xl",
                                    "border",
                                    "dark:border-gray-700 border-gray-200",
                                    "shadow-lg",
                                )}
                            />
                        ))}
                    </div>

                    {/* Features */}
                    <div className={clsx("w-full", "px-4 landing-md:px-10")}>
                        <div
                            className={clsx(
                                "grid",
                                "grid-cols-1 landing-sm:grid-cols-2",
                                "gap-4 landing-sm:gap-6",
                            )}
                        >
                            {features.map((item) => (
                                <div
                                    key={item.title}
                                    className={clsx(
                                        "p-6 landing-sm:p-8",
                                        "dark:bg-landing-noise",
                                        "dark:bg-gray-800 bg-gray-50",
                                        "rounded-2xl landing-sm:rounded-3xl",
                                    )}
                                >
                                    <div
                                        className={clsx(
                                            "text-lg",
                                            "font-semibold",
                                            "dark:text-gray-0 text-gray-900",
                                        )}
                                    >
                                        {item.title}
                                    </div>
                                    <div
                                        className={clsx(
                                            "mt-2",
                                            "text-base",
                                            "dark:text-gray-400 text-gray-600",
                                        )}
                                    >
                                        {item.description}
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Getting started */}
                    <div className={clsx("w-full", "px-4 landing-md:px-10")}>
                        <h2
                            className={clsx(
                                "text-2xl landing-sm:text-[32px] landing-sm:leading-[40px]",
                                "tracking-tight",
                                "font-semibold",
                                "p-0",
                                "dark:text-gray-0 text-gray-900",
                            )}
                        >
                            Getting started.
                        </h2>
                        <ol
                            className={clsx(
                                "mt-6",
                                "flex flex-col",
                                "gap-4",
                            )}
                        >
                            {steps.map((step, index) => (
                                <li
                                    key={index}
                                    className={clsx(
                                        "flex items-start",
                                        "gap-4",
                                        "p-4 landing-sm:p-6",
                                        "dark:bg-landing-noise",
                                        "dark:bg-gray-800 bg-gray-50",
                                        "rounded-2xl",
                                    )}
                                >
                                    <div
                                        className={clsx(
                                            "flex-shrink-0",
                                            "w-7 h-7",
                                            "flex items-center justify-center",
                                            "rounded-full",
                                            "text-sm font-semibold",
                                            "dark:bg-refine-cyan-alt/10 bg-refine-blue/10",
                                            "dark:text-refine-cyan-alt text-refine-blue",
                                        )}
                                    >
                                        {index + 1}
                                    </div>
                                    <div
                                        className={clsx(
                                            "text-base",
                                            "dark:text-gray-400 text-gray-600",
                                            "pt-0.5",
                                        )}
                                    >
                                        {step}
                                    </div>
                                </li>
                            ))}
                        </ol>
                    </div>

                    <MobileFaq
                        className={clsx(
                            "px-4 landing-sm:px-10 landing-lg:px-0",
                            "w-full landing-lg:max-w-[792px] mx-auto",
                        )}
                    />
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

const sectionWidth = clsx(
    "mx-auto",
    "w-full",
    "max-w-[592px]",
    "landing-sm:max-w-[656px]",
    "landing-md:max-w-[896px]",
    "landing-lg:max-w-[1200px]",
);

export default Mobile;
