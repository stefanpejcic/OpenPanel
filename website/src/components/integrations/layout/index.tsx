import React from "react";
import Head from "@docusaurus/Head";
import clsx from "clsx";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";

const IntegrationsLayout = ({ children }: React.PropsWithChildren<{}>) => {
    return (
        <CommonLayout
            description="See every feature built into OpenPanel: multiple PHP versions, per-user isolation, Nginx/OpenLiteSpeed, Podman, backups, and built-in security."
            image="/img/og/features.png"
        >
            <Head>
                <title>Features | OpenPanel</title>
                <meta property="og:title" content="Features | OpenPanel" />
                <meta property="og:image:width" content="1200" />
                <meta property="og:image:height" content="630" />
                <meta property="og:image:type" content="image/png" />
                <html data-page="integrations" data-customized="true" />
            </Head>
            <div className={clsx("refine-prose, pb-16")}>
                <CommonHeader hasSticky />
                <div
                    className={clsx(
                        "max-w-[944px]",
                        "mx-auto",
                        "pt-16 px-4 sm:px-6",
                    )}
                >
                    {children}
                </div>
            </div>
            <BlogFooter />
        </CommonLayout>
    );
};

export default IntegrationsLayout;
