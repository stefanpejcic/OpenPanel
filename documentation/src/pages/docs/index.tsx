import React from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import clsx from "clsx";
import { LandingTryItSection } from "@site/src/refine-theme/docs-section";

const Demo: React.FC = () => {
    return (
        <CommonLayout>
            <Head title="Documentation | OpenPanel">
                <html data-page="docs" data-customized="true" />
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
                            "mb-12",
                            "text-gray-900 dark:text-gray-0",
                            "text-xl md:text-[40px] md:leading-[56px]",
                        )}
                    >
                        Documentation for{" "}
                        <span className="text-transparent bg-clip-text bg-gradient-to-r text-gradient-to-r from-[#0FBDBD] to-[#26D97F]">
                            OpenPanel
                        </span>
                         and{" "}
                        <span className="text-transparent bg-clip-text bg-gradient-to-r text-gradient-to-r from-[#FF9933] to-[#FF4C4D]">
                            OpenAdmin
                        </span>
                        .
                    </h1>
                    
                    <LandingTryItSection />
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Demo;
