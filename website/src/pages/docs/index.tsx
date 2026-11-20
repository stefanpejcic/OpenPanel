import React from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import clsx from "clsx";
import { LandingTryItSection } from "@site/src/refine-theme/docs-section";

const Demo: React.FC = () => {
    return (
        <CommonLayout
            description="Browse the OpenPanel and OpenAdmin documentation: installation, API reference, terminal commands, and how-to guides for running OpenPanel hosting infrastructure."
            image="/img/og/docs.png"
        >
            <Head>
                <title>Documentation | OpenPanel</title>
                <meta property="og:title" content="Documentation | OpenPanel" />
                <meta property="og:image:width" content="1200" />
                <meta property="og:image:height" content="630" />
                <meta property="og:image:type" content="image/png" />
                <html data-page="docs" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div
                    className={clsx(
                        "not-prose",
                        "xl:py-16 lg:py-10",
                        "md:text-4xl pb-12 md:pb-6 pt-6 text-xl",
                        "w-full max-w-[592px] landing-sm:max-w-[656px] landing-md:max-w-[896px] landing-lg:max-w-[1200px]",
                        "px-4 landing-sm:px-0",
                        "mx-auto",
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
                        , and{" "}
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
