import React from "react";
import Head from "@docusaurus/Head";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { RefineWeek } from "@site/src/components/refine-week";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";

const RefineWeekBlesta = () => {
    return (
        <CommonLayout>
            <div className="refine-prose">
                <Head title="Blesta module for OpenPanel | OpenPanel">
                    <html data-page="week-of-refine" data-customized="true" />
                </Head>

                <CommonHeader hasSticky={true} />
                <RefineWeek variant="blesta" />
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default RefineWeekBlesta;
