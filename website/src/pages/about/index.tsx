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
import Translate from "@docusaurus/Translate";
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
            <Translate id="about.hero.title" values={{
              less_time: (
                <span className="text-transparent bg-clip-text bg-gradient-to-r text-gradient-to-r from-[#0FBDBD] to-[#26D97F]">
                  <Translate id="about.hero.less_time">less time</Translate>
                </span>
              ),
              fewer_resources: (
                <span className="text-transparent bg-clip-text bg-gradient-to-r text-gradient-to-r from-[#FF9933] to-[#FF4C4D]">
                  <Translate id="about.hero.fewer_resources">fewer resources</Translate>
                </span>
              )
            }}>
              {"We're helping web hosting provides to host busy websites, in {less_time}, with {fewer_resources}. ✌️"}
            </Translate>
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
              " leading-6",
              "md:text-2xl md:leading-8",
              "text-center text-gray-800 dark:text-gray-200",
            )}
            style={{
              margin: 10,
            }}
          >
            <Translate id="about.who_are_we">Who are we</Translate>
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
            "lg:max-w-[912px] lg:py-16",
            "md:max-w-[624px] md:py-20",
            "sm:max-w-[480px] py-10",
            "max-w-[328px]",
            "w-full mx-auto",
          )}
        >
          <h4
            className={clsx(
              " leading-6",
              "md:text-2xl md:leading-8",
              "text-center text-gray-800 dark:text-gray-200",
            )}
            style={{
              margin: 0,
            }}
          >
            <Translate id="about.idea.title">The Idea 💡</Translate>
          </h4>

          <p>
            <Translate id="about.idea.desc1">OpenPanel was born out of frustration with the premium hosting panels we were using at UNLIMITED.RS for our shared hosting and managed VPS clients. The costs were skyrocketing, yet these solutions lacked the flexibility needed to truly meet our clients' unique needs. During one particularly challenging project, a colleague, Čeda, famously remarked, "Even we could do better." This statement set Stefan Pejčić, our co-founder and CEO, on a path to create what would eventually become OpenPanel.</Translate>
          </p>
          <p>
            <Translate id="about.idea.desc2" values={{
              chance: (
                <b>
                  <Translate id="about.idea.chance">our chance to build something better</Translate>
                </b>
              )
            }}>
              {"Having worked with various hosting panels over the years, we knew their bugs, limitations, and how they often failed to deliver. We saw this as {chance} —something that could truly scale and adapt."}
            </Translate>
          </p>

        </div>
        <div
          className={clsx(
            "lg:max-w-[912px] lg:py-16",
            "md:max-w-[624px] md:py-20",
            "sm:max-w-[480px] py-10",
            "max-w-[328px]",
            "w-full mx-auto",
          )}
        >
          <h4
            className={clsx(
              " leading-6",
              "md:text-2xl md:leading-8",
              "text-center text-gray-800 dark:text-gray-200",
            )}
            style={{
              margin: 0,
            }}
          >
            <Translate id="about.early_stages.title">Early stages 🚧</Translate>
          </h4>
          <p>
            <Translate id="about.early_stages.desc" values={{
              started_in_2023: (
                <b>
                  <Translate id="about.early_stages.started">started in 2023</Translate>
                </b>
              )
            }}>
              {"OpenPanel {started_in_2023}, as a basic bootstrap interface with a database schema designed to handle anywhere from 10 to 500 users on a single server without issues. It was initially a LAMP stack with a fancy, yet buggy, GUI. But as we grew, so did our ideas."}
            </Translate>
          </p>


        </div>
        <div
          className={clsx(
            "lg:max-w-[912px] lg:py-16",
            "md:max-w-[624px] md:py-20",
            "sm:max-w-[480px] py-10",
            "max-w-[328px]",
            "w-full mx-auto",
          )}
        >


          <h4
            className={clsx(
              " leading-6",
              "md:text-2xl md:leading-8",
              "text-center text-gray-800 dark:text-gray-200",
            )}
            style={{
              margin: 0,
            }}
          >
            <Translate id="about.financing.title">Our Financing 💸</Translate>
          </h4>
          <p>
            <Translate id="about.financing.desc" values={{
              independent: (
                <b>
                  <Translate id="about.financing.independent">chose to remain independent</Translate>
                </b>
              ),
              panel: (
                <b>
                  <Translate id="about.financing.panel">a modular, stable and fairly priced control panel</Translate>
                </b>
              )
            }}>
              {"Over the years, we received several offers from competitors and venture capitalists, but with those offers came conditions that would have compromised our vision. We {independent} and self-funded. Committed to doing something different: {panel}."}
            </Translate>
          </p>

        </div>
        <div
          className={clsx(
            "lg:max-w-[912px] lg:py-16",
            "md:max-w-[624px] md:py-20",
            "sm:max-w-[480px] py-10",
            "max-w-[328px]",
            "w-full mx-auto",
          )}
        >

          <h4
            className={clsx(
              " leading-6",
              "md:text-2xl md:leading-8",
              "text-center text-gray-800 dark:text-gray-200",
            )}
            style={{
              margin: 0,
            }}
          >
            <Translate id="about.why_openpanel.title">Why OpenPanel ? 🤔</Translate>
          </h4>
          <p>
            <Translate id="about.why_openpanel.desc" values={{
              not_associated: (
                <b>
                  <Translate id="about.why_openpanel.not_associated">in no way associated with the older openpanel.com project</Translate>
                </b>
              )
            }}>
              {"The name \"OpenPanel\" itself was initially just a placeholder, but it stuck. The project was first published under .co domain, as the .com and .org were already taken by a similar project almost two decades ago. The OpenPanel you know and use today is {not_associated}, but we extend our gratitude to them for their unselfishness in allowing us to take over the openpanel.com domain in August 2024."}
            </Translate>
          </p>


        </div>
        <div
          className={clsx(
            "lg:max-w-[912px] lg:py-16",
            "md:max-w-[624px] md:py-20",
            "sm:max-w-[480px] py-10",
            "max-w-[328px]",
            "w-full mx-auto",
          )}
        >


          <h4
            className={clsx(
              " leading-6",
              "md:text-2xl md:leading-8",
              "text-center text-gray-800 dark:text-gray-200",
            )}
            style={{
              margin: 0,
            }}
          >
            <Translate id="about.why_not_opensource.title">Why not open-sourced ? 🤷</Translate>
          </h4>
          <p>
            <Translate id="about.why_not_opensource.desc" values={{
              guarantee_quality: (
                <b>
                  <Translate id="about.why_not_opensource.guarantee_quality">guarantee the quality</Translate>
                </b>
              ),
              secure_product: (
                <b>
                  <Translate id="about.why_not_opensource.secure_product">a more secure product</Translate>
                </b>
              )
            }}>
              {"While we considered going fully open-source, we decided to keep certain parts proprietary to ensure we can {guarantee_quality} of our features, expedite the development of new ones, and provide quicker fixes for bugs. This approach also allows us to maintain {secure_product}, giving our users peace of mind."}
            </Translate>
          </p>


        </div>
        <div
          className={clsx(
            "lg:max-w-[912px] lg:py-16",
            "md:max-w-[624px] md:py-20",
            "sm:max-w-[480px] py-10",
            "max-w-[328px]",
            "w-full mx-auto",
          )}
        >
          <h4
            className={clsx(
              " leading-6",
              "md:text-2xl md:leading-8",
              "text-center text-gray-800 dark:text-gray-200",
            )}
            style={{
              margin: 0,
            }}
          >
            <Translate id="about.next.title">What is next ? 🔮</Translate>
          </h4>
          <p>
            <Translate id="about.next.desc1">Stefan Pejcic folyamatos vezetése alatt az OpenPanel elkötelezett marad amellett, hogy a technológiai trendek előtt járjon, és kezelje az azokkal járó kockázatokat.</Translate>
          </p>
          <p>
            <Translate id="about.next.desc2" values={{
              advanced_panel: (
                <b>
                  <Translate id="about.next.advanced_panel">the most advanced, secure, and user-friendly hosting panel</Translate>
                </b>
              )
            }}>
              {"We are dedicated to constantly improving our platform to offer {advanced_panel} on the market. Our journey is far from over, and with your continued support, we’re excited to push the boundaries of what’s possible in server management."}
            </Translate>
          </p>
          <p>
            <Translate id="about.next.desc3" values={{
              seamless_experience: (
                <b>
                  <Translate id="about.next.seamless_experience">make your hosting experience as seamless as possible</Translate>
                </b>
              )
            }}>
              {"Stay tuned for new features, enhanced performance, and more ways to {seamless_experience}. The best is yet to come."}
            </Translate>
          </p>



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
              " leading-6",
              "md:text-2xl md:leading-8",
              "text-center text-gray-800 dark:text-gray-200",
              "mb-8 lg:mb-16",
            )}
          >
            <Translate id="about.contact_us">Contact Us</Translate>
          </h4>

          <div className="flex flex-col md:gap-8 lg:flex-row lg:gap-10 xl:gap-24">
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
                  IJsbaanpad 2, 1076 CV Amsterdam (NL)
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
                  href="mailto:info@openpanel.com"
                  className="text-gray-700 dark:text-gray-300 hover:no-underline no-underline"
                >
                  info@openpanel.com
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
