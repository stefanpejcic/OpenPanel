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
            <Head title="About | OpenPanel Project">
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
                        We&apos;re helping web hosting provides to host busy websites, in{" "}
                        <span className="text-transparent bg-clip-text bg-gradient-to-r text-gradient-to-r from-[#0FBDBD] to-[#26D97F]">
                            less time
                        </span>
                        , with{" "}
                        <span className="text-transparent bg-clip-text bg-gradient-to-r text-gradient-to-r from-[#FF9933] to-[#FF4C4D]">
                            fewer resources
                        </span>
                        .
                    </h1>
                </div>









        <div
          className={clsx(
            "lg:max-w-[912px] lg:py-32",
            "md:max-w-[624px] md:py-20",
            "sm:max-w-[480px] py-10",
            "max-w-[328px]",
            "w-full mx-auto",
          )}
        >
          <h4
            className={clsx(
              "text-sm leading-6",
              "md:text-2xl md:leading-8",
              "text-center text-gray-800 dark:text-gray-200",
            )}
            style={{
              margin: 0,
            }}
          >
            Our Team
          </h4>

          <div
            className={clsx(
              "grid",
              "lg:grid-cols-[repeat(4,192px)] lg:gap-12",
              "md:grid-cols-[repeat(3,176px)]",
              "sm:grid-cols-[repeat(3,144px)]",
              "grid-cols-[repeat(2,144px)] gap-6",
              "align-top",
              "mt-6 md:mt-12 lg:mt-16",
            )}
          >
            {team.map(({ name, avatar, role1, role2 }) => (
              <div
                key={name}
                className="flex justify-start flex-col text-center not-prose"
              >
                <img
                  srcSet={`${avatar} 1500w`}
                  src={avatar}
                  alt={name}
                  className="w-full m-0 mb-6"
                />
                <span
                  className={clsx(
                    "text-xs leading-4",
                    "lg:text-base lg:leading-6",
                    "text-gray-900 dark:text-gray-0 font-semibold",
                  )}
                >
                  {name}
                </span>
                <span
                  className={clsx(
                    "text-xs leading-4",
                    "lg:text-base lg:leading-6",
                    "text-gray-500 dark:text-gray-400",
                  )}
                >
                  {role1}
                </span>
                {role2 && (
                  <span
                    className={clsx(
                      "text-xs leading-4",
                      "lg:text-base lg:leading-6",
                      "text-gray-500 dark:text-gray-400",
                    )}
                  >
                    {role2}
                  </span>
                )}
              </div>
            ))}
            <div
              className={clsx(
                "flex",
                "flex-col",
                "justify-between lg:justify-start",
                "text-center",
              )}
            >
              <div className="w-full not-prose m-0">
                <JoinUsIcon
                  className={clsx(
                    "m-0 w-full lg:mb-6",
                    "lg:h-[240px]",
                    "md:h-[220px]",
                    "h-[180px]",
                  )}
                  isDark={colorMode === "dark"}
                />
              </div>
              <div className="flex flex-col items-center justify-center">
                <a
                  target="_blank"
                  href="https://www.linkedin.com/company/openpanel"
                  className={clsx(
                    "block",
                    "text-xs leading-4",
                    "lg:text-base lg:leading-6",
                    "no-underline hover:no-underline text-refine-link-light dark:text-refine-link-dark font-semibold mb-0",
                  )}
                  rel="noreferrer"
                >
                  Join our team!
                </a>
                <a
                  target="_blank"
                  href="https://www.linkedin.com/company/openpanel"
                  className={clsx(
                    "block",
                    "text-xs leading-4",
                    "lg:text-base lg:leading-6",
                    "no-underline hover:no-underline text-refine-link-light dark:text-refine-link-dark m-0",
                  )}
                  rel="noreferrer"
                >
                  See open positions
                </a>
              </div>
            </div>
          </div>
        </div>
                










                




                

                <div
                    className={clsx(
                        "xl:max-w-[1016px] lg:py-16",
                        "lg:max-w-[912px] lg:py-16",
                        "md:max-w-[624px] md:py-10",
                        "sm:max-w-[480px] py-8",
                        "max-w-[328px]",
                        "w-full mx-auto",
                    )}
                >
                    <h4
                        className={clsx(
                            "text-sm leading-6",
                            "md:text-2xl md:leading-8",
                            "text-center text-gray-800 dark:text-gray-200",
                            "mb-8 lg:mb-16",
                        )}
                    >
                        We are Here
                    </h4>

                    <div className="flex flex-col md:gap-8 lg:flex-row lg:gap-10 xl:gap-24">
                        <div className="w-full shrink-0 lg:order-last lg:h-[416px] lg:w-[624px]">
                            <Link to="https://goo.gl/maps/D4NZ5gn6VsWaRtXT6">
                                <img
                                    className="m-0 p-0"
                                    src="https://refine.ams3.cdn.digitaloceanspaces.com/website/static/about/images/map.png"
                                    srcSet="https://refine.ams3.cdn.digitaloceanspaces.com/website/static/about/images/map2x.png 1500w"
                                />
                            </Link>
                        </div>
                        <div className="flex justify-start flex-col items-start gap-8 lg:pt-12 pt-6">
                            <div className="flex w-max items-center justify-center gap-6">
                                <div
                                    className={clsx(
                                        "flex justify-center items-center",
                                        "w-[48px] h-[48px]",
                                        "rounded-full ",
                                        "bg-refine-orange bg-opacity-10",
                                        "shrink-0",
                                    )}
                                >
                                    <MarkerIcon className="text-refine-orange" />
                                </div>
                                <span className="text-gray-700 dark:text-gray-300">
                                    SOON
                                    
                                </span>
                            </div>
                            <div className="flex justify-center items-center gap-6">
                                <div
                                    className={clsx(
                                        "flex justify-center items-center",
                                        "w-[48px] h-[48px]",
                                        "rounded-full ",
                                        "bg-refine-pink bg-opacity-10",
                                        "shrink-0",
                                    )}
                                >
                                    <MailIcon className="text-refine-pink" />
                                </div>
                                <a
                                    href="mailto:info@openpanel.co"
                                    className="text-gray-700 dark:text-gray-300 hover:no-underline no-underline"
                                >
                                    info@openpanel.co
                                </a>
                            </div>
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
