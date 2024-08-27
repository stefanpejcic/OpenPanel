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
            The Idea
          </h4>

            <p className="text-xs sm:text-base">
OpenPanel was born out of frustration with the premium hosting panels we were using at <a href="https://unlimited.rs?utm=openpanel" target="_blank">UNLIMITED.RS</a> for our shared hosting and managed VPS clients. The costs were skyrocketing, yet these solutions lacked the flexibility needed to truly meet our clients' unique needs. During one particularly challenging project, a colleague, Čeda, famously remarked, "Even we could do better." This statement set Stefan Pejčić, our co-founder and CEO, on a path to create what would eventually become OpenPanel.
            </p>
            <p className="text-xs sm:text-base">
Having worked with various hosting panels over the years, we knew their bugs, limitations, and how they often failed to deliver. We saw this as <b>our chance to build something better</b>—something that could truly scale and adapt.
            </p>

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
            Early stages 
          </h4>
<p className="text-xs sm:text-base">
OpenPanel <b>started in 2023</b>, as a basic bootstrap interface with a database schema designed to handle anywhere from 10 to 500 users on a single server without issues. It was initially a LAMP stack with a fancy, yet buggy, GUI. But as we grew, so did our ideas.
</p>
            <p className="text-xs sm:text-base">
One of our more ambitious concepts was to <b>use Docker containers</b> to isolate users' websites. However, this approach proved unscalable for shared servers with hundreds of users. Managing the overhead of memory and disk space became a significant challenge, especially when a single WordPress site needed at least three containers (database, PHP, and web server).
</p>

<p className="text-xs sm:text-base">
We eventually pivoted to a <b>one Docker per user</b> approach, sacrificing some flexibility but gaining a more stable environment. This method allowed hosting providers to manage resources more effectively, although it requires us to build and </b>maintain our custom images</b> rather than using official Docker ones.
    </p>        



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
            Our Financing
          </h4>
<p className="text-xs sm:text-base">
Over the years, we received several offers from competitors and venture capitalists, but with those offers came conditions that would have compromised our vision. We <b>chose to remain independent</b> and self-funded. Committed to doing something different: <b>a modular, stable and fairly priced control panel</b>.
</p>

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
            Why OpenPanel ?
          </h4>
<p className="text-xs sm:text-base">
The name "OpenPanel" itself was initially just a placeholder, but it stuck. The project was first published under .co domain, as the .com and .org were already taken by a similar project almost two decades ago. The OpenPanel you know and use today is <b>in no way associated with the older openpanel.com project</b>, but we extend our gratitude to them for their unselfishness in allowing us to take over the openpanel.com domain in August 2024.
</p>


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
            Why not open-sourced ?
          </h4>
<p className="text-xs sm:text-base">
While we considered going fully open-source, we decided to keep certain parts proprietary to ensure we can <b>guarantee the quality</b> of our features, expedite the development of new ones, and provide quicker fixes for bugs. This approach also allows us to maintain <b>a more secure product</b>, giving our users peace of mind.
</p>


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
            Pricing
          </h4>
<p className="text-xs sm:text-base">
OpenPanel Is available in two options: <b>Enterprise Edition</b> that can rival any premium panel in features and flexibility, and a <b>Community Edition</b> funded by the Enterprise version, offering all the essential features needed by small agencies, freelancers, and hobby users.
</p>
            <p className="text-xs sm:text-base">
And one promise we stand by—<b>we will never charge per user or per website</b>.
</p>



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
            What is next ?
          </h4>
<p className="text-xs sm:text-base">
Under the continued leadership of Stefan Pejčić, OpenPanel remains committed to staying ahead of technological trends and addressing the risks that come with them.
</p>
            <p className="text-xs sm:text-base">
We are dedicated to constantly improving our platform to offer <b>the most advanced, secure, and user-friendly hosting panel</b> on the market.
Our journey is far from over, and with your continued support, we’re excited to push the boundaries of what’s possible in server management.
</p>
            <p className="text-xs sm:text-base">
Stay tuned for new features, enhanced performance, and more ways to <b>make your hosting experience as seamless as possible</b>. The best is yet to come.
</p>

            
            
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
