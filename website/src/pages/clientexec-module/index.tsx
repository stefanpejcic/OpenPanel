import React from "react";
import Head from "@docusaurus/Head";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { RefineWeek } from "@site/src/components/refine-week";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";

const RefineWeekClientExec = () => {
    return (
        <CommonLayout
            description="Automate hosting account provisioning between ClientExec and OpenPanel - install, configure, and provision in about 5 minutes."
            image="/img/og/clientexec-module.png"
        >
            <div className="refine-prose">
                <Head>
                    <title>ClientExec module for OpenPanel | OpenPanel</title>
                    <meta
                        property="og:title"
                        content="ClientExec module for OpenPanel | OpenPanel"
                    />
                    <meta property="og:image:width" content="1200" />
                    <meta property="og:image:height" content="630" />
                    <meta property="og:image:type" content="image/png" />
                    <html data-page="week-of-refine" data-customized="true" />
                </Head>

                <CommonHeader hasSticky={true} />
                <RefineWeek variant="clientexec" />
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default RefineWeekClientExec;
