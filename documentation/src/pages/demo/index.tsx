import React from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import clsx from "clsx";

const Demo: React.FC = () => {
    return (
        <CommonLayout>
            <Head title="Live Demo | OpenPanel">
                <html data-page="demo" data-customized="true" />
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
                        Live Demo for{" "}
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
Explore OpenAdmin & OpenPanel® with no strings attached.
</p>
                    <div className="flex">
                        {/* First Column */}
                        <div className="w-1/2 pr-4">
                            <a href="https://demo.openpanel.co/openpanel/" target="_blank" rel="noopener noreferrer">
                                <h2>OpenPanel Demo</h2>
                                <img src="/img/panel/v1/dashboard/dashboard.png" alt="Demo OpenPanel" />
                            </a>
                            <p>
                            If you are a website owner, then we recommend interacting with the OpenPanel interface. This is where you can maintain your website.
                            </p>
                          
                        </div>

                        {/* Second Column */}
                        <div className="w-1/2 pl-4">
                            <a href="https://demo.openpanel.co/openadmin/" target="_blank" rel="noopener noreferrer">
                                <h2>OpenAdmin Demo</h2>
                                <img src="/img/admin/openadmin_dashboard.png" alt="Demo OpenAdmin" />
                            </a>
                            <p>
                            If you are a web host, then we recommend interacting with the OpenAdmin interface. This is where you can run and maintain your server.
                            </p>
                        </div>
                    </div>
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Demo;
