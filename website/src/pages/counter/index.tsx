import React, { useState, useEffect } from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import clsx from "clsx";


const Demo: React.FC = () => {
    const [version, setVersion] = useState("8000"); // default/fallback

    useEffect(() => {
        const fetchVersion = async () => {
            try {
                const response = await fetch("https://usage-api.openpanel.org/active_install");
                const data = await response.json();
                if (data && data.active_install) {
                    setVersion(data.active_install);
                }
            } catch (err) {
                console.error("Failed to fetch data:", err);
            }
        };
        fetchVersion();
    }, []);

    return (
        <CommonLayout>
            <Head title="Active Installations Counter | OpenPanel">
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

                    {/* Display latest version number */}
                    <p className="mb-6 text-lg">
                        Active installations: <strong>{version}</strong>
                    </p>
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Demo;
