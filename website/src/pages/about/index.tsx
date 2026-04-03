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
                        . ✌️
                    </h1>
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
      "leading-6",
      "md:text-2xl md:leading-8",
      "text-center text-gray-800 dark:text-gray-200",
    )}
    style={{ margin: 0 }}
  >
    We built the hosting panel we always wished existed.
  </h4>
  <p>
    Managing a web hosting server has been broken for too long. Legacy control panels built on outdated architecture, bloated with features nobody uses, charging per-seat fees that punish growth, and treating security as an afterthought. We knew there had to be a better way. So we built one.
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
      "leading-6",
      "md:text-2xl md:leading-8",
      "text-center text-gray-800 dark:text-gray-200",
    )}
    style={{ margin: 0 }}
  >
    Our Story
  </h4>
  <p>
    OpenPanel started from a simple frustration: existing hosting panels weren't keeping up with how modern servers are actually run. Containers had changed everything about infrastructure — but hosting control panels were still stuck in the past.
  </p>
  <p>
    We set out to build a panel that embraced Docker from day one. Not as an add-on. Not as a compatibility layer. As the foundation. That decision shapes everything about how OpenPanel works — the isolation, the flexibility, the security, and the user experience.
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
      "leading-6",
      "md:text-2xl md:leading-8",
      "text-center text-gray-800 dark:text-gray-200",
    )}
    style={{ margin: 0 }}
  >
    What We Believe
  </h4>
  <p>
    Isolation is a right, not a premium feature. Every user on your server deserves their own environment — their own MySQL, their own PHP version, their own resources — without interference from anyone else. That's not enterprise-tier. That's just how hosting should work.
  </p>
  <p>
    Security should be built in, not bolted on. WAF protection, malware scanning, two-factor authentication, IP controls — these shouldn't be things you configure after the fact. They should be there from the moment you install.
  </p>
  <p>
    Control panels should give you control. Over your stack. Over your branding. Over your pricing. Over your data. You're running the server — the panel should work for you, not the other way around.
  </p>
  <p>
    Open source matters. Transparency builds trust. You can audit our code, contribute to it, fork it, or self-host it. No vendor lock-in. No black boxes.
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
      "leading-6",
      "md:text-2xl md:leading-8",
      "text-center text-gray-800 dark:text-gray-200",
    )}
    style={{ margin: 0 }}
  >
    Who OpenPanel Is For
  </h4>
  <p>
    Whether you're a solo developer managing a handful of sites, a growing agency handling client hosting, or a hosting provider serving hundreds of users — OpenPanel scales with you.
  </p>
  <p>
    The free Community Edition gets you started with no strings attached. When your business grows, the Enterprise Edition grows with it — adding advanced features, custom branding, and hands-on support at a flat monthly rate, not a per-seat tax.
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
      "leading-6",
      "md:text-2xl md:leading-8",
      "text-center text-gray-800 dark:text-gray-200",
    )}
    style={{ margin: 0 }}
  >
    Built Different
  </h4>
  <p>
    Unlike traditional panels that assign shared resources and hope for the best, OpenPanel gives each user a VPS-like experience inside a shared server. That means better performance, stronger security, and happier customers — without the cost of dedicated infrastructure for everyone.
  </p>
  <p>
    We support the web servers you actually use. The databases your stack depends on. The PHP versions your clients still need. And a command-line interface powerful enough for the admins who live in the terminal.
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
      "leading-6",
      "md:text-2xl md:leading-8",
      "text-center text-gray-800 dark:text-gray-200",
    )}
    style={{ margin: 0 }}
  >
    Our Commitment
  </h4>
  <p>
    We're continuously improving OpenPanel based on real feedback from real hosting operators. Every update is driven by what the people running servers actually need — not feature bloat, not enterprise upsells, not marketing.
  </p>
  <p>
    We're building the hosting panel for the next decade. We're glad you're here.
  </p>
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
