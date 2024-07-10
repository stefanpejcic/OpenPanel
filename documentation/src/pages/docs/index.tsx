import Head from "@docusaurus/Head";
import { FooterRedditIcon as RedditIcon } from "@site/src/refine-theme/icons/footer-reddit";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { GithubIcon } from "@site/src/refine-theme/icons/github";
import { JoinUsIcon } from "@site/src/refine-theme/icons/join-us";
import { MailIcon } from "@site/src/refine-theme/icons/mail";
import { MarkerIcon } from "@site/src/refine-theme/icons/marker";
import { OpenSourceIcon } from "@site/src/refine-theme/icons/open-source";
import { DiscordIcon, TwitterIcon } from "@site/src/refine-theme/icons/popover";
import clsx from "clsx";
import React from "react";
import { useColorMode } from "@docusaurus/theme-common";
import { LandingTryItSection } from "../refine-theme/demo-section";

const About: React.FC = () => {
    const { colorMode } = useColorMode();
    return (
        <>
            <Head title="Documentation | OpenPanel">
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

<div className="flex-1 flex flex-col pt-4 lg:pt-6 pb-32 max-w-[1040px] w-full mx-auto px-2">
<p>
Whether you are a new user or an experienced administrator, our comprehensive documentation is designed to get you up and running quickly, while also providing deep insights into advanced features and customization options. Explore our resources below to make the most out of our products.
</p>      

<LandingTryItSection />

    
<div className="flex">
                        {/* First Column */}
                        <div className="w-1/2 pr-4">
                            <a href="/docs/panel/intro/" rel="noopener noreferrer">
                                <h2>OpenPanel Docs</h2>
                                <img src="/img/panel/v1/dashboard/dashboard.png" alt="OpenPanel Documentation" />
                                
                            </a>

<p>
OpenPanel offers a robust interface for end-users aiming to simplify the complexities of web and server management. From adding domains to managing your websites, our documentation covers everything you need to seamlessly navigate through the interface.
</p>
                        </div>

                        {/* Second Column */}
                        <div className="w-1/2 pl-4">
                            <a href="/docs/admin/intro/" rel="noopener noreferrer">
                                <h2>OpenAdmin Docs</h2>
                                <img src="/img/admin/openadmin_dashboard.png" alt="OpenAdmin Documentation" />
                                
                            </a>
                            <p>
OpenAdmin is tailored for administrators seeking fine-grained control over server configurations and management. Our documentation provides in-depth knowledge to help you customize and secure your server environment.
                            </p>
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
