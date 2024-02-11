import Head from "@docusaurus/Head";
import { FooterRedditIcon as RedditIcon } from "@site/src/refine-theme/icons/footer-reddit";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { Istanbul500Icon } from "@site/src/refine-theme/icons/500";
import { GithubIcon } from "@site/src/refine-theme/icons/github";
import { JoinUsIcon } from "@site/src/refine-theme/icons/join-us";
import { MailIcon } from "@site/src/refine-theme/icons/mail";
import { MarkerIcon } from "@site/src/refine-theme/icons/marker";
import { OpenSourceIcon } from "@site/src/refine-theme/icons/open-source";
import { DiscordIcon, TwitterIcon } from "@site/src/refine-theme/icons/popover";
import Link from "@docusaurus/Link";
import clsx from "clsx";
import React from "react";
import { backedBy } from "../../assets/backed-by";
import { team } from "../../assets/team";
import { useColorMode } from "@docusaurus/theme-common";
import { YCombinatorCircleIcon } from "@site/src/refine-theme/icons/ycombinator-circle";
import { CommonThemedImage } from "@site/src/refine-theme/common-themed-image";

const About: React.FC = () => {
    const { colorMode } = useColorMode();
    return (
        <>
            <Head title="About | Refine">
                <html data-page="about" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />
                <div
                    className={clsx(
                        "not-prose",
                        "xl:max-w-[944px] xl:py-16",
                        "lg:max-w-[912px] lg:py-10",
                        "md:max-w-[624px] md:text-4xl  md:pb-6 pt-6",
                        "sm:max-w-[480px] text-xl",
                        "max-w-[328px]",
                        "w-full mx-auto",
                    )}
                >
                    <h1
                        className={clsx(
                            "font-semibold",
                            "!mb-0",
                            "text-gray-900 dark:text-gray-0",
                            "text-xl md:text-[40px] md:leading-[56px]",
                        )}
                    >
                        Documentation for{" "}
                        <span className="text-transparent bg-clip-text bg-gradient-to-r text-gradient-to-r from-[#0FBDBD] to-[#26D97F]">
                            OpenPanel
                        </span>
                        , and{" "}
                        <span className="text-transparent bg-clip-text bg-gradient-to-r text-gradient-to-r from-[#FF9933] to-[#FF4C4D]">
                            OpenAdmin
                        </span>
                        .
                    </h1>
                </div>

<div className="flex-1 flex flex-col pt-8 lg:pt-16 pb-32 max-w-[800px] w-full mx-auto px-2">
                
<div className="flex">
                        {/* First Column */}
                        <div className="w-1/2 pr-4">
                            <a href="https://demo.openpanel.co/openapanel/" target="_blank" rel="noopener noreferrer">
                                <h2>OpenPanel Docs</h2>
                                <img src="/img/panel/v1/dashboard/dashboard.png" alt="Demo OpenPanel" style={{ width: 'auto', height: '350px' }} />
                                
                            </a>

<p>
    
</p>
                          
                        </div>

                        {/* Second Column */}
                        <div className="w-1/2 pl-4">
                            <a href="https://demo.openpanel.co/openadmin/" target="_blank" rel="noopener noreferrer">
                                <h2>OpenAdmin Docs</h2>
                                <img src="/img/admin/openadmin_dashboard.png" alt="Demo OpenAdmin" style={{ width: 'auto', height: '350px' }} />
                                
                            </a>
<p>

</p>
                        </div>
                    </div>
</div>



                
                <div
                    className={clsx(
                        "w-[328px] sm:w-[480px] md:w-[624px] lg:w-[912px] xl:w-[1120px]",
                        "mx-auto mt-10 w-full md:mt-20 lg:mt-32",
                        "flex flex-col xl:flex-row",
                        "gap-4 xl:gap-16",
                    )}
                >
                    <div
                        className={clsx(
                            "flex justify-center items-center",
                            "w-[48px] h-[48px]",
                            "md:w-[64px] md:h-[64px]",
                            "rounded-full ",
                            "bg-refine-red bg-opacity-10",
                            "shrink-0",
                            "xl:hidden",
                        )}
                    >
                        <OpenSourceIcon className="text-refine-red" />
                    </div>
                    <div className="flex flex-col gap-2 lg:flex-row lg:gap-8 xl:gap-16">
                        <div
                            className={clsx(
                                "flex flex-col gap-6 lg:flex-row",
                                "w-[328px] sm:w-[480px] md:w-[624px] lg:w-[912px] xl:w-[576px]",
                            )}
                        >
                            <div
                                className={clsx(
                                    "justify-center items-center",
                                    "w-[48px] h-[48px]",
                                    "md:w-[64px] md:h-[64px]",
                                    "rounded-full ",
                                    "bg-refine-red bg-opacity-10",
                                    "shrink-0",
                                    "hidden xl:flex",
                                )}
                            >
                                <OpenSourceIcon className="text-refine-red" />
                            </div>

                            <div>
                                <p className="mb-3 text-sm font-semibold text-gray-900 dark:text-gray-300 md:text-lg lg:text-2xl">
                                    User-friendly guides
                                </p>

                                <p className="text-xs sm:text-base text-gray-900 dark:text-gray-300">
                                    Soon..
                                </p>
                            </div>
                        </div>
                        <div className="grid w-full shrink-0 grid-cols-2 gap-4 lg:w-[400px]">
                            <a
                                target="_blank"
                                href="/docs/panel/intro/"
                                className={clsx(
                                    "flex h-max flex-row justify-start gap-3",
                                    "dark:bg-gray-900",
                                    "border border-gray-200 dark:border-gray-700",
                                    "rounded-xl p-4",
                                    "no-underline hover:no-underline",
                                )}
                                rel="noreferrer"
                            >
                                <div>
                                    <GithubIcon
                                        className="text-2xl text-gray-900 dark:text-gray-0"
                                        width="24px"
                                        height="24px"
                                    />
                                </div>
                                <div className="text-xs md:text-base">
                                    <div className="mb-0 text-gray-500 dark:text-gray-400">
                                        Docs for end-users
                                    </div>
                                    <div className="font-semibold text-gray-900 dark:text-gray-0 no-underline hover:no-underline">
                                        OpenPanel
                                    </div>
                                </div>
                            </a>
                            <a
                                target="_blank"
                                href="/docs/admin/intro/"
                                rel="noreferrer"
                                className={clsx(
                                    "flex  h-max flex-row justify-start gap-3",
                                    "dark:bg-gray-900",
                                    "border border-gray-200 dark:border-gray-700",
                                    "p-4 rounded-xl",
                                    "no-underline hover:no-underline",
                                )}
                            >
                                <div>
                                    <DiscordIcon
                                        className="text-2xl"
                                        width="24px"
                                        height="24px"
                                    />
                                </div>
                                <div className="text-xs md:text-base">
                                    <div className="mb-0 text-gray-500 dark:text-gray-400">
                                        Docs for hosters
                                    </div>
                                    <div className="font-semibold text-gray-900 dark:text-gray-0 no-underline hover:no-underline">
                                        OpenAdmin
                                    </div>
                                </div>
                            </a>
                            <a
                                target="_blank"
                                href="/docs/category/openpanel-cli/"
                                rel="noreferrer"
                                className={clsx(
                                    "flex  h-max flex-row justify-start gap-3",
                                    "dark:bg-gray-900",
                                    "border border-gray-200 dark:border-gray-700",
                                    "p-4 rounded-xl",
                                    "no-underline hover:no-underline",
                                )}
                            >
                                <div>
                                    <RedditIcon
                                        className="text-2xl"
                                        width="24px"
                                        height="24px"
                                        color="#FF4500"
                                    />
                                </div>
                                <div className="text-xs md:text-base">
                                    <div className="mb-0 text-gray-500 dark:text-gray-400">
                                        Docs for Admins
                                    </div>
                                    <div className="font-semibold text-gray-900 dark:text-gray-0 no-underline hover:no-underline">
                                       OpenCLI
                                    </div>
                                </div>
                            </a>
                            <a
                                target="_blank"
                                href="/docs/category/advanced/"
                                rel="noreferrer"
                                className={clsx(
                                    "flex  h-max flex-row justify-start gap-3",
                                    "dark:bg-gray-900",
                                    "border border-gray-200 dark:border-gray-700",
                                    "p-4 rounded-xl",
                                    "no-underline hover:no-underline",
                                )}
                            >
                                <div>
                                    <TwitterIcon
                                        className="text-2xl"
                                        width="24px"
                                        height="24px"
                                    />
                                </div>
                                <div className="text-xs md:text-base">
                                    <div className="mb-0 text-gray-500 dark:text-gray-400">
                                        Docs for Developers
                                    </div>
                                    <div className="font-semibold text-gray-900 dark:text-gray-0 no-underline hover:no-underline">
                                        Advanced
                                    </div>
                                </div>
                            </a>
                        </div>
                    </div>
                </div>


                <BlogFooter />
            </div>
        </>
    );
};

export default function AboutPage() {
    return (
        <CommonLayout>
            <About />
        </CommonLayout>
    );
}
