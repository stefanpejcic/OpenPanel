import React from "react";
import Head from "@docusaurus/Head";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { RefineWeek } from "@site/src/components/refine-week";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";

const RefineWeekWisecp = () => {
    return (
        <CommonLayout
            description="Automate hosting account provisioning between WISECP and OpenPanel - install, configure, and provision in about 5 minutes."
            image="/img/og/wisecp-module.png"
        >
            <div className="refine-prose">
                <Head>
                    <title>WISECP module for OpenPanel | OpenPanel</title>
                    <meta
                        property="og:title"
                        content="WISECP module for OpenPanel | OpenPanel"
                    />
                    <meta property="og:image:width" content="1200" />
                    <meta property="og:image:height" content="630" />
                    <meta property="og:image:type" content="image/png" />
                    <html data-page="week-of-refine" data-customized="true" />
                </Head>

                <CommonHeader hasSticky={true} />
                <RefineWeek variant="wisecp" />
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default RefineWeekWisecp;
