import React from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import clsx from "clsx";
import USE_CASES from "../../assets/use-cases";

const Demo: React.FC = () => {
    return (
        <CommonLayout>
            <Head title="Live Demo | OpenPanel">
                <html data-page="demo" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="flex-1 flex flex-col pt-8 lg:pt-16 pb-32 max-w-[800px] w-full mx-auto px-2">
                    <h1>Live Demo</h1>
<p>
Explore OpenAdmin & OpenPanel® with no strings attached.
</p>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                    {USE_CASES.map((useCase) => (
                        <div
                            key={useCase.route}
                            className="flex flex-col justify-between"
                        >
                            <div>
                                <Link to={`/docs/${useCase.route}/intro`}>
                                    <img
                                        src={useCase.thumbImgURL}
                                        alt={useCase.title}
                                    />
                                </Link>
                                <div className="font-montserrat font-bold text-[#242436] my-2">
                                    {useCase.title}
                                </div>
                                <div
                                    className="font-montserrat text-[#242436] text-sm"
                                    dangerouslySetInnerHTML={{
                                        __html: useCase.description,
                                    }}
                                />
                            </div>

                            <Link
                                className="shadow-startTiles appearance-none flex items-center justify-center mt-2 mb-4 no-underline font-bold font-montserrat text-sm h-8 w-44 text-[#1890FF] text-center bg-white rounded-[4px] cursor-pointer border border-solid border-[#EDEDEF]"
                                to={`/docs/${useCase.route}/intro`}
                            >
                                Visit Docs
                            </Link>
                        </div>
                    ))}
                </div>




                    
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
